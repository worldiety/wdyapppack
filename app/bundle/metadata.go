package bundle

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/foundation/suggest"
)

var (
	identifierPattern = regexp.MustCompile(`^[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+$`)
	versionPattern    = regexp.MustCompile(`^[0-9]+(\.[0-9]+){0,2}$`)
	schemePattern     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*$`)
)

// Categories returns the application categories Apple defines, by their
// short name.
func Categories() []string {
	return []string{
		"business", "developer-tools", "education", "entertainment", "finance",
		"games", "graphics-design", "healthcare-fitness", "lifestyle", "medical",
		"music", "news", "photography", "productivity", "reference",
		"social-networking", "sports", "travel", "utilities", "video", "weather",
		"action-games", "adventure-games", "arcade-games", "board-games",
		"card-games", "casino-games", "dice-games", "educational-games",
		"family-games", "kids-games", "music-games", "puzzle-games",
		"racing-games", "role-playing-games", "simulation-games",
		"sports-games", "strategy-games", "trivia-games", "word-games",
	}
}

const categoryPrefix = "public.app-category."

// category resolves a category given in full or by its short name.
func category(s string) (string, problem.Message, bool) {
	short := strings.TrimPrefix(strings.ToLower(s), categoryPrefix)
	if slices.Contains(Categories(), short) {
		return categoryPrefix + short, problem.Message{}, true
	}
	how := "Use one of the values listed by: wdyapppack help categories"
	if sug, ok := suggest.Closest(short, Categories()); ok {
		how = fmt.Sprintf("Did you mean %q? %s", sug, how)
	}
	return "", msgHow("WAP-CATEGORY-UNKNOWN", fmt.Sprintf("%q is not an application category.", s), how), false
}

// UsageKinds maps the kinds of access a usage description can be given for to
// the Info.plist key macOS reads.
func UsageKinds() map[string]string {
	return map[string]string{
		"apple-events":          "NSAppleEventsUsageDescription",
		"bluetooth":             "NSBluetoothAlwaysUsageDescription",
		"calendars":             "NSCalendarsUsageDescription",
		"camera":                "NSCameraUsageDescription",
		"contacts":              "NSContactsUsageDescription",
		"desktop":               "NSDesktopFolderUsageDescription",
		"documents":             "NSDocumentsFolderUsageDescription",
		"downloads":             "NSDownloadsFolderUsageDescription",
		"local-network":         "NSLocalNetworkUsageDescription",
		"location":              "NSLocationUsageDescription",
		"microphone":            "NSMicrophoneUsageDescription",
		"network-volumes":       "NSNetworkVolumesUsageDescription",
		"photos":                "NSPhotoLibraryUsageDescription",
		"reminders":             "NSRemindersUsageDescription",
		"removable-volumes":     "NSRemovableVolumesUsageDescription",
		"speech-recognition":    "NSSpeechRecognitionUsageDescription",
		"system-administration": "NSSystemAdministrationUsageDescription",
	}
}

// ProtectedKeys are the Info.plist keys wdyapppack writes itself, with the
// option that sets each, or "" when it follows from the bundle.
func ProtectedKeys() map[string]string {
	return map[string]string{
		"CFBundleExecutable":            "--executable-name",
		"CFBundleIdentifier":            "--identifier",
		"CFBundleName":                  "--name",
		"CFBundleDisplayName":           "--display-name",
		"CFBundleShortVersionString":    "--version",
		"CFBundleVersion":               "--build-number",
		"CFBundlePackageType":           "",
		"CFBundleInfoDictionaryVersion": "",
		"CFBundleIconFile":              "--icon",
		"CFBundleSupportedPlatforms":    "",
		"LSMinimumSystemVersion":        "--min-macos",
		"LSApplicationCategoryType":     "--category",
		"NSHumanReadableCopyright":      "--copyright",
		"LSUIElement":                   "--agent",
		"LSBackgroundOnly":              "--background-only",
		"LSEnvironment":                 "--env",
		"CFBundleURLTypes":              "--url-scheme",
	}
}

// compareVersions compares two dotted versions number by number.
func compareVersions(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < max(len(pa), len(pb)); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if c := cmp.Compare(x, y); c != 0 {
			return c
		}
	}
	return 0
}

// metadata is what the Info.plist is made of, after validation.
type metadata struct {
	name, displayName, identifier, version, buildNumber string
	executable, minOS, category, copyright              string
	icon                                                bool
	agent, backgroundOnly                               bool
	urlSchemes                                          []string
	usage, environment                                  map[string]string
	custom                                              map[string]any
}

// infoPlist assembles the Info.plist dictionary.
func (m metadata) infoPlist() map[string]any {
	p := map[string]any{}
	maps.Copy(p, m.custom)
	p["CFBundleDevelopmentRegion"] = "en"
	p["CFBundleDisplayName"] = cmp.Or(m.displayName, m.name)
	p["CFBundleExecutable"] = m.executable
	p["CFBundleIdentifier"] = m.identifier
	p["CFBundleInfoDictionaryVersion"] = "6.0"
	p["CFBundleName"] = m.name
	p["CFBundlePackageType"] = "APPL"
	p["CFBundleShortVersionString"] = m.version
	p["CFBundleSupportedPlatforms"] = []any{"MacOSX"}
	p["CFBundleVersion"] = cmp.Or(m.buildNumber, m.version)
	p["LSMinimumSystemVersion"] = m.minOS
	if _, ok := p["NSHighResolutionCapable"]; !ok {
		p["NSHighResolutionCapable"] = true
	}
	if m.icon {
		p["CFBundleIconFile"] = "AppIcon"
	}
	if m.category != "" {
		p["LSApplicationCategoryType"] = m.category
	}
	if m.copyright != "" {
		p["NSHumanReadableCopyright"] = m.copyright
	}
	if m.agent {
		p["LSUIElement"] = true
	}
	if m.backgroundOnly {
		p["LSBackgroundOnly"] = true
	}
	if len(m.environment) > 0 {
		env := map[string]any{}
		for k, v := range m.environment {
			env[k] = v
		}
		p["LSEnvironment"] = env
	}
	if len(m.urlSchemes) > 0 {
		schemes := make([]any, len(m.urlSchemes))
		for i, s := range m.urlSchemes {
			schemes[i] = s
		}
		p["CFBundleURLTypes"] = []any{map[string]any{
			"CFBundleURLName":    m.identifier,
			"CFBundleURLSchemes": schemes,
		}}
	}
	kinds := UsageKinds()
	for kind, text := range m.usage {
		p[kinds[kind]] = text
	}
	return p
}

// normalizePlist converts values decoded from JSON into property list values:
// whole numbers become integers, other numbers reals. Null has no equivalent.
func normalizePlist(field string, v any) (any, error) {
	switch x := v.(type) {
	case json.Number:
		if i, err := x.Int64(); err == nil && !strings.ContainsAny(string(x), ".eE") {
			return i, nil
		}
		f, err := x.Float64()
		return f, err
	case float64:
		if x == float64(int64(x)) {
			return int64(x), nil
		}
		return x, nil
	case int:
		return int64(x), nil
	case string, bool, int64, time.Time, []byte:
		return x, nil
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			n, err := normalizePlist(fmt.Sprintf("%s[%d]", field, i), e)
			if err != nil {
				return nil, err
			}
			out[i] = n
		}
		return out, nil
	case map[string]any:
		out := map[string]any{}
		for k, e := range x {
			n, err := normalizePlist(field+"."+k, e)
			if err != nil {
				return nil, err
			}
			out[k] = n
		}
		return out, nil
	case nil:
		return nil, fmt.Errorf("%s is null, which a property list cannot hold", field)
	}
	return nil, fmt.Errorf("%s holds a %T, which a property list cannot hold", field, v)
}

// trimArchSuffix removes the architecture a build script typically appends to
// a program's name: mytool-darwin-arm64 becomes mytool.
func trimArchSuffix(name string) string {
	lower := strings.ToLower(name)
	for _, suffix := range []string{
		"-darwin-arm64", "-darwin-amd64", "-darwin-x86_64", "-darwin-universal",
		"_darwin_arm64", "_darwin_amd64", "_darwin_x86_64", "_darwin_universal",
		"-macos-arm64", "-macos-amd64", "-macos-x86_64", "-macos-universal",
		"-arm64", "-amd64", "-x86_64", "-universal", "_arm64", "_amd64", "_x86_64", "_universal",
		"-darwin", "_darwin", "-macos",
	} {
		if strings.HasSuffix(lower, suffix) && len(name) > len(suffix) {
			return name[:len(name)-len(suffix)]
		}
	}
	return name
}
