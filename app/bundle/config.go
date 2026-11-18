package bundle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/confdoc"
	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/foundation/suggest"
)

// DefaultConfigFile is read when no configuration file is named.
const DefaultConfigFile = "wdyapppack.json"

// Placeholder marks a value init could not choose.
const Placeholder = "CHANGE-ME"

// ConfigFields describes every field of the configuration file this context
// reads. The help, the schema and the starter configuration are made from it.
func ConfigFields() []confdoc.Field {
	return []confdoc.Field{
		{Path: "name", Kind: confdoc.String, Required: true, Flag: "name", Example: `"My Tool"`,
			Description: "The name of the application. It is shown in the Finder and the menu bar (CFBundleName), and the bundle is written as <name>.app. At most 15 characters are shown in the menu bar."},
		{Path: "displayName", Kind: confdoc.String, Flag: "display-name", Default: "the name",
			Description: "A longer name for the Finder and the Dock (CFBundleDisplayName)."},
		{Path: "identifier", Kind: confdoc.String, Required: true, Flag: "identifier", Example: `"com.example.mytool"`,
			Description: "The bundle identifier in reverse domain notation (CFBundleIdentifier): letters, digits, dots and hyphens, at least one dot. It must stay the same across versions, because macOS attaches preferences and permissions to it."},
		{Path: "version", Kind: confdoc.String, Required: true, Flag: "version", Example: `"1.2.0"`,
			Description: "The version users see (CFBundleShortVersionString): one to three numbers separated by dots. May be taken from the environment: \"${VERSION}\"."},
		{Path: "buildNumber", Kind: confdoc.String, Flag: "build-number", Default: "the version",
			Description: "The build number (CFBundleVersion): one to three numbers separated by dots, increasing with every build that is handed out."},
		{Path: "binaries", Kind: confdoc.Objects, Required: true, Flag: "binary",
			Description: "The program to package. Give one program, or one per architecture: several are merged into one universal program."},
		{Path: "binaries.arch", Kind: confdoc.String, Enum: []string{"arm64", "amd64", "x86_64"},
			Description: "The architecture the program is built for. Optional: it is read from the program and only checked when given."},
		{Path: "binaries.path", Kind: confdoc.String, Required: true, Example: `"dist/mytool-darwin-arm64"`,
			Description: "Path to the program, relative to the configuration file."},
		{Path: "executableName", Kind: confdoc.String, Flag: "executable-name", Default: "the file name of the first program, without an architecture suffix",
			Description: "The file name of the program inside Contents/MacOS (CFBundleExecutable)."},
		{Path: "icon", Kind: confdoc.Object, Flag: "icon",
			Description: "The icon of the application."},
		{Path: "icon.path", Kind: confdoc.String, Required: true, Example: `"assets/icon.png"`,
			Description: "A PNG or JPEG image (square, ideally 1024×1024 pixels, transparent background), an .icns file or an .iconset directory, relative to the configuration file."},
		{Path: "icon.shape", Kind: confdoc.String, Enum: []string{"none", "squircle"}, Default: "none", Flag: "icon-shape",
			Description: "squircle cuts an image into the rounded square of macOS 26 with the margin of Apple's grid; none uses the image as it is."},
		{Path: "minimumMacOS", Kind: confdoc.String, Flag: "min-macos", Default: "the version the program was built for, else 11.0", Example: `"12.0"`,
			Description: "The oldest macOS the application runs on (LSMinimumSystemVersion). It may not be lower than the program's own."},
		{Path: "category", Kind: confdoc.String, Flag: "category", Example: `"developer-tools"`,
			Description: "The application category (LSApplicationCategoryType), as listed by: wdyapppack help categories. The short form is enough."},
		{Path: "copyright", Kind: confdoc.String, Flag: "copyright", Example: `"© 2026 Example GmbH"`,
			Description: "The copyright notice (NSHumanReadableCopyright), shown in the Finder's information window."},
		{Path: "agent", Kind: confdoc.Boolean, Flag: "agent", Default: "false",
			Description: "true hides the application from the Dock and the application switcher (LSUIElement), for menu bar tools."},
		{Path: "backgroundOnly", Kind: confdoc.Boolean, Flag: "background-only", Default: "false",
			Description: "true marks a program without any user interface (LSBackgroundOnly)."},
		{Path: "resources", Kind: confdoc.Objects, Flag: "resource",
			Description: "Files and directories copied into Contents/Resources."},
		{Path: "resources.from", Kind: confdoc.String, Required: true,
			Description: "The file or directory to copy, relative to the configuration file."},
		{Path: "resources.to", Kind: confdoc.String, Default: "the name of from",
			Description: "Where to put it below Contents/Resources."},
		{Path: "frameworks", Kind: confdoc.Strings, Flag: "framework",
			Description: "Libraries (.dylib) and frameworks (.framework) copied into Contents/Frameworks."},
		{Path: "helpers", Kind: confdoc.Strings, Flag: "helper",
			Description: "Further programs copied into Contents/Helpers. They are checked like the main program."},
		{Path: "urlSchemes", Kind: confdoc.Strings, Flag: "url-scheme", Example: `["mytool"]`,
			Description: "URL schemes the application opens (CFBundleURLTypes), without ://."},
		{Path: "usageDescriptions", Kind: confdoc.StringMap, Flag: "usage", Example: `{"camera": "Scans QR codes."}`,
			Description: "Why the application needs access, shown by macOS when it asks the user: kinds are listed by wdyapppack help info-plist."},
		{Path: "environment", Kind: confdoc.StringMap, Flag: "env", Example: `{"MYTOOL_MODE": "app"}`,
			Description: "Environment variables macOS sets when it starts the application (LSEnvironment)."},
		{Path: "plist", Kind: confdoc.AnyMap, Flag: "plist", Example: `{"NSSupportsAutomaticGraphicsSwitching": true}`,
			Description: "Further Info.plist keys with their values. Keys wdyapppack manages are refused; they have options of their own."},
		{Path: "plistFile", Kind: confdoc.String, Flag: "plist-file",
			Description: "An XML property list whose keys are merged into Info.plist, relative to the configuration file."},
		{Path: "out", Kind: confdoc.String, Flag: "out", Default: "dist",
			Description: "The directory the bundle is written to, relative to the configuration file."},
	}
}

var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// decodeConfig reads the configuration file strictly: a syntax error names its
// line, an unknown field is refused with the closest known one.
func decodeConfig(path string, data []byte, extraSections []string) (Config, error) {
	var c Config
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	dec.UseNumber()
	err := dec.Decode(&c)
	if err == nil && dec.More() {
		err = errors.New("there is more than one JSON value in the file")
	}
	if err == nil {
		return c, nil
	}

	var syntax *json.SyntaxError
	var typ *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntax):
		return c, problem.Misuse(msg("WAP-CONFIG-SYNTAX",
			fmt.Sprintf("%s, line %d: %s.", path, lineOf(data, syntax.Offset), syntax.Error())))
	case errors.As(err, &typ):
		return c, problem.Misuse(msg("WAP-CONFIG-TYPE",
			fmt.Sprintf("%s, line %d: the field %s holds %s, but must hold %s.", path, lineOf(data, typ.Offset), typ.Field, typ.Value, typ.Type)))
	}
	if name, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		name = strings.Trim(name, `"`)
		var known []string
		for _, f := range ConfigFields() {
			known = append(known, f.Path[strings.LastIndex(f.Path, ".")+1:])
		}
		known = append(known, extraSections...)
		how := "Correct or remove the field. Every field is listed by: wdyapppack help config"
		if s, ok := suggest.Closest(name, known); ok {
			how = fmt.Sprintf("Did you mean %q? %s", s, how)
		}
		return c, problem.Misuse(msgHow("WAP-CONFIG-FIELD-UNKNOWN", fmt.Sprintf("%s contains the unknown field %q.", path, name), how))
	}
	return c, problem.Misuse(msg("WAP-CONFIG-SYNTAX", fmt.Sprintf("%s: %v.", path, err)))
}

func lineOf(data []byte, offset int64) int {
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	return bytes.Count(data[:offset], []byte("\n")) + 1
}

// expand replaces ${NAME} in every text value of the configuration.
func expand(c *Config, lookup func(string) (string, bool)) error {
	var errs problem.Collector
	seen := map[string]bool{}
	sub := func(field string, s *string) {
		*s = envRef.ReplaceAllStringFunc(*s, func(ref string) string {
			name := envRef.FindStringSubmatch(ref)[1]
			v, ok := lookup(name)
			if !ok && !seen[name] {
				seen[name] = true
				errs.Error(problem.Usage, msg("WAP-CONFIG-ENV-UNSET",
					fmt.Sprintf("The field %s refers to ${%s}, which is not set.", field, name)))
			}
			return v
		})
	}
	for _, f := range []struct {
		name string
		p    *string
	}{
		{"name", &c.Name}, {"displayName", &c.DisplayName}, {"identifier", &c.Identifier},
		{"version", &c.Version}, {"buildNumber", &c.BuildNumber}, {"executableName", &c.ExecutableName},
		{"minimumMacOS", &c.MinimumMacOS}, {"category", &c.Category}, {"copyright", &c.Copyright},
		{"plistFile", &c.PlistFile}, {"out", &c.Out},
	} {
		sub(f.name, f.p)
	}
	for i := range c.Binaries {
		sub("binaries.path", &c.Binaries[i].Path)
		sub("binaries.arch", &c.Binaries[i].Arch)
	}
	if c.Icon != nil {
		sub("icon.path", &c.Icon.Path)
	}
	for i := range c.Resources {
		sub("resources.from", &c.Resources[i].From)
		sub("resources.to", &c.Resources[i].To)
	}
	for _, list := range []struct {
		name string
		s    []string
	}{{"frameworks", c.Frameworks}, {"helpers", c.Helpers}, {"urlSchemes", c.URLSchemes}} {
		for i := range list.s {
			sub(list.name, &list.s[i])
		}
	}
	for _, m := range []struct {
		name string
		m    map[string]string
	}{{"usageDescriptions", c.UsageDescriptions}, {"environment", c.Environment}} {
		for k, v := range m.m {
			sub(m.name+"."+k, &v)
			m.m[k] = v
		}
	}
	for k, v := range c.Plist {
		c.Plist[k] = expandAny("plist."+k, v, sub)
	}
	return errs.Err()
}

func expandAny(field string, v any, sub func(string, *string)) any {
	switch x := v.(type) {
	case string:
		sub(field, &x)
		return x
	case []any:
		for i := range x {
			x[i] = expandAny(field, x[i], sub)
		}
		return x
	case map[string]any:
		for k := range x {
			x[k] = expandAny(field+"."+k, x[k], sub)
		}
		return x
	}
	return v
}

// resolve makes the paths of the configuration relative to its directory.
func resolve(c *Config, dir string) {
	join := func(p *string) {
		if *p != "" && !filepath.IsAbs(*p) {
			*p = filepath.Join(dir, *p)
		}
	}
	for i := range c.Binaries {
		join(&c.Binaries[i].Path)
	}
	if c.Icon != nil {
		join(&c.Icon.Path)
	}
	for i := range c.Resources {
		join(&c.Resources[i].From)
	}
	for i := range c.Frameworks {
		join(&c.Frameworks[i])
	}
	for i := range c.Helpers {
		join(&c.Helpers[i])
	}
	join(&c.PlistFile)
	if c.Out != "" {
		join(&c.Out)
	}
}
