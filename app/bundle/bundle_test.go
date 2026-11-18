package bundle

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/pkg/icns"
	"github.com/worldiety/wdyapppack/pkg/macho"
	"github.com/worldiety/wdyapppack/pkg/macho/machotest"
	"github.com/worldiety/wdyapppack/pkg/plist"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/binary"
	rbundle "github.com/worldiety/wdyapppack/requirements/fun/bundle"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/fun/icon"
)

const app = "dist/My Tool.app"

func TestBuildWritesTheLayoutMacOSExpects(t *testing.T) {
	m := workspace(t)
	res, err := buildWith(m, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	if res.Path != app {
		t.Fatalf("written to %s", res.Path)
	}
	for _, p := range []string{"Contents/Info.plist", "Contents/PkgInfo", "Contents/MacOS/mytool", "Contents/Resources/AppIcon.icns"} {
		if st, _ := m.Stat(filepath.Join(app, p)); !st.Exists {
			t.Errorf("%s is missing", p)
		}
	}
	exe := m.files[filepath.Join(app, "Contents/MacOS/mytool")]
	if exe.mode&0o111 == 0 {
		t.Error("the program is not executable")
	}
	if pkg, _ := m.Read(filepath.Join(app, "Contents/PkgInfo")); string(pkg) != "APPL????" {
		t.Errorf("PkgInfo %q", pkg)
	}
	spec.Verified(t, rbundle.RBundleLayout)
}

func TestInfoPlistHasEveryKeyAnApplicationNeeds(t *testing.T) {
	m := workspace(t)
	cfg := validConfig()
	cfg.Copyright = "© 2026 Example"
	cfg.Agent = true
	cfg.BackgroundOnly = true
	cfg.Environment = map[string]string{"MODE": "app"}
	cfg.URLSchemes = []string{"mytool"}
	cfg.UsageDescriptions = map[string]string{"camera": "Scans codes."}
	if _, err := buildWith(m, cfg); err != nil {
		t.Fatal(err)
	}
	p := infoOf(t, m, app)
	want := map[string]any{
		"CFBundleName":                  "My Tool",
		"CFBundleDisplayName":           "My Tool",
		"CFBundleIdentifier":            "com.example.mytool",
		"CFBundleExecutable":            "mytool",
		"CFBundlePackageType":           "APPL",
		"CFBundleInfoDictionaryVersion": "6.0",
		"CFBundleShortVersionString":    "1.2.0",
		"CFBundleVersion":               "1.2.0",
		"CFBundleIconFile":              "AppIcon",
		"CFBundleDevelopmentRegion":     "en",
		"LSMinimumSystemVersion":        "12.0",
		"NSHighResolutionCapable":       true,
		"NSHumanReadableCopyright":      "© 2026 Example",
		"LSUIElement":                   true,
		"LSBackgroundOnly":              true,
		"NSCameraUsageDescription":      "Scans codes.",
	}
	for k, v := range want {
		if p[k] != v {
			t.Errorf("%s = %v, want %v", k, p[k], v)
		}
	}
	if env := p["LSEnvironment"].(map[string]any); env["MODE"] != "app" {
		t.Errorf("LSEnvironment %v", env)
	}
	types := p["CFBundleURLTypes"].([]any)[0].(map[string]any)
	if types["CFBundleURLSchemes"].([]any)[0] != "mytool" {
		t.Errorf("URL types %v", types)
	}
	if plat := p["CFBundleSupportedPlatforms"].([]any); plat[0] != "MacOSX" {
		t.Errorf("platforms %v", plat)
	}
	spec.Verified(t, rbundle.RBundlePlistDefaults)
}

// The program is copied byte for byte and started directly: no launcher, no
// rewritten load paths, even when it loads a library from outside macOS.
func TestTheProgramIsPackagedAsItIs(t *testing.T) {
	m := workspace(t)
	prog := arm64(func(o *machotest.Options) { o.Libraries = append(o.Libraries, "/opt/homebrew/lib/libfoo.dylib") })
	m.put("dist/mytool-darwin-arm64", prog, 0o755)
	res, err := buildWith(m, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	got, _ := m.Read(filepath.Join(app, "Contents/MacOS/mytool"))
	if !bytes.Equal(got, prog) {
		t.Fatal("the program was changed")
	}
	if infoOf(t, m, app)["CFBundleExecutable"] != "mytool" {
		t.Fatal("the bundle does not start the program itself")
	}
	if !slices.Contains(noteCodes(res.Report), "WAP-DYLIB-EXTERNAL") {
		t.Fatalf("no note on the external library: %v", res.Notes)
	}
	spec.Verified(t, dec.RDecDirectLaunchOnly, dec.RDecNoLoadpathRewrite, binary.RBinaryDylibExternal)
}

func TestRequiredMetadataIsReportedTogether(t *testing.T) {
	m := workspace(t)
	_, err := buildWith(m, Config{Binaries: []Binary{{Path: "dist/mytool-darwin-arm64"}}})
	wantCodes(t, err, "WAP-NAME-MISSING", "WAP-IDENTIFIER-MISSING", "WAP-VERSION-MISSING")

	cfg := validConfig()
	cfg.Name = "a/b"
	cfg.Identifier = "mytool"
	cfg.Version = "v1.2"
	cfg.BuildNumber = "1.2.3.4"
	_, err = buildWith(m, cfg)
	wantCodes(t, err, "WAP-NAME-INVALID", "WAP-IDENTIFIER-INVALID", "WAP-VERSION-INVALID")
	if m.writes != 0 {
		t.Fatal("something was written despite the errors")
	}
	spec.Verified(t, rbundle.RBundleMetadataRequired)
}

func TestCategoryShortAndFull(t *testing.T) {
	for in, want := range map[string]string{
		"developer-tools":                  "public.app-category.developer-tools",
		"public.app-category.puzzle-games": "public.app-category.puzzle-games",
	} {
		m := workspace(t)
		cfg := validConfig()
		cfg.Category = in
		if _, err := buildWith(m, cfg); err != nil {
			t.Fatal(err)
		}
		if got := infoOf(t, m, app)["LSApplicationCategoryType"]; got != want {
			t.Errorf("%s → %v", in, got)
		}
	}
	cfg := validConfig()
	cfg.Category = "developer-tool"
	_, err := buildWith(workspace(t), cfg)
	wantCodes(t, err, "WAP-CATEGORY-UNKNOWN")
	if !strings.Contains(err.Error(), "") || !strings.Contains(messagesOf(err), `"developer-tools"`) {
		t.Errorf("no suggestion: %v", messagesOf(err))
	}
	spec.Verified(t, rbundle.RBundleCategory)
}

func TestAdditionalMetadataWithTypes(t *testing.T) {
	m := workspace(t)
	extra, _ := plist.Encode(map[string]any{"FromFile": "yes", "Shared": "file"})
	m.put("extra.plist", extra, 0o644)
	cfg := validConfig()
	cfg.PlistFile = "extra.plist"
	cfg.Plist = map[string]any{
		"Shared":                  "config",
		"AnInt":                   json.Number("3"),
		"AReal":                   json.Number("1.5"),
		"ABool":                   false,
		"AnArray":                 []any{"a", "b"},
		"NSHighResolutionCapable": false,
	}
	if _, err := buildWith(m, cfg); err != nil {
		t.Fatal(err)
	}
	p := infoOf(t, m, app)
	if p["FromFile"] != "yes" || p["Shared"] != "config" || p["AnInt"] != int64(3) || p["AReal"] != 1.5 || p["ABool"] != false || len(p["AnArray"].([]any)) != 2 {
		t.Fatalf("%v", p)
	}
	if p["NSHighResolutionCapable"] != false {
		t.Fatal("a default could not be overridden")
	}
	spec.Verified(t, rbundle.RBundlePlistCustom)
}

func TestProtectedKeysAreRefused(t *testing.T) {
	m := workspace(t)
	extra, _ := plist.Encode(map[string]any{"CFBundlePackageType": "BNDL"})
	m.put("extra.plist", extra, 0o644)
	cfg := validConfig()
	cfg.Plist = map[string]any{"CFBundleExecutable": "other"}
	cfg.PlistFile = "extra.plist"
	_, err := buildWith(m, cfg)
	wantCodes(t, err, "WAP-PLIST-PROTECTED")
	if n := strings.Count(messagesOf(err), "WAP-PLIST-PROTECTED"); n != 2 {
		t.Fatalf("%d protected key messages: %s", n, messagesOf(err))
	}
	spec.Verified(t, rbundle.RBundlePlistProtected)
}

func TestResourcesFrameworksAndHelpers(t *testing.T) {
	m := workspace(t)
	m.put("libs/libfoo.dylib", machotest.Program(machotest.Options{Arch: "arm64", Dylib: true}), 0o644)
	m.put("dist/helper", arm64(func(o *machotest.Options) { o.Signed = false }), 0o755)
	m.put("dist/notaprogram", machotest.ELF(), 0o755)
	cfg := validConfig()
	cfg.Resources = []Resource{{From: "assets/data"}, {From: "assets/data/a.txt", To: "x/renamed.txt"}}
	cfg.Frameworks = []string{"libs/libfoo.dylib"}
	cfg.Helpers = []string{"dist/helper"}
	res, err := buildWith(m, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"Contents/Resources/data/a.txt", "Contents/Resources/data/sub/b.txt", "Contents/Resources/x/renamed.txt", "Contents/Frameworks/libfoo.dylib", "Contents/Helpers/helper"} {
		if st, _ := m.Stat(filepath.Join(app, p)); !st.Exists {
			t.Errorf("%s missing", p)
		}
	}
	if !slices.Contains(noteCodes(res.Report), "WAP-ARM64-UNSIGNED") {
		t.Error("the unsigned helper was not checked like the program")
	}

	cfg.Helpers = []string{"dist/notaprogram"}
	cfg.Resources = []Resource{{From: "assets/data", To: "../escape"}}
	_, err = buildWith(workspace(t), cfg)
	wantCodes(t, err, "WAP-NAME-INVALID")
	m2 := workspace(t)
	m2.put("libs/libfoo.dylib", []byte("x"), 0o644)
	m2.put("dist/notaprogram", machotest.ELF(), 0o755)
	cfg.Resources = nil
	_, err = buildWith(m2, cfg)
	wantCodes(t, err, "WAP-BINARY-NOT-MACHO")
	spec.Verified(t, rbundle.RBundlePayload)
}

func TestSameInputSameBundle(t *testing.T) {
	when := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	a, b := workspace(t), workspace(t)
	cfg := validConfig()
	cfg.Plist = map[string]any{"Z": "z", "A": "a", "M": map[string]any{"y": "1", "b": "2"}}
	if _, err := buildWith(a, cfg, func(r *BuildRequest) { r.ModTime = when }); err != nil {
		t.Fatal(err)
	}
	if _, err := buildWith(b, cfg, func(r *BuildRequest) { r.ModTime = when }); err != nil {
		t.Fatal(err)
	}
	for p, f := range a.files {
		if !bytes.Equal(f.data, b.files[p].data) || f.mode != b.files[p].mode {
			t.Fatalf("%s differs between two builds", p)
		}
	}
	spec.Verified(t, rbundle.RBundleReproducible)
}

func TestTheBundleIsReadBack(t *testing.T) {
	m := workspace(t)
	m.corrupt = true
	_, err := buildWith(m, validConfig())
	wantCodes(t, err, "WAP-SELF-CHECK")
	spec.Verified(t, rbundle.RBundleSelfCheck)
}

func TestNotesFailOnlyInStrictMode(t *testing.T) {
	m := workspace(t)
	m.put("assets/icon.png", pngImage(t, 600, 600, true), 0o644)
	res, err := buildWith(m, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(noteCodes(res.Report), "WAP-ICON-SMALL") {
		t.Fatalf("notes %v", res.Notes)
	}
	_, err = buildWith(workspace(t), validConfig(), func(r *BuildRequest) { r.Strict = true })
	if err != nil {
		t.Fatalf("a build without notes failed in strict mode: %v", err)
	}
	m = workspace(t)
	m.put("assets/icon.png", pngImage(t, 600, 600, true), 0o644)
	_, err = buildWith(m, validConfig(), func(r *BuildRequest) { r.Strict = true })
	wantCodes(t, err, "WAP-ICON-SMALL")
	spec.Verified(t, rbundle.RBundleStrict)
}

func TestDryRunWritesNothing(t *testing.T) {
	m := workspace(t)
	res, err := buildWith(m, validConfig(), func(r *BuildRequest) { r.DryRun = true })
	if err != nil {
		t.Fatal(err)
	}
	if m.writes != 0 {
		t.Fatal("a dry run wrote")
	}
	if !res.DryRun || !slices.Contains(res.Files, "My Tool.app/Contents/MacOS/mytool") || !strings.Contains(res.InfoPlist, "CFBundleIdentifier") {
		t.Fatalf("a dry run does not show what would be written: %+v", res)
	}
	spec.Verified(t, guide.RGuideDryRun)
}

func TestNothingIsOverwrittenUnasked(t *testing.T) {
	m := workspace(t)
	if _, err := buildWith(m, validConfig()); err != nil {
		t.Fatal(err)
	}
	_, err := buildWith(m, validConfig())
	wantCodes(t, err, "WAP-OUTPUT-EXISTS")
	if _, err := buildWith(m, validConfig(), func(r *BuildRequest) { r.Force = true }); err != nil {
		t.Fatalf("--force did not replace: %v", err)
	}
	m.put("wdyapppack.json", []byte("{}"), 0o644)
	_, err = NewWriteConfig(m)(auth.System(), WriteConfigRequest{})
	wantCodes(t, err, "WAP-OUTPUT-EXISTS")
	spec.Verified(t, guide.RGuideNoOverwrite)
}

func TestOnlyMacOSProgramsArePackaged(t *testing.T) {
	for name, data := range map[string][]byte{
		"elf":    machotest.ELF(),
		"pe":     machotest.PE(),
		"script": []byte("#!/bin/sh\n"),
		"text":   []byte("hello"),
	} {
		m := workspace(t)
		m.put("dist/mytool-darwin-arm64", data, 0o755)
		_, err := buildWith(m, validConfig())
		wantCodes(t, err, "WAP-BINARY-NOT-MACHO")
		if name == "elf" && !strings.Contains(messagesOf(err), "GOOS=darwin") {
			t.Errorf("no explanation how to build for macOS: %s", messagesOf(err))
		}
	}
	m := workspace(t)
	m.put("dist/mytool-darwin-arm64", machotest.Program(machotest.Options{Arch: "arm64", Dylib: true}), 0o755)
	_, err := buildWith(m, validConfig())
	wantCodes(t, err, "WAP-BINARY-NOT-EXECUTABLE")
	spec.Verified(t, binary.RBinaryMachoOnly)
}

func TestArchitecturesAreRecognisedAndChecked(t *testing.T) {
	m := workspace(t)
	cfg := validConfig()
	cfg.Binaries = []Binary{{Arch: "amd64", Path: "dist/mytool-darwin-arm64"}}
	_, err := buildWith(m, cfg)
	wantCodes(t, err, "WAP-ARCH-MISMATCH")

	cfg.Binaries = []Binary{{Arch: "ppc", Path: "dist/mytool-darwin-arm64"}}
	_, err = buildWith(m, cfg)
	wantCodes(t, err, "WAP-ARCH-UNKNOWN")

	m.put("dist/other-arm64", arm64(), 0o755)
	cfg.Binaries = []Binary{{Path: "dist/mytool-darwin-arm64"}, {Path: "dist/other-arm64"}}
	_, err = buildWith(m, cfg)
	wantCodes(t, err, "WAP-ARCH-DUPLICATE")

	cfg.Binaries = []Binary{{Arch: "arm64", Path: "dist/mytool-darwin-arm64"}}
	res, err := buildWith(m, cfg)
	if err != nil || !slices.Equal(res.Archs, []string{"arm64"}) {
		t.Fatalf("%v %v", res.Archs, err)
	}
	spec.Verified(t, binary.RBinaryArchDetect)
}

func TestMinimumMacOS(t *testing.T) {
	m := workspace(t)
	res, _ := buildWith(m, validConfig())
	if res.MinimumMacOS != "12.0" {
		t.Errorf("not taken from the program: %s", res.MinimumMacOS)
	}

	m = workspace(t)
	m.put("dist/mytool-darwin-arm64", arm64(func(o *machotest.Options) { o.MinOS = "" }), 0o755)
	res, _ = buildWith(m, validConfig())
	if res.MinimumMacOS != DefaultMinimumMacOS {
		t.Errorf("no default: %s", res.MinimumMacOS)
	}

	cfg := validConfig()
	cfg.MinimumMacOS = "11.0"
	_, err := buildWith(workspace(t), cfg)
	wantCodes(t, err, "WAP-MINOS-TOO-LOW")

	cfg.MinimumMacOS = "13.1"
	res, err = buildWith(workspace(t), cfg)
	if err != nil || res.MinimumMacOS != "13.1" {
		t.Errorf("a higher version was not accepted: %v", err)
	}
	spec.Verified(t, binary.RBinaryMinos)
}

func TestUnsignedAppleSiliconProgramIsNoted(t *testing.T) {
	m := workspace(t)
	m.put("dist/mytool-darwin-arm64", arm64(func(o *machotest.Options) { o.Signed = false }), 0o755)
	res, err := buildWith(m, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(noteCodes(res.Report), "WAP-ARM64-UNSIGNED") || !strings.Contains(res.Notes[0].How+res.Notes[len(res.Notes)-1].How, "sign") {
		t.Fatalf("notes %v", res.Notes)
	}
	spec.Verified(t, binary.RBinaryArm64Signed)
}

func TestUniversalProgramFromTwo(t *testing.T) {
	m := workspace(t)
	cfg := validConfig()
	cfg.Binaries = []Binary{{Arch: "arm64", Path: "dist/mytool-darwin-arm64"}, {Arch: "amd64", Path: "dist/mytool-darwin-amd64"}}
	res, err := buildWith(m, cfg)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := m.Read(filepath.Join(app, "Contents/MacOS/mytool"))
	info, err := macho.Inspect(data)
	if err != nil || info.Kind != macho.Universal || len(info.Slices) != 2 {
		t.Fatalf("%+v %v", info, err)
	}
	if res.MinimumMacOS != "12.0" {
		t.Errorf("the minimum is not the highest of the programs: %s", res.MinimumMacOS)
	}

	merged, err := NewMergeUniversal(m)(auth.System(), MergeRequest{Inputs: cfg.Binaries, Out: "dist/u"})
	if err != nil || !slices.Equal(merged.Archs, []string{"arm64", "x86_64"}) {
		t.Fatalf("%+v %v", merged, err)
	}
	if m.files["dist/u"].mode&0o111 == 0 {
		t.Error("the universal program is not executable")
	}
	_, err = NewMergeUniversal(m)(auth.System(), MergeRequest{Inputs: cfg.Binaries[:1]})
	wantCodes(t, err, "WAP-UNIVERSAL-INPUTS")
	spec.Verified(t, binary.RBinaryUniversal)
}

func TestIconFormats(t *testing.T) {
	m := workspace(t)
	m.put("icons/a.jpg", jpegImage(t, 1024), 0o644)
	entries, _ := icns.FromImage(mustDecode(t, pngImage(t, 1024, 1024, false)))
	m.put("icons/a.icns", icns.Encode(entries), 0o644)
	pngs, _ := icns.Render(mustDecode(t, pngImage(t, 1024, 1024, false)))
	for i, s := range icns.Slots() {
		m.put("icons/a.iconset/"+s.Iconset, pngs[i], 0o644)
	}
	m.put("icons/a.gif", []byte("GIF89a"), 0o644)

	for _, p := range []string{"assets/icon.png", "icons/a.jpg", "icons/a.icns", "icons/a.iconset"} {
		cfg := validConfig()
		cfg.Icon = &Icon{Path: p}
		res, err := buildWith(m, cfg, func(r *BuildRequest) { r.DryRun = true })
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if len(res.IconSizes) == 0 {
			t.Errorf("%s gave no icon", p)
		}
	}
	cfg := validConfig()
	cfg.Icon = &Icon{Path: "icons/a.gif"}
	_, err := buildWith(m, cfg)
	wantCodes(t, err, "WAP-ICON-FORMAT")
	spec.Verified(t, icon.RIconFormats)
}

func TestIconHasEverySize(t *testing.T) {
	m := workspace(t)
	res, err := buildWith(m, validConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.IconSizes, []int{16, 32, 64, 128, 256, 512, 1024}) {
		t.Fatalf("sizes %v", res.IconSizes)
	}
	data, _ := m.Read(filepath.Join(app, "Contents/Resources/AppIcon.icns"))
	if _, err := icns.Decode(data); err != nil {
		t.Fatal(err)
	}
	spec.Verified(t, icon.RIconSizes)
}

func TestIconMustBeSquareAndLarge(t *testing.T) {
	cases := map[string][]byte{
		"WAP-ICON-NOT-SQUARE": pngImage(t, 1024, 800, false),
		"WAP-ICON-TOO-SMALL":  pngImage(t, 256, 256, false),
	}
	for code, img := range cases {
		m := workspace(t)
		m.put("assets/icon.png", img, 0o644)
		_, err := buildWith(m, validConfig())
		wantCodes(t, err, code)
	}
	m := workspace(t)
	m.put("assets/icon.png", pngImage(t, 512, 512, false), 0o644)
	res, err := buildWith(m, validConfig())
	if err != nil || !slices.Contains(noteCodes(res.Report), "WAP-ICON-SMALL") {
		t.Fatalf("%v %v", res.Notes, err)
	}
	spec.Verified(t, icon.RIconSquare)
}

func TestIconShape(t *testing.T) {
	m := workspace(t)
	m.put("assets/icon.png", pngImage(t, 1024, 1024, true), 0o644)
	res, err := buildWith(m, validConfig(), func(r *BuildRequest) { r.DryRun = true })
	if err != nil || !slices.Contains(noteCodes(res.Report), "WAP-ICON-SHAPE") {
		t.Fatalf("no note on opaque corners: %v %v", res.Notes, err)
	}

	cfg := validConfig()
	cfg.Icon.Shape = ShapeSquircle
	res, err = buildWith(m, cfg)
	if err != nil || slices.Contains(noteCodes(res.Report), "WAP-ICON-SHAPE") {
		t.Fatalf("%v %v", res.Notes, err)
	}
	data, _ := m.Read(filepath.Join(app, "Contents/Resources/AppIcon.icns"))
	entries, _ := icns.Decode(data)
	big := mustDecode(t, entries[len(entries)-1].Data)
	if _, _, _, a := big.At(0, 0).RGBA(); a != 0 {
		t.Fatal("the shaped icon has an opaque corner")
	}

	cfg.Icon = &Icon{Path: "x.icns", Shape: ShapeSquircle}
	_, err = buildWith(workspace(t), cfg)
	wantCodes(t, err, "WAP-FILE-NOT-FOUND")
	spec.Verified(t, icon.RIconShape)
}

func TestConvertIconToIcnsAndIconset(t *testing.T) {
	m := workspace(t)
	res, err := NewConvertIcon(m)(auth.System(), ConvertIconRequest{In: "assets/icon.png"})
	if err != nil || res.Path != "assets/icon.icns" {
		t.Fatalf("%+v %v", res, err)
	}
	res, err = NewConvertIcon(m)(auth.System(), ConvertIconRequest{In: "assets/icon.png", Iconset: true})
	if err != nil || res.Path != "assets/icon.iconset" {
		t.Fatalf("%+v %v", res, err)
	}
	for _, s := range icns.Slots() {
		if st, _ := m.Stat("assets/icon.iconset/" + s.Iconset); !st.Exists {
			t.Errorf("%s missing", s.Iconset)
		}
	}
	spec.Verified(t, icon.RIconIconset)
}

func TestLoadConfigResolvesPathsAgainstTheFile(t *testing.T) {
	m := newMem()
	m.put("project/wdyapppack.json", []byte(`{
  "name": "My Tool",
  "identifier": "com.example.mytool",
  "version": "1.0",
  "binaries": [{"path": "dist/mytool"}],
  "icon": {"path": "icon.png"},
  "out": "build",
  "sign": {"identity": "-"}
}`), 0o644)
	load := NewLoadConfig(m)
	got, err := load(auth.System(), LoadConfigRequest{Path: "project/wdyapppack.json"})
	if err != nil {
		t.Fatal(err)
	}
	c := got.Config
	if c.Binaries[0].Path != "project/dist/mytool" || c.Icon.Path != "project/icon.png" || c.Out != "project/build" {
		t.Fatalf("paths not resolved: %+v", c)
	}
	if string(c.Sign) == "" {
		t.Fatal("the sign section was lost")
	}

	got, err = load(auth.System(), LoadConfigRequest{Optional: true})
	if err != nil || got.Found {
		t.Fatalf("an absent default is not optional: %v", err)
	}
	_, err = load(auth.System(), LoadConfigRequest{Path: "nope.json"})
	wantCodes(t, err, "WAP-CONFIG-MISSING")

	m.put("bad.json", []byte("{\n  \"name\": \"x\",\n}"), 0o644)
	_, err = load(auth.System(), LoadConfigRequest{Path: "bad.json"})
	wantCodes(t, err, "WAP-CONFIG-SYNTAX")
	if !strings.Contains(messagesOf(err), "line 3") {
		t.Errorf("no line: %s", messagesOf(err))
	}
	m.put("typed.json", []byte(`{"name": 3}`), 0o644)
	_, err = load(auth.System(), LoadConfigRequest{Path: "typed.json"})
	wantCodes(t, err, "WAP-CONFIG-TYPE")
	spec.Verified(t, config.RConfigFile, dec.RDecJsonConfig)
}

func TestUnknownFieldsAreRefused(t *testing.T) {
	m := newMem()
	m.put("c.json", []byte(`{"nmae": "x"}`), 0o644)
	_, err := NewLoadConfig(m)(auth.System(), LoadConfigRequest{Path: "c.json"})
	wantCodes(t, err, "WAP-CONFIG-FIELD-UNKNOWN")
	if !strings.Contains(messagesOf(err), `Did you mean "name"?`) {
		t.Errorf("no suggestion: %s", messagesOf(err))
	}
	m.put("c.json", []byte(`{"icon": {"pth": "x"}}`), 0o644)
	_, err = NewLoadConfig(m)(auth.System(), LoadConfigRequest{Path: "c.json"})
	wantCodes(t, err, "WAP-CONFIG-FIELD-UNKNOWN")
	spec.Verified(t, config.RConfigStrict)
}

func TestEnvironmentVariables(t *testing.T) {
	m := newMem()
	m.put("c.json", []byte(`{"version": "${VERSION}", "name": "${NAME} Tool", "plist": {"K": ["${VERSION}"]}, "copyright": "${MISSING}"}`), 0o644)
	env := map[string]string{"VERSION": "2.3.4", "NAME": "My"}
	lookup := func(k string) (string, bool) { v, ok := env[k]; return v, ok }
	_, err := NewLoadConfig(m)(auth.System(), LoadConfigRequest{Path: "c.json", Lookup: lookup})
	wantCodes(t, err, "WAP-CONFIG-ENV-UNSET")
	if !strings.Contains(messagesOf(err), "MISSING") {
		t.Error("the unset variable is not named")
	}
	env["MISSING"] = "©"
	got, err := NewLoadConfig(m)(auth.System(), LoadConfigRequest{Path: "c.json", Lookup: lookup})
	if err != nil {
		t.Fatal(err)
	}
	if got.Config.Version != "2.3.4" || got.Config.Name != "My Tool" || got.Config.Plist["K"].([]any)[0] != "2.3.4" {
		t.Fatalf("%+v", got.Config)
	}
	spec.Verified(t, config.RConfigEnv)
}

func TestStarterConfiguration(t *testing.T) {
	m := workspace(t)
	res, err := NewWriteConfig(m)(auth.System(), WriteConfigRequest{
		Binary: "dist/mytool-darwin-arm64",
		Extra:  map[string]json.RawMessage{"sign": json.RawMessage(`{"identity":"-"}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(res.Content), &doc); err != nil {
		t.Fatalf("%v:\n%s", err, res.Content)
	}
	for _, f := range ConfigFields() {
		if !strings.Contains(f.Path, ".") {
			if _, ok := doc[f.Path]; !ok {
				t.Errorf("the starter lacks %s", f.Path)
			}
		}
	}
	if doc["name"] != "mytool" || doc["minimumMacOS"] != "12.0" || doc["sign"] == nil {
		t.Fatalf("%s", res.Content)
	}
	if !strings.Contains(doc["identifier"].(string), Placeholder) {
		t.Fatal("the identifier is not a placeholder")
	}

	// The starter loads, and a build refuses its placeholder.
	loaded, err := NewLoadConfig(m)(auth.System(), LoadConfigRequest{Path: res.Path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = buildWith(m, loaded.Config)
	wantCodes(t, err, "WAP-CONFIG-PLACEHOLDER")
	spec.Verified(t, config.RConfigInit)
}

func TestEveryMessageCodeIsExplained(t *testing.T) {
	ex := explanations()
	for code, e := range ex {
		if e.summary == "" || e.why == "" || e.how == "" || e.see == "" {
			t.Errorf("%s is not fully explained", code)
		}
	}
	src := sources(t)
	for _, code := range quotedCodes(src) {
		if _, ok := ex[code]; !ok {
			t.Errorf("%s is used but not explained", code)
		}
	}
}
