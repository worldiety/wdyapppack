package bundle

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/foundation/suggest"
	"github.com/worldiety/wdyapppack/pkg/icns"
	"github.com/worldiety/wdyapppack/pkg/plist"
)

// BuildRequest is what a build is asked to do.
type BuildRequest struct {
	// Config is the description of the bundle, with options already applied.
	Config Config
	// ModTime, when not zero, is given to every file written, for
	// reproducible output.
	ModTime time.Time
	// Strict turns notes into errors.
	Strict bool
	Force  bool
	DryRun bool
}

// BuildResult describes the bundle.
type BuildResult struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Identifier   string   `json:"identifier"`
	Version      string   `json:"version"`
	BuildNumber  string   `json:"buildNumber"`
	Executable   string   `json:"executable"`
	Archs        []string `json:"archs"`
	MinimumMacOS string   `json:"minimumMacOS"`
	IconSizes    []int    `json:"iconSizes,omitempty"`
	// Files lists the bundle's contents, relative to its parent directory.
	Files     []string `json:"files"`
	InfoPlist string   `json:"infoPlist"`
	DryRun    bool     `json:"dryRun"`
	Report
}

// DefaultMinimumMacOS is assumed when neither the user nor the program says.
// It is the first macOS for Apple silicon.
const DefaultMinimumMacOS = "11.0"

func build(files Files, req BuildRequest) (BuildResult, error) {
	cfg := req.Config
	var c problem.Collector

	meta := checkMetadata(cfg, &c)

	// Programs.
	var progs []program
	if len(cfg.Binaries) == 0 {
		c.Error(problem.Usage, msg("WAP-BINARY-MISSING", "No program was given."))
	}
	for _, b := range cfg.Binaries {
		if strings.Contains(b.Path, Placeholder) {
			c.Error(problem.Usage, msg("WAP-CONFIG-PLACEHOLDER", fmt.Sprintf("The program path %q is the placeholder written by init.", b.Path)))
			continue
		}
		if p, ok := loadProgram(files, b, "program", &c); ok {
			programNotes(p, &c)
			progs = append(progs, p)
		}
	}
	var exe []byte
	var archs []string
	var progMinOS string
	if len(progs) == len(cfg.Binaries) && len(progs) > 0 {
		var ok bool
		exe, archs, progMinOS, ok = assemble(progs, &c)
		if !ok {
			exe = nil
		}
	}

	// Minimum macOS.
	meta.minOS = cmp.Or(progMinOS, DefaultMinimumMacOS)
	if cfg.MinimumMacOS != "" {
		switch {
		case !versionPattern.MatchString(cfg.MinimumMacOS):
			c.Error(problem.Usage, msg("WAP-MINOS-INVALID", fmt.Sprintf("%q is not a macOS version.", cfg.MinimumMacOS)))
		case progMinOS != "" && compareVersions(cfg.MinimumMacOS, progMinOS) < 0:
			c.Error(problem.Failed, msg("WAP-MINOS-TOO-LOW",
				fmt.Sprintf("The minimum macOS %s is lower than %s, which the program was built for.", cfg.MinimumMacOS, progMinOS)))
		default:
			meta.minOS = cfg.MinimumMacOS
		}
	}

	// Executable name.
	meta.executable = cfg.ExecutableName
	if meta.executable == "" && len(cfg.Binaries) > 0 {
		meta.executable = trimArchSuffix(filepath.Base(cfg.Binaries[0].Path))
	}
	if strings.ContainsAny(meta.executable, "/:") {
		c.Error(problem.Usage, msgHow("WAP-NAME-INVALID", fmt.Sprintf("The executable name %q contains / or :.", meta.executable), "Choose a plain file name with --executable-name."))
	}

	// Icon.
	var iconEntries []icns.Entry
	if cfg.Icon != nil && cfg.Icon.Path != "" {
		shape := cmp.Or(cfg.Icon.Shape, ShapeNone)
		if shape != ShapeNone && shape != ShapeSquircle {
			c.Error(problem.Usage, msgHow("WAP-ICON-FORMAT", fmt.Sprintf("%q is not an icon shape.", shape), "Use none or squircle."))
		} else if e, ok := makeIcon(files, cfg.Icon.Path, shape, &c); ok {
			iconEntries = e
			meta.icon = true
		}
	} else {
		c.Note(msg("WAP-ICON-NONE", "No icon was given."))
	}

	// Additional metadata from a file.
	if cfg.PlistFile != "" {
		if data, ok := readInput(files, cfg.PlistFile, "property list", &c); ok {
			v, err := plist.Decode(data)
			m, isDict := v.(map[string]any)
			switch {
			case err != nil:
				c.Error(problem.Failed, msg("WAP-PLIST-FILE", fmt.Sprintf("%s: %v.", cfg.PlistFile, err)))
			case !isDict:
				c.Error(problem.Failed, msg("WAP-PLIST-FILE", fmt.Sprintf("The top level of %s is not a dictionary.", cfg.PlistFile)))
			default:
				checkProtected(m, cfg.PlistFile, &c)
				for k, v := range m {
					if _, set := meta.custom[k]; !set {
						meta.custom[k] = v
					}
				}
			}
		}
	}

	// Payload.
	nodes := []Node{{Path: "Contents", Dir: true, Mode: 0o755}}
	nodes = append(nodes, payload(files, cfg, archs, &c)...)

	if utf8Len(meta.name) > 15 {
		c.Note(msg("WAP-NAME-LONG", fmt.Sprintf("The name %q has %d characters.", meta.name, utf8Len(meta.name))))
	}

	dest := cfg.BundlePath()
	if meta.name != "" && !req.Force {
		if st, err := files.Stat(dest); err == nil && st.Exists {
			c.Error(problem.Failed, msg("WAP-OUTPUT-EXISTS", fmt.Sprintf("%s exists already.", dest)))
		}
	}

	if req.Strict {
		c.Strict()
	}
	if err := c.Err(); err != nil {
		return BuildResult{}, err
	}

	info, err := plist.Encode(meta.infoPlist())
	if err != nil {
		return BuildResult{}, problem.Fail(msg("WAP-PLIST-VALUE", fmt.Sprintf("Info.plist cannot be written: %v.", err)))
	}
	nodes = append(nodes,
		Node{Path: "Contents/Info.plist", Mode: 0o644, Data: info},
		Node{Path: "Contents/PkgInfo", Mode: 0o644, Data: []byte("APPL????")},
		Node{Path: "Contents/MacOS", Dir: true, Mode: 0o755},
		Node{Path: "Contents/MacOS/" + meta.executable, Mode: 0o755, Data: exe},
	)
	if meta.icon {
		nodes = append(nodes,
			Node{Path: "Contents/Resources", Dir: true, Mode: 0o755},
			Node{Path: "Contents/Resources/AppIcon.icns", Mode: 0o644, Data: icns.Encode(iconEntries)})
	}
	nodes = normalizeTree(nodes)

	res := BuildResult{
		Path:         dest,
		Name:         meta.name,
		Identifier:   meta.identifier,
		Version:      meta.version,
		BuildNumber:  cmp.Or(meta.buildNumber, meta.version),
		Executable:   meta.executable,
		Archs:        archs,
		MinimumMacOS: meta.minOS,
		IconSizes:    iconSizes(iconEntries),
		InfoPlist:    string(info),
		DryRun:       req.DryRun,
		Report:       Report{Notes: c.Notes()},
	}
	for _, n := range nodes {
		p := meta.name + ".app/" + n.Path
		if n.Dir {
			p += "/"
		}
		res.Files = append(res.Files, p)
	}
	if req.DryRun {
		return res, nil
	}

	if err := files.WriteTree(dest, nodes, WriteOptions{Force: req.Force, ModTime: req.ModTime}); err != nil {
		return BuildResult{}, problem.From(writeFailed(dest, err))
	}
	if err := selfCheck(files, dest, meta); err != nil {
		return BuildResult{}, err
	}
	return res, nil
}

// checkMetadata validates the values that become Info.plist.
func checkMetadata(cfg Config, c *problem.Collector) metadata {
	m := metadata{
		name:           cfg.Name,
		displayName:    cfg.DisplayName,
		identifier:     cfg.Identifier,
		version:        cfg.Version,
		buildNumber:    cfg.BuildNumber,
		copyright:      cfg.Copyright,
		agent:          cfg.Agent,
		backgroundOnly: cfg.BackgroundOnly,
		urlSchemes:     cfg.URLSchemes,
		environment:    cfg.Environment,
		usage:          map[string]string{},
		custom:         map[string]any{},
	}

	switch {
	case cfg.Name == "":
		c.Error(problem.Usage, msg("WAP-NAME-MISSING", "The bundle has no name."))
	case strings.Contains(cfg.Name, Placeholder):
		c.Error(problem.Usage, msg("WAP-CONFIG-PLACEHOLDER", fmt.Sprintf("The name %q is the placeholder written by init.", cfg.Name)))
	case strings.ContainsAny(cfg.Name, "/:") || strings.HasPrefix(cfg.Name, ".") || strings.TrimSpace(cfg.Name) != cfg.Name:
		c.Error(problem.Usage, msg("WAP-NAME-INVALID", fmt.Sprintf("%q cannot be the name of a bundle.", cfg.Name)))
	}

	switch {
	case cfg.Identifier == "":
		c.Error(problem.Usage, msg("WAP-IDENTIFIER-MISSING", "The bundle has no identifier."))
	case strings.Contains(cfg.Identifier, Placeholder):
		c.Error(problem.Usage, msg("WAP-CONFIG-PLACEHOLDER", fmt.Sprintf("The identifier %q is the placeholder written by init.", cfg.Identifier)))
	case !identifierPattern.MatchString(cfg.Identifier):
		c.Error(problem.Usage, msg("WAP-IDENTIFIER-INVALID", fmt.Sprintf("%q is not in reverse domain notation.", cfg.Identifier)))
	}

	switch {
	case cfg.Version == "":
		c.Error(problem.Usage, msg("WAP-VERSION-MISSING", "The bundle has no version."))
	case !versionPattern.MatchString(cfg.Version):
		c.Error(problem.Usage, msg("WAP-VERSION-INVALID", fmt.Sprintf("The version %q is not one to three numbers separated by dots.", cfg.Version)))
	}
	if cfg.BuildNumber != "" && !versionPattern.MatchString(cfg.BuildNumber) {
		c.Error(problem.Usage, msg("WAP-VERSION-INVALID", fmt.Sprintf("The build number %q is not one to three numbers separated by dots.", cfg.BuildNumber)))
	}

	if cfg.Category != "" {
		if full, e, ok := category(cfg.Category); ok {
			m.category = full
		} else {
			c.Error(problem.Usage, e)
		}
	}

	for _, s := range cfg.URLSchemes {
		if !schemePattern.MatchString(s) {
			c.Error(problem.Usage, msg("WAP-URL-SCHEME-INVALID", fmt.Sprintf("%q is not a URL scheme.", s)))
		}
	}

	kinds := UsageKinds()
	var known []string
	for k := range kinds {
		known = append(known, k)
	}
	slices.Sort(known)
	for kind, text := range cfg.UsageDescriptions {
		if _, ok := kinds[kind]; !ok {
			how := "Use one of: " + strings.Join(known, ", ") + "."
			if s, ok := suggest.Closest(kind, known); ok {
				how = fmt.Sprintf("Did you mean %q? %s", s, how)
			}
			c.Error(problem.Usage, msgHow("WAP-USAGE-UNKNOWN", fmt.Sprintf("%q is not a kind of access.", kind), how))
			continue
		}
		m.usage[kind] = text
	}

	for k, v := range cfg.Plist {
		n, err := normalizePlist(k, v)
		if err != nil {
			c.Error(problem.Usage, msg("WAP-PLIST-VALUE", err.Error()+"."))
			continue
		}
		m.custom[k] = n
	}
	checkProtected(m.custom, "the additional metadata", c)
	return m
}

func checkProtected(custom map[string]any, source string, c *problem.Collector) {
	protected := ProtectedKeys()
	keys := make([]string, 0, len(custom))
	for k := range custom {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		opt, ok := protected[k]
		if !ok {
			continue
		}
		how := "Remove the key; wdyapppack sets it from the bundle itself."
		if opt != "" {
			how = fmt.Sprintf("Remove the key and use %s instead.", opt)
		}
		c.Error(problem.Usage, msgHow("WAP-PLIST-PROTECTED", fmt.Sprintf("%s sets %s.", source, k), how))
	}
}

// payload copies resources, frameworks and helpers. archs are the
// architectures of the main program, which every helper must cover; nil when
// the main program could not be read.
func payload(files Files, cfg Config, archs []string, c *problem.Collector) []Node {
	var nodes []Node
	add := func(from, to, role string) {
		st, err := files.Stat(from)
		if err != nil || !st.Exists {
			c.Error(problem.Failed, msg("WAP-FILE-NOT-FOUND", fmt.Sprintf("The %s %s does not exist.", role, from)))
			return
		}
		if !st.Dir {
			nodes = append(nodes, Node{Path: to, Mode: st.Mode.Perm(), From: from})
			return
		}
		nodes = append(nodes, Node{Path: to, Dir: true, Mode: 0o755})
		list, err := files.List(from)
		if err != nil {
			c.Error(problem.Failed, msg("WAP-FILE-UNREADABLE", fmt.Sprintf("The %s %s cannot be listed: %v.", role, from, err)))
			return
		}
		for _, f := range list {
			n := Node{Path: to + "/" + f.Path, Mode: f.Mode.Perm(), Dir: f.Dir, Link: f.Link}
			if !f.Dir && f.Link == "" {
				n.From = filepath.Join(from, filepath.FromSlash(f.Path))
			}
			nodes = append(nodes, n)
		}
	}

	for _, r := range cfg.Resources {
		to := r.To
		if to == "" {
			to = filepath.Base(r.From)
		}
		to = path.Clean(filepath.ToSlash(to))
		if strings.HasPrefix(to, "../") || to == ".." || path.IsAbs(to) {
			c.Error(problem.Usage, msgHow("WAP-NAME-INVALID", fmt.Sprintf("The resource target %q leaves Contents/Resources.", r.To), "Give a path below Contents/Resources, without .."))
			continue
		}
		add(r.From, "Contents/Resources/"+to, "resource")
	}
	for _, f := range cfg.Frameworks {
		add(f, "Contents/Frameworks/"+filepath.Base(f), "framework")
	}
	seen := map[string]bool{}
	for _, h := range cfg.Helpers {
		if n, ok := helper(files, h, archs, seen, c); ok {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

// helper reads the programs of a helper, merges several into a universal one
// and checks that it runs wherever the main program does.
func helper(files Files, h Helper, archs []string, seen map[string]bool, c *problem.Collector) (Node, bool) {
	if len(h.Binaries) == 0 {
		what := "A helper names no program."
		if h.Name != "" {
			what = fmt.Sprintf("The helper %s names no program.", h.Name)
		}
		c.Error(problem.Usage, msgHow("WAP-BINARY-MISSING", what,
			`Give its program under "binaries": [{"path": "dist/helper"}], one entry per architecture.`))
		return Node{}, false
	}
	name := h.FileName()
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/:") {
		c.Error(problem.Usage, msgHow("WAP-NAME-INVALID", fmt.Sprintf("The helper name %q cannot be a file name in Contents/Helpers.", name),
			"Choose a plain file name without / and :."))
		return Node{}, false
	}
	if seen[name] {
		c.Error(problem.Usage, msg("WAP-HELPER-DUPLICATE", fmt.Sprintf("Two helpers are named %s.", name)))
		return Node{}, false
	}
	seen[name] = true

	var progs []program
	for _, b := range h.Binaries {
		if p, ok := loadProgram(files, b, "helper", c); ok {
			programNotes(p, c)
			progs = append(progs, p)
		}
	}
	if len(progs) != len(h.Binaries) {
		return Node{}, false
	}
	data, helperArchs, _, ok := assemble(progs, c)
	if !ok {
		return Node{}, false
	}
	var missing []string
	for _, a := range archs {
		if !slices.Contains(helperArchs, a) {
			missing = append(missing, a)
		}
	}
	if len(missing) > 0 {
		c.Error(problem.Failed, msg("WAP-HELPER-ARCH-MISSING",
			fmt.Sprintf("The helper %s is built for %s, but the program is also built for %s.", name, joinArchs(helperArchs), joinArchs(missing))))
		return Node{}, false
	}
	return Node{Path: "Contents/Helpers/" + name, Mode: 0o755, Data: data}, true
}

// normalizeTree adds the directories every entry needs, removes duplicates and
// sorts, so that the same input always produces the same tree.
func normalizeTree(nodes []Node) []Node {
	byPath := map[string]Node{}
	for _, n := range nodes {
		if prev, ok := byPath[n.Path]; ok && prev.Dir && n.Dir {
			continue
		}
		byPath[n.Path] = n
		for dir := path.Dir(n.Path); dir != "." && dir != "/"; dir = path.Dir(dir) {
			if _, ok := byPath[dir]; !ok {
				byPath[dir] = Node{Path: dir, Dir: true, Mode: 0o755}
			}
		}
	}
	out := make([]Node, 0, len(byPath))
	for _, n := range byPath {
		out = append(out, n)
	}
	slices.SortFunc(out, func(a, b Node) int { return cmp.Compare(a.Path, b.Path) })
	return out
}

func writeFailed(dest string, err error) error {
	if errors.Is(err, fs.ErrExist) {
		return problem.Fail(msg("WAP-OUTPUT-EXISTS", fmt.Sprintf("%s exists already.", dest)))
	}
	return problem.Fail(msg("WAP-WRITE-FAILED", fmt.Sprintf("%s could not be written: %v.", dest, err)))
}

// selfCheck reads the bundle back and checks what macOS will check first.
func selfCheck(files Files, dest string, meta metadata) error {
	fail := func(what string) error {
		return problem.Fail(msg("WAP-SELF-CHECK", fmt.Sprintf("%s: %s.", dest, what)))
	}
	data, err := files.Read(filepath.Join(dest, "Contents", "Info.plist"))
	if err != nil {
		return fail("Info.plist cannot be read back")
	}
	v, err := plist.Decode(data)
	p, ok := v.(map[string]any)
	if err != nil || !ok {
		return fail("Info.plist cannot be parsed back")
	}
	for _, k := range []string{"CFBundleExecutable", "CFBundleIdentifier", "CFBundleName", "CFBundlePackageType", "CFBundleShortVersionString", "CFBundleVersion"} {
		if s, _ := p[k].(string); s == "" {
			return fail("Info.plist lacks " + k)
		}
	}
	st, err := files.Stat(filepath.Join(dest, "Contents", "MacOS", meta.executable))
	if err != nil || !st.Exists || st.Mode&0o111 == 0 {
		return fail("the program is missing or not executable")
	}
	if meta.icon {
		st, err := files.Stat(filepath.Join(dest, "Contents", "Resources", "AppIcon.icns"))
		if err != nil || !st.Exists {
			return fail("the icon named in Info.plist is missing")
		}
	}
	return nil
}

func utf8Len(s string) int { return len([]rune(s)) }
