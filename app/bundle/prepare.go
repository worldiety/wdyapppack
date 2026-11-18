package bundle

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/icns"
)

// LoadConfigRequest names the configuration file to read.
type LoadConfigRequest struct {
	// Path of the file; DefaultConfigFile when empty.
	Path string
	// Optional allows the file to be absent, which is what happens when no
	// file was named and the default does not exist.
	Optional bool
	// Lookup resolves environment variables; nil treats every one as unset.
	Lookup func(string) (string, bool)
}

// LoadedConfig is the configuration read, with relative paths resolved
// against the file's directory.
type LoadedConfig struct {
	Config Config `json:"config"`
	Path   string `json:"path"`
	Found  bool   `json:"found"`
}

func loadConfig(files Files, req LoadConfigRequest) (LoadedConfig, error) {
	p := cmp.Or(req.Path, DefaultConfigFile)
	st, err := files.Stat(p)
	if err == nil && !st.Exists {
		if req.Optional {
			return LoadedConfig{Path: p}, nil
		}
		return LoadedConfig{}, problem.Misuse(msg("WAP-CONFIG-MISSING", fmt.Sprintf("The configuration file %s does not exist.", p)))
	}
	data, err := files.Read(p)
	if err != nil {
		return LoadedConfig{}, problem.Fail(msg("WAP-FILE-UNREADABLE", fmt.Sprintf("The configuration file %s cannot be read: %v.", p, err)))
	}
	cfg, err := decodeConfig(p, data, []string{"sign", "notarize", "dmg"})
	if err != nil {
		return LoadedConfig{}, err
	}
	lookup := req.Lookup
	if lookup == nil {
		lookup = func(string) (string, bool) { return "", false }
	}
	if err := expand(&cfg, lookup); err != nil {
		return LoadedConfig{}, err
	}
	resolve(&cfg, filepath.Dir(p))
	return LoadedConfig{Config: cfg, Path: p, Found: true}, nil
}

// WriteConfigRequest asks for a starter configuration.
type WriteConfigRequest struct {
	// Out is where to write it; DefaultConfigFile when empty.
	Out string
	// Binary, Name and Identifier fill in what is known already.
	Binary     string
	Name       string
	Identifier string
	// Extra holds the starter sections of other parts of the program, by
	// field name, so that the file describes the whole way to a download.
	Extra  map[string]json.RawMessage
	Force  bool
	DryRun bool
}

// WrittenConfig is the starter configuration.
type WrittenConfig struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	DryRun  bool   `json:"dryRun"`
	Report
}

// starter is the configuration as init writes it: every field present, in the
// order of the documentation, so that the file shows what can be set.
type starter struct {
	Schema            string            `json:"$schema"`
	Name              string            `json:"name"`
	DisplayName       string            `json:"displayName"`
	Identifier        string            `json:"identifier"`
	Version           string            `json:"version"`
	BuildNumber       string            `json:"buildNumber"`
	Binaries          []Binary          `json:"binaries"`
	ExecutableName    string            `json:"executableName"`
	Icon              Icon              `json:"icon"`
	MinimumMacOS      string            `json:"minimumMacOS"`
	Category          string            `json:"category"`
	Copyright         string            `json:"copyright"`
	Agent             bool              `json:"agent"`
	BackgroundOnly    bool              `json:"backgroundOnly"`
	Resources         []Resource        `json:"resources"`
	Frameworks        []string          `json:"frameworks"`
	Helpers           []string          `json:"helpers"`
	URLSchemes        []string          `json:"urlSchemes"`
	UsageDescriptions map[string]string `json:"usageDescriptions"`
	Environment       map[string]string `json:"environment"`
	Plist             map[string]any    `json:"plist"`
	PlistFile         string            `json:"plistFile"`
	Out               string            `json:"out"`
}

func writeConfig(files Files, req WriteConfigRequest) (WrittenConfig, error) {
	out := cmp.Or(req.Out, DefaultConfigFile)
	var c problem.Collector

	s := starter{
		Schema:            "wdyapppack.schema.json",
		Name:              cmp.Or(req.Name, Placeholder),
		Identifier:        cmp.Or(req.Identifier, "com.example."+Placeholder),
		Version:           "1.0.0",
		Binaries:          []Binary{{Path: Placeholder}},
		Icon:              Icon{Shape: ShapeNone},
		Resources:         []Resource{},
		Frameworks:        []string{},
		Helpers:           []string{},
		URLSchemes:        []string{},
		UsageDescriptions: map[string]string{},
		Environment:       map[string]string{},
		Plist:             map[string]any{},
		Out:               "dist",
	}

	if req.Binary != "" {
		if p, ok := loadProgram(files, Binary{Path: req.Binary}, "program", &c); ok {
			programNotes(p, &c)
			base := trimArchSuffix(filepath.Base(req.Binary))
			if req.Name == "" {
				s.Name = base
			}
			s.ExecutableName = base
			rel := req.Binary
			if r, err := filepath.Rel(filepath.Dir(out), req.Binary); err == nil && filepath.IsAbs(out) == filepath.IsAbs(req.Binary) {
				rel = r
			}
			archs := p.info.Archs()
			arch := ""
			if len(archs) == 1 {
				arch = archs[0]
			}
			s.Binaries = []Binary{{Arch: arch, Path: filepath.ToSlash(rel)}}
			for _, sl := range p.info.Slices {
				if s.MinimumMacOS == "" || compareVersions(sl.MinOS, s.MinimumMacOS) > 0 {
					s.MinimumMacOS = sl.MinOS
				}
			}
		}
	}
	if req.Identifier != "" && !identifierPattern.MatchString(req.Identifier) {
		c.Error(problem.Usage, msg("WAP-IDENTIFIER-INVALID", fmt.Sprintf("%q is not in reverse domain notation.", req.Identifier)))
	}
	if !req.Force {
		if st, err := files.Stat(out); err == nil && st.Exists {
			c.Error(problem.Failed, msg("WAP-OUTPUT-EXISTS", fmt.Sprintf("%s exists already.", out)))
		}
	}
	if err := c.Err(); err != nil {
		return WrittenConfig{}, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return WrittenConfig{}, err
	}
	content := buf.String()
	if len(req.Extra) > 0 {
		keys := make([]string, 0, len(req.Extra))
		for k := range req.Extra {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		content = strings.TrimSuffix(strings.TrimSpace(content), "}")
		content = strings.TrimRight(content, "\n ")
		for _, k := range keys {
			var ind bytes.Buffer
			if err := json.Indent(&ind, req.Extra[k], "  ", "  "); err != nil {
				return WrittenConfig{}, err
			}
			content += fmt.Sprintf(",\n  %q: %s", k, ind.String())
		}
		content += "\n}\n"
	}

	res := WrittenConfig{Path: out, Content: content, DryRun: req.DryRun, Report: Report{Notes: c.Notes()}}
	if req.DryRun {
		return res, nil
	}
	if err := files.WriteFile(out, []byte(content), 0o644, WriteOptions{Force: req.Force}); err != nil {
		return WrittenConfig{}, writeFailed(out, err)
	}
	return res, nil
}

// ConvertIconRequest asks for an icon to be made from an image.
type ConvertIconRequest struct {
	In  string
	Out string
	// Shape is ShapeNone or ShapeSquircle.
	Shape string
	// Iconset writes an .iconset directory instead of an .icns file.
	Iconset bool
	ModTime time.Time
	Force   bool
	DryRun  bool
}

// ConvertedIcon describes the icon written.
type ConvertedIcon struct {
	Path    string `json:"path"`
	Sizes   []int  `json:"sizes"`
	Iconset bool   `json:"iconset"`
	DryRun  bool   `json:"dryRun"`
	Report
}

func convertIcon(files Files, req ConvertIconRequest) (ConvertedIcon, error) {
	var c problem.Collector
	shape := cmp.Or(req.Shape, ShapeNone)
	if shape != ShapeNone && shape != ShapeSquircle {
		c.Error(problem.Usage, msgHow("WAP-ICON-FORMAT", fmt.Sprintf("%q is not an icon shape.", shape), "Use none or squircle."))
		return ConvertedIcon{}, c.Err()
	}
	entries, ok := makeIcon(files, req.In, shape, &c)

	out := req.Out
	if out == "" {
		base := strings.TrimSuffix(req.In, filepath.Ext(req.In))
		if req.Iconset {
			out = base + ".iconset"
		} else {
			out = base + ".icns"
		}
	}
	if !req.Force {
		if st, err := files.Stat(out); err == nil && st.Exists {
			c.Error(problem.Failed, msg("WAP-OUTPUT-EXISTS", fmt.Sprintf("%s exists already.", out)))
		}
	}
	if !ok {
		return ConvertedIcon{}, c.Err()
	}

	var nodes []Node
	if req.Iconset {
		bySlot := map[string]icns.Slot{}
		for _, s := range icns.Slots() {
			bySlot[s.Type] = s
		}
		for _, e := range entries {
			s, known := bySlot[e.Type]
			if !known {
				continue
			}
			if e.Size() == 0 {
				c.Error(problem.Failed, msg("WAP-ICONSET-INVALID", fmt.Sprintf("The %s image of %s is not a PNG and cannot go into an icon set.", e.Type, req.In)))
				continue
			}
			nodes = append(nodes, Node{Path: s.Iconset, Mode: 0o644, Data: e.Data})
		}
	}
	if err := c.Err(); err != nil {
		return ConvertedIcon{}, err
	}

	res := ConvertedIcon{Path: out, Sizes: iconSizes(entries), Iconset: req.Iconset, DryRun: req.DryRun, Report: Report{Notes: c.Notes()}}
	if req.DryRun {
		return res, nil
	}
	opts := WriteOptions{Force: req.Force, ModTime: req.ModTime}
	var err error
	if req.Iconset {
		err = files.WriteTree(out, normalizeTree(nodes), opts)
	} else {
		err = files.WriteFile(out, icns.Encode(entries), 0o644, opts)
	}
	if err != nil {
		return ConvertedIcon{}, writeFailed(out, err)
	}
	return res, nil
}

// MergeRequest asks for programs to be merged into a universal one.
type MergeRequest struct {
	Inputs []Binary
	Out    string
	Force  bool
	DryRun bool
}

// Merged describes the universal program.
type Merged struct {
	Path   string   `json:"path"`
	Archs  []string `json:"archs"`
	Size   int      `json:"size"`
	DryRun bool     `json:"dryRun"`
	Report
}

func mergeUniversal(files Files, req MergeRequest) (Merged, error) {
	var c problem.Collector
	if len(req.Inputs) < 2 {
		return Merged{}, problem.Misuse(msg("WAP-UNIVERSAL-INPUTS", fmt.Sprintf("%d program(s) were given.", len(req.Inputs))))
	}
	var progs []program
	for _, in := range req.Inputs {
		if p, ok := loadProgram(files, in, "program", &c); ok {
			progs = append(progs, p)
		}
	}
	out := req.Out
	if out == "" {
		first := req.Inputs[0].Path
		out = filepath.Join(filepath.Dir(first), trimArchSuffix(filepath.Base(first))+"-universal")
	}
	if !req.Force {
		if st, err := files.Stat(out); err == nil && st.Exists {
			c.Error(problem.Failed, msg("WAP-OUTPUT-EXISTS", fmt.Sprintf("%s exists already.", out)))
		}
	}
	if err := c.Err(); err != nil {
		return Merged{}, err
	}
	data, archs, _, ok := assemble(progs, &c)
	if !ok {
		return Merged{}, c.Err()
	}
	for _, p := range progs {
		programNotes(p, &c)
	}
	res := Merged{Path: out, Archs: archs, Size: len(data), DryRun: req.DryRun, Report: Report{Notes: c.Notes()}}
	if req.DryRun {
		return res, nil
	}
	if err := files.WriteFile(out, data, 0o755, WriteOptions{Force: req.Force}); err != nil {
		return Merged{}, writeFailed(out, err)
	}
	return res, nil
}
