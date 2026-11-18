// Package clibundle exposes the bundle context on the command line.
//
// Each command takes the use cases it runs, for the same reason a route does:
// a command that took the bundle of all use cases could not be tested without
// building the whole context.
package clibundle

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/worldiety/wdyapppack/app/bundle"
	"github.com/worldiety/wdyapppack/foundation/cli"
	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/macho"
)

// Env looks up an environment variable.
type Env func(name string) (string, bool)

const platformsAll = "every operating system: nothing from Apple is needed."

// Build is the command that creates a bundle.
func Build(load bundle.LoadConfig, build bundle.BuildBundle, env Env) cli.Command {
	return cli.Command{
		Name:    "build",
		Group:   "Create",
		Summary: "create a .app bundle from a prebuilt program",
		Description: cli.UseCaseHelp[bundle.BuildBundle]() + `

Every value can come from the configuration file (wdyapppack.json in the current directory, or the file named by --config or WDYAPPPACK_CONFIG) or from an option. An option given on the command line wins over the file: a single value replaces the file's, a repeated option replaces the file's list, and --usage, --env and --plist add to or override single keys.

Paths in the configuration file are relative to the file; paths given as options are relative to the current directory.`,
		Flags: []cli.Flag{
			{Name: "config", Type: cli.String, Value: "file", Default: bundle.DefaultConfigFile + " if it exists", Usage: "The configuration file. Also read from the environment variable WDYAPPPACK_CONFIG."},
			{Name: "binary", Type: cli.List, Value: "[arch=]path", Config: "binaries", Usage: "The program to package. Give it once per architecture to build a universal bundle; the architecture prefix (arm64=, amd64=) is optional and checked against the program."},
			{Name: "name", Type: cli.String, Value: "text", Config: "name", Usage: "The application name, shown in the Finder and the menu bar. The bundle is written as <name>.app."},
			{Name: "display-name", Type: cli.String, Value: "text", Config: "displayName", Usage: "A longer name for the Finder and the Dock."},
			{Name: "identifier", Type: cli.String, Value: "reverse.dns", Config: "identifier", Usage: "The bundle identifier, e.g. com.example.mytool. Keep it the same across versions."},
			{Name: "version", Type: cli.String, Value: "x.y.z", Config: "version", Usage: "The version users see: one to three numbers separated by dots."},
			{Name: "build-number", Type: cli.String, Value: "x.y.z", Config: "buildNumber", Default: "the version", Usage: "The build number, increasing with every build handed out."},
			{Name: "executable-name", Type: cli.String, Value: "file", Config: "executableName", Default: "the program's file name without architecture suffix", Usage: "The file name of the program inside Contents/MacOS."},
			{Name: "icon", Type: cli.String, Value: "path", Config: "icon.path", Usage: "A square PNG or JPEG (ideally 1024×1024 with a transparent background), an .icns file or an .iconset directory."},
			{Name: "icon-shape", Type: cli.Enum, Value: "none|squircle", Enum: []string{bundle.ShapeNone, bundle.ShapeSquircle}, Default: "none", Config: "icon.shape", Usage: "squircle cuts the image into the rounded square of macOS 26, with the margin of Apple's grid."},
			{Name: "min-macos", Type: cli.String, Value: "version", Config: "minimumMacOS", Default: "the program's own, else 11.0", Usage: "The oldest macOS the application runs on."},
			{Name: "category", Type: cli.String, Value: "name", Config: "category", Usage: "The application category, e.g. developer-tools. See: wdyapppack help categories."},
			{Name: "copyright", Type: cli.String, Value: "text", Config: "copyright", Usage: "The copyright notice shown in the Finder."},
			{Name: "agent", Type: cli.Bool, Config: "agent", Usage: "Hide the application from the Dock and the application switcher, for menu bar tools."},
			{Name: "background-only", Type: cli.Bool, Config: "backgroundOnly", Usage: "Mark a program without any user interface."},
			{Name: "resource", Type: cli.List, Value: "path[:dest]", Config: "resources", Usage: "A file or directory to copy into Contents/Resources, optionally under another name."},
			{Name: "framework", Type: cli.List, Value: "path", Config: "frameworks", Usage: "A library (.dylib) or framework to copy into Contents/Frameworks."},
			{Name: "helper", Type: cli.List, Value: "path", Config: "helpers", Usage: "Another program to copy into Contents/Helpers. It is checked like the main program."},
			{Name: "url-scheme", Type: cli.List, Value: "scheme", Config: "urlSchemes", Usage: "A URL scheme the application opens, without ://."},
			{Name: "usage", Type: cli.List, Value: "kind=text", Config: "usageDescriptions", Usage: "Why the application needs a kind of access, e.g. camera=\"Scans QR codes.\". Kinds: wdyapppack help info-plist."},
			{Name: "env", Type: cli.List, Value: "KEY=VALUE", Config: "environment", Usage: "An environment variable macOS sets when it starts the application."},
			{Name: "plist", Type: cli.List, Value: "key[:type]=value", Config: "plist", Usage: "Any further Info.plist key. Types: string (default), bool, int, real, array (comma separated)."},
			{Name: "plist-file", Type: cli.String, Value: "path", Config: "plistFile", Usage: "An XML property list whose keys are merged into Info.plist."},
			{Name: "out", Type: cli.String, Value: "dir", Config: "out", Default: "dist", Usage: "The directory the bundle is written to."},
			{Name: "timestamp", Type: cli.String, Value: "RFC3339", Default: "SOURCE_DATE_EPOCH if set, else now", Usage: "The time given to every file, so that two builds of the same input are identical."},
			{Name: "strict", Type: cli.Bool, Usage: "Treat every note as an error, for build servers."},
		},
		Writes:   true,
		Replaces: true,
		Examples: []cli.Example{
			cli.Ex("package one program with everything on the command line", `wdyapppack build --binary dist/mytool --name "My Tool" --identifier com.example.mytool --version 1.2.0 --icon assets/icon.png`),
			cli.Ex("build a universal bundle for Apple silicon and Intel", `wdyapppack build --binary arm64=dist/mytool-arm64 --binary amd64=dist/mytool-amd64 --name "My Tool" --identifier com.example.mytool --version 1.2.0`),
			cli.Ex("build from wdyapppack.json, overriding the version", `wdyapppack build --version 1.2.1`),
			cli.Ex("see what would be written, without writing", `wdyapppack build --dry-run`),
			cli.Ex("a menu bar tool that asks for the camera", `wdyapppack build --agent --usage "camera=Scans QR codes." --plist NSSupportsAutomaticGraphicsSwitching:bool=true`),
		},
		SeeAlso:   []string{"bundle", "config", "info-plist", "icons", "binaries", "sign", "zip"},
		Platforms: platformsAll,
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			path, optional := configPath(inv, env)
			loaded, err := load(inv.Subject, bundle.LoadConfigRequest{Path: path, Optional: optional, Lookup: env})
			if err != nil {
				return cli.Report{}, err
			}
			cfg := loaded.Config
			if err := applyFlags(inv, &cfg); err != nil {
				return cli.Report{}, err
			}
			when, err := timestamp(inv, env)
			if err != nil {
				return cli.Report{}, err
			}
			inv.Logf("configuration: %s (found: %v)", loaded.Path, loaded.Found)

			res, err := build(inv.Subject, bundle.BuildRequest{
				Config: cfg, ModTime: when, Strict: inv.Bool("strict"), Force: inv.Force, DryRun: inv.DryRun,
			})
			if err != nil {
				return cli.Report{}, err
			}
			return cli.Report{Text: buildText(inv, res), Data: res, Notes: res.Notes, Next: buildNext(res)}, nil
		},
	}
}

func configPath(inv *cli.Invocation, env Env) (string, bool) {
	if inv.IsSet("config") {
		return inv.String("config"), false
	}
	if p, ok := env("WDYAPPPACK_CONFIG"); ok && p != "" {
		return p, false
	}
	return bundle.DefaultConfigFile, true
}

func timestamp(inv *cli.Invocation, env Env) (time.Time, error) {
	if s := inv.String("timestamp"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}, usage("--timestamp %q is not a time in RFC 3339 form.", "Write it like 2026-01-31T12:00:00Z.", s)
		}
		return t, nil
	}
	if s, ok := env("SOURCE_DATE_EPOCH"); ok && s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Time{}, usage("SOURCE_DATE_EPOCH %q is not a number of seconds.", "Set it to the Unix time of the last commit, e.g. $(git log -1 --format=%ct).", s)
		}
		return time.Unix(n, 0).UTC(), nil
	}
	return time.Time{}, nil
}

func usage(what, how string, args ...any) error {
	return problem.Misuse(problem.Message{
		Code: "WAP-OPTION-VALUE",
		What: fmt.Sprintf(what, args...),
		Why:  "The value cannot be used as it is written.",
		How:  how,
		See:  "build",
	})
}

// applyFlags lays the options given on the command line over the
// configuration. Only options actually given count, so that an option's
// default never overrides the file.
func applyFlags(inv *cli.Invocation, c *bundle.Config) error {
	var errs problem.Collector
	str := func(flag string, p *string) {
		if inv.IsSet(flag) {
			*p = inv.String(flag)
		}
	}
	boolean := func(flag string, p *bool) {
		if inv.IsSet(flag) {
			*p = inv.Bool(flag)
		}
	}
	str("name", &c.Name)
	str("display-name", &c.DisplayName)
	str("identifier", &c.Identifier)
	str("version", &c.Version)
	str("build-number", &c.BuildNumber)
	str("executable-name", &c.ExecutableName)
	str("min-macos", &c.MinimumMacOS)
	str("category", &c.Category)
	str("copyright", &c.Copyright)
	str("plist-file", &c.PlistFile)
	str("out", &c.Out)
	boolean("agent", &c.Agent)
	boolean("background-only", &c.BackgroundOnly)

	if inv.IsSet("icon") || inv.IsSet("icon-shape") {
		if c.Icon == nil {
			c.Icon = &bundle.Icon{}
		}
		str("icon", &c.Icon.Path)
		str("icon-shape", &c.Icon.Shape)
	}
	if inv.IsSet("binary") {
		c.Binaries = nil
		for _, v := range inv.List("binary") {
			c.Binaries = append(c.Binaries, ParseBinary(v))
		}
	}
	if inv.IsSet("resource") {
		c.Resources = nil
		for _, v := range inv.List("resource") {
			from, to := v, ""
			if i := strings.LastIndex(v, ":"); i > 1 {
				from, to = v[:i], v[i+1:]
			}
			c.Resources = append(c.Resources, bundle.Resource{From: from, To: to})
		}
	}
	if inv.IsSet("framework") {
		c.Frameworks = inv.List("framework")
	}
	if inv.IsSet("helper") {
		c.Helpers = inv.List("helper")
	}
	if inv.IsSet("url-scheme") {
		c.URLSchemes = inv.List("url-scheme")
	}
	pairs := func(flag string, into *map[string]string) {
		for _, v := range inv.List(flag) {
			k, val, ok := strings.Cut(v, "=")
			if !ok || k == "" {
				errs.Merge(usage("--%s %q is not of the form key=value.", "Write --"+flag+" key=value.", flag, v))
				continue
			}
			if *into == nil {
				*into = map[string]string{}
			}
			(*into)[k] = val
		}
	}
	pairs("usage", &c.UsageDescriptions)
	pairs("env", &c.Environment)
	for _, v := range inv.List("plist") {
		k, val, err := ParsePlistValue(v)
		if err != nil {
			errs.Merge(err)
			continue
		}
		if c.Plist == nil {
			c.Plist = map[string]any{}
		}
		c.Plist[k] = val
	}
	return errs.Err()
}

// ParseBinary reads [arch=]path. The prefix counts as an architecture only
// when it is one, so that a path containing = is still a path.
func ParseBinary(v string) bundle.Binary {
	if arch, p, ok := strings.Cut(v, "="); ok {
		if _, known := macho.NormalizeArch(arch); known {
			return bundle.Binary{Arch: arch, Path: p}
		}
	}
	return bundle.Binary{Path: v}
}

// ParsePlistValue reads key[:type]=value.
func ParsePlistValue(v string) (string, any, error) {
	spec, val, ok := strings.Cut(v, "=")
	if !ok || spec == "" {
		return "", nil, plistUsage(v, "it has no =")
	}
	key, typ, _ := strings.Cut(spec, ":")
	switch typ {
	case "", "string":
		return key, val, nil
	case "bool":
		b, err := strconv.ParseBool(val)
		if err != nil {
			return "", nil, plistUsage(v, "the value is not true or false")
		}
		return key, b, nil
	case "int", "integer":
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return "", nil, plistUsage(v, "the value is not a whole number")
		}
		return key, json.Number(strconv.FormatInt(n, 10)), nil
	case "real":
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return "", nil, plistUsage(v, "the value is not a number")
		}
		return key, json.Number(strconv.FormatFloat(f, 'g', -1, 64) + dotIfWhole(f)), nil
	case "array":
		var out []any
		if val != "" {
			for _, s := range strings.Split(val, ",") {
				out = append(out, s)
			}
		}
		return key, out, nil
	}
	return "", nil, plistUsage(v, fmt.Sprintf("%q is not a type", typ))
}

func dotIfWhole(f float64) string {
	if f == float64(int64(f)) {
		return ".0"
	}
	return ""
}

func plistUsage(v, reason string) error {
	return problem.Misuse(problem.Message{
		Code: "WAP-PLIST-VALUE",
		What: fmt.Sprintf("--plist %q cannot be used: %s.", v, reason),
		Why:  "Each additional key needs a name, a type and a value of that type.",
		How:  "Write --plist Key=text, Key:bool=true, Key:int=3, Key:real=1.5 or Key:array=a,b.",
		See:  "info-plist",
	})
}

func buildText(inv *cli.Invocation, r bundle.BuildResult) string {
	var b strings.Builder
	verb := "Created"
	if r.DryRun {
		verb = "Would create (dry run, nothing written)"
	}
	fmt.Fprintf(&b, "%s %s\n\n", verb, r.Path)
	row := func(k, v string) { fmt.Fprintf(&b, "  %-12s %s\n", k, v) }
	row("name", r.Name)
	row("identifier", r.Identifier)
	row("version", r.Version+" (build "+r.BuildNumber+")")
	row("program", fmt.Sprintf("Contents/MacOS/%s for %s, macOS %s or later", r.Executable, strings.Join(r.Archs, " and "), r.MinimumMacOS))
	if len(r.IconSizes) > 0 {
		row("icon", fmt.Sprintf("Contents/Resources/AppIcon.icns, %d to %d pixels", r.IconSizes[0], r.IconSizes[len(r.IconSizes)-1]))
	} else {
		row("icon", "none")
	}
	b.WriteString("\nCONTENTS\n")
	fmt.Fprintf(&b, "  %s.app/\n", r.Name)
	for _, f := range r.Files {
		depth := strings.Count(strings.TrimSuffix(f, "/"), "/")
		name := f[strings.LastIndex(strings.TrimSuffix(f, "/"), "/")+1:]
		fmt.Fprintf(&b, "  %s%s\n", strings.Repeat("  ", depth), name)
	}
	if r.DryRun || inv.Verbose {
		b.WriteString("\nINFO.PLIST\n")
		for _, l := range strings.Split(strings.TrimRight(r.InfoPlist, "\n"), "\n") {
			b.WriteString("  " + l + "\n")
		}
	}
	return b.String()
}

func buildNext(r bundle.BuildResult) []string {
	if r.DryRun {
		return []string{"wdyapppack build   (without --dry-run, to write it)"}
	}
	app := cli.Quote(r.Path)
	return []string{
		"wdyapppack sign " + app + " --adhoc            # to start it on this Mac",
		"wdyapppack sign " + app + " --identity <Developer ID>   # to hand it out, then: wdyapppack notarize",
		"wdyapppack zip " + app + "   or   wdyapppack dmg " + app,
	}
}

// Init is the command that writes a starter configuration.
func Init(write bundle.WriteConfig, extra map[string]json.RawMessage) cli.Command {
	return cli.Command{
		Name:        "init",
		Group:       "Create",
		Summary:     "write a starter configuration file",
		Description: cli.UseCaseHelp[bundle.WriteConfig](),
		Flags: []cli.Flag{
			{Name: "binary", Type: cli.String, Value: "path", Usage: "The program to package; its name, architecture and minimum macOS are filled in."},
			{Name: "name", Type: cli.String, Value: "text", Usage: "The application name."},
			{Name: "identifier", Type: cli.String, Value: "reverse.dns", Usage: "The bundle identifier, e.g. com.example.mytool."},
			{Name: "out", Type: cli.String, Value: "file", Default: bundle.DefaultConfigFile, Usage: "Where to write the configuration."},
		},
		Writes:   true,
		Replaces: true,
		Examples: []cli.Example{
			cli.Ex("start from a program", `wdyapppack init --binary dist/mytool --identifier com.example.mytool`),
			cli.Ex("see the file without writing it", `wdyapppack init --dry-run`),
		},
		SeeAlso:   []string{"config", "build", "schema"},
		Platforms: platformsAll,
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			res, err := write(inv.Subject, bundle.WriteConfigRequest{
				Out: inv.String("out"), Binary: inv.String("binary"), Name: inv.String("name"),
				Identifier: inv.String("identifier"), Extra: extra, Force: inv.Force, DryRun: inv.DryRun,
			})
			if err != nil {
				return cli.Report{}, err
			}
			text := "Wrote " + res.Path + "\n"
			if res.DryRun {
				text = "Would write " + res.Path + " (dry run, nothing written):\n\n" + res.Content
			}
			next := []string{
				"edit " + cli.Quote(res.Path) + " and replace every " + bundle.Placeholder,
				"wdyapppack schema > wdyapppack.schema.json   # completion and checks in your editor",
				"wdyapppack build --dry-run",
			}
			return cli.Report{Text: text, Data: res, Notes: res.Notes, Next: next}, nil
		},
	}
}

// Icon is the command that converts an image into an icon.
func Icon(convert bundle.ConvertIcon, env Env) cli.Command {
	return cli.Command{
		Name:        "icon",
		Group:       "Create",
		Summary:     "convert a PNG or JPEG image into a macOS .icns icon",
		Description: cli.UseCaseHelp[bundle.ConvertIcon](),
		Args:        []cli.Arg{{Name: "image", Usage: "a square PNG or JPEG image, ideally 1024×1024 pixels", Required: true}},
		Flags: []cli.Flag{
			{Name: "out", Type: cli.String, Value: "path", Default: "the image's name with .icns or .iconset", Usage: "Where to write the icon."},
			{Name: "shape", Type: cli.Enum, Value: "none|squircle", Enum: []string{bundle.ShapeNone, bundle.ShapeSquircle}, Default: "none", Usage: "squircle cuts the image into the rounded square of macOS 26."},
			{Name: "iconset", Type: cli.Bool, Usage: "Write an .iconset directory of PNG files instead of an .icns file."},
			{Name: "timestamp", Type: cli.String, Value: "RFC3339", Default: "SOURCE_DATE_EPOCH if set, else now", Usage: "The time given to the files written."},
		},
		Writes:   true,
		Replaces: true,
		Examples: []cli.Example{
			cli.Ex("make AppIcon.icns", `wdyapppack icon assets/logo.png --out AppIcon.icns`),
			cli.Ex("cut the macOS 26 shape and write an icon set", `wdyapppack icon assets/logo.png --shape squircle --iconset`),
		},
		SeeAlso:   []string{"icons", "build"},
		Platforms: platformsAll,
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			when, err := timestamp(inv, env)
			if err != nil {
				return cli.Report{}, err
			}
			res, err := convert(inv.Subject, bundle.ConvertIconRequest{
				In: inv.Arg(0), Out: inv.String("out"), Shape: inv.String("shape"), Iconset: inv.Bool("iconset"),
				ModTime: when, Force: inv.Force, DryRun: inv.DryRun,
			})
			if err != nil {
				return cli.Report{}, err
			}
			verb := "Wrote"
			if res.DryRun {
				verb = "Would write (dry run)"
			}
			sizes := make([]string, len(res.Sizes))
			for i, s := range res.Sizes {
				sizes[i] = strconv.Itoa(s)
			}
			text := fmt.Sprintf("%s %s with %s pixels\n", verb, res.Path, strings.Join(sizes, ", "))
			return cli.Report{Text: text, Data: res, Notes: res.Notes,
				Next: []string{"wdyapppack build --icon " + cli.Quote(res.Path)}}, nil
		},
	}
}

// Universal is the command that merges programs into a universal one.
func Universal(merge bundle.MergeUniversal) cli.Command {
	return cli.Command{
		Name:        "universal",
		Group:       "Create",
		Summary:     "merge arm64 and x86_64 programs into one",
		Description: cli.UseCaseHelp[bundle.MergeUniversal](),
		Flags: []cli.Flag{
			{Name: "in", Type: cli.List, Value: "[arch=]path", Usage: "A program to merge; give one per architecture. The architecture prefix is optional and checked."},
			{Name: "out", Type: cli.String, Value: "path", Default: "the first program's name with -universal", Usage: "Where to write the universal program."},
		},
		Writes:   true,
		Replaces: true,
		Examples: []cli.Example{
			cli.Ex("merge two Go builds", `wdyapppack universal --in arm64=dist/mytool-arm64 --in amd64=dist/mytool-amd64 --out dist/mytool`),
		},
		SeeAlso:   []string{"architectures", "binaries"},
		Platforms: platformsAll,
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			var in []bundle.Binary
			for _, v := range inv.List("in") {
				in = append(in, ParseBinary(v))
			}
			res, err := merge(inv.Subject, bundle.MergeRequest{Inputs: in, Out: inv.String("out"), Force: inv.Force, DryRun: inv.DryRun})
			if err != nil {
				return cli.Report{}, err
			}
			verb := "Wrote"
			if res.DryRun {
				verb = "Would write (dry run)"
			}
			text := fmt.Sprintf("%s %s for %s (%d bytes)\n", verb, res.Path, strings.Join(res.Archs, " and "), res.Size)
			return cli.Report{Text: text, Data: res, Notes: res.Notes,
				Next: []string{"wdyapppack build --binary " + cli.Quote(res.Path)}}, nil
		},
	}
}
