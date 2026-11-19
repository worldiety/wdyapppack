package main

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/foundation/cli"
	"github.com/worldiety/wdyapppack/foundation/confdoc"
	"github.com/worldiety/wdyapppack/foundation/permission"
	"github.com/worldiety/wdyapppack/pkg/macho"
	"github.com/worldiety/wdyapppack/pkg/plist"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

func app() *cli.App { return newApp(func(string) (string, bool) { return "", false }, &bytes.Buffer{}) }

func invoke(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	if dir != "" {
		t.Chdir(dir)
	}
	var out, errOut bytes.Buffer
	code := app().Run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

// everything renders the whole help: the overview, every command and every
// topic.
func everything(t *testing.T) map[string]string {
	t.Helper()
	a := app()
	out := map[string]string{"overview": a.Overview()}
	for i := range a.Commands {
		out["command "+a.Commands[i].Name] = a.CommandHelp(&a.Commands[i])
	}
	for i := range a.Topics {
		out["topic "+a.Topics[i].Name] = a.TopicHelp(&a.Topics[i])
	}
	return out
}

// A newcomer finds everything from the program: every command and every topic
// is named in the overview, every command explains every option with an
// example that parses, every message code is documented and points to a
// topic that exists.
func TestTheProgramExplainsItself(t *testing.T) {
	a := app()
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	overview := a.Overview()
	for _, c := range a.Commands {
		if !strings.Contains(overview, c.Name) {
			t.Errorf("the overview does not name %s", c.Name)
		}
		for _, e := range c.Examples {
			if err := a.ParseExample(e.Command); err != nil {
				t.Errorf("example of %s does not parse: %s: %v", c.Name, e.Command, err)
			}
		}
	}
	for _, topic := range a.Topics {
		if !strings.Contains(overview, topic.Name) {
			t.Errorf("the overview does not name topic %s", topic.Name)
		}
	}
	for _, l := range strings.Split(overview, "\n") {
		if len([]rune(l)) > cli.Width {
			t.Errorf("overview line longer than %d: %q", cli.Width, l)
		}
	}
	seen := map[string]bool{}
	for _, d := range a.Errors {
		if seen[d.Code] {
			t.Errorf("%s documented twice", d.Code)
		}
		seen[d.Code] = true
	}
	// Every code in the source is documented.
	codes := regexp.MustCompile(`"(WAP-[A-Z0-9-]+)"`)
	_ = filepath.WalkDir("../..", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		data, _ := os.ReadFile(p)
		for _, m := range codes.FindAllStringSubmatch(string(data), -1) {
			if !seen[m[1]] {
				t.Errorf("%s uses %s, which help errors does not list", p, m[1])
			}
		}
		return nil
	})
	spec.Verified(t, nfr.RNfrSelfExplaining)
}

func TestHelpTopicsCoverTheBackground(t *testing.T) {
	want := []string{"bundle", "config", "info-plist", "icons", "binaries", "architectures", "signing",
		"notarization", "distribution", "categories", "recipes", "ci", "agents", "errors", "exit-codes"}
	for _, topic := range want {
		code, out, _ := invoke(t, "", "help", topic)
		if code != 0 || len(out) < 400 {
			t.Errorf("topic %s: exit %d, %d bytes", topic, code, len(out))
		}
	}
	spec.Verified(t, guide.RGuideHelpTopics)
}

func TestRequirementsAreListed(t *testing.T) {
	_, out, _ := invoke(t, "", "help", "requirements")
	reqs := spec.Requirements()
	if len(reqs) < 60 {
		t.Fatalf("only %d requirements are linked in", len(reqs))
	}
	for _, r := range reqs {
		if !strings.Contains(out, string(r.ID)) {
			t.Errorf("%s is not listed", r.ID)
		}
	}
	spec.Verified(t, guide.RGuideRequirements)
}

func TestVersion(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}} {
		code, out, _ := invoke(t, "", args...)
		if code != 0 || !strings.HasPrefix(out, "wdyapppack ") || !strings.Contains(out, "go1") {
			t.Fatalf("%v: %d %q", args, code, out)
		}
	}
	_, out, _ := invoke(t, "", "version", "--format", "json")
	var v struct {
		OK     bool    `json:"ok"`
		Result Version `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil || v.Result.Version == "" {
		t.Fatalf("%v %s", err, out)
	}
	spec.Verified(t, guide.RGuideVersion)
}

func TestSchemaDescribesTheStarterConfiguration(t *testing.T) {
	_, out, _ := invoke(t, "", "schema")
	var s struct {
		Properties map[string]struct {
			Description string `json:"description"`
		} `json:"properties"`
		AdditionalProperties bool `json:"additionalProperties"`
	}
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		t.Fatal(err)
	}
	if s.AdditionalProperties {
		t.Fatal("the schema accepts unknown fields")
	}
	dir := t.TempDir()
	if code, _, e := invoke(t, dir, "init"); code != 0 {
		t.Fatal(e)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "wdyapppack.json"))
	var starter map[string]any
	if err := json.Unmarshal(data, &starter); err != nil {
		t.Fatal(err)
	}
	for k := range starter {
		p, ok := s.Properties[k]
		if !ok || (p.Description == "" && k != "$schema") {
			t.Errorf("the schema does not describe %s", k)
		}
	}
	spec.Verified(t, config.RConfigSchema)
}

// wdyapppack depends on nothing but the standard library and speclink's
// specification package.
func TestStandardLibraryOnly(t *testing.T) {
	data, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	var deps []string
	inBlock := false
	for _, l := range strings.Split(string(data), "\n") {
		l = strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(l, "require ("):
			inBlock = true
		case inBlock && l == ")":
			inBlock = false
		case inBlock && l != "":
			deps = append(deps, strings.Fields(l)[0])
		case strings.HasPrefix(l, "require "):
			deps = append(deps, strings.Fields(l)[1])
		}
	}
	if !slices.Equal(deps, []string{"github.com/worldiety/speclink/spec"}) {
		t.Fatalf("dependencies: %v", deps)
	}
	spec.Verified(t, cst.RCstStdlibOnly)
}

// Every use case runs as a subject that holds every permission: what the tool
// may do is decided by the operating system user who runs it.
func TestUseCasesRunAsTheSystem(t *testing.T) {
	a := app()
	perms := permission.All()
	if len(perms) < 10 {
		t.Fatalf("only %d permissions", len(perms))
	}
	for _, p := range perms {
		if err := a.Subject.Audit(p.ID); err != nil {
			t.Errorf("%s is refused", p.ID)
		}
	}
	spec.Verified(t, dec.RDecCliRunsAsSystem)
}

var german = regexp.MustCompile(`[äöüÄÖÜß]|\b(?:und|nicht|oder|wird|werden|Datei|Fehler|Anwendung)\b`)

// Everything the program prints, and every comment in its source, is English.
func TestEnglishOnly(t *testing.T) {
	for name, text := range everything(t) {
		for _, l := range strings.Split(text, "\n") {
			if german.MatchString(l) {
				t.Errorf("%s: %q", name, l)
			}
		}
	}
	fset := token.NewFileSet()
	_ = filepath.WalkDir("../..", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.Contains(p, "/requirements/") {
			return err
		}
		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, cg := range f.Comments {
			if german.MatchString(cg.Text()) {
				t.Errorf("%s: %q", p, cg.Text())
			}
		}
		return nil
	})
	spec.Verified(t, nfr.RNfrEnglish)
}

// goBuild compiles the test program for macOS.
func goBuild(t *testing.T, dir, arch string) string {
	t.Helper()
	out := filepath.Join(dir, "hello-"+arch)
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = "testdata/hello"
	cmd.Env = append(os.Environ(), "GOOS=darwin", "GOARCH="+arch, "CGO_ENABLED=0", "GOFLAGS=-mod=mod")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s", b)
	}
	return out
}

// The whole way from a compiled Go program to a bundle and a zip archive,
// through the real program and the real file system, on whatever operating
// system the test runs on.
func TestFromAGoProgramToABundleOnAnySystem(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a program")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	dir := t.TempDir()
	arm := goBuild(t, dir, "arm64")
	amd := goBuild(t, dir, "amd64")
	icon, _ := os.ReadFile("testdata/icon.png")
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), icon, 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := invoke(t, dir, "build", "--binary", filepath.Base(arm), "--binary", "amd64="+filepath.Base(amd),
		"--name", "Hello World", "--identifier", "com.example.hello", "--version", "1.2.3",
		"--icon", "icon.png", "--icon-shape", "squircle", "--format", "json", "--timestamp", "2026-01-01T00:00:00Z")
	if code != 0 {
		t.Fatalf("exit %d: %s %s", code, out, errOut)
	}
	app := filepath.Join(dir, "dist", "Hello World.app")
	data, err := os.ReadFile(filepath.Join(app, "Contents", "MacOS", "hello"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := macho.Inspect(data)
	if err != nil || info.Kind != macho.Universal || !slices.Equal(info.Archs(), []string{"x86_64", "arm64"}) {
		t.Fatalf("%+v %v", info, err)
	}
	st, _ := os.Stat(filepath.Join(app, "Contents", "MacOS", "hello"))
	if st.Mode()&0o111 == 0 {
		t.Fatal("the program is not executable")
	}
	raw, _ := os.ReadFile(filepath.Join(app, "Contents", "Info.plist"))
	v, err := plist.Decode(raw)
	if err != nil || v.(map[string]any)["CFBundleIdentifier"] != "com.example.hello" {
		t.Fatalf("%v %v", v, err)
	}

	if code, _, e := invoke(t, dir, "zip", app); code != 0 {
		t.Fatal(e)
	}
	if _, err := os.Stat(filepath.Join(dir, "dist", "HelloWorld-1.2.3.zip")); err != nil {
		t.Fatal(err)
	}

	// The same program for Linux is refused with the way to fix it.
	code, _, errOut = invoke(t, dir, "build", "--binary", "icon.png", "--name", "X", "--identifier", "com.example.x", "--version", "1", "--force")
	if code != 1 || !strings.Contains(errOut, "WAP-BINARY-NOT-MACHO") {
		t.Fatalf("exit %d: %s", code, errOut)
	}

	// Where Apple's tools are missing, the commands that need them say so.
	if runtime.GOOS != "darwin" {
		code, _, errOut = invoke(t, dir, "sign", app, "--adhoc")
		if code != 3 || !strings.Contains(errOut, "WAP-PLATFORM-MACOS") {
			t.Fatalf("exit %d: %s", code, errOut)
		}
	}
	spec.Verified(t, nfr.RNfrCrossPlatformCore)
}

// The parts that do not need Apple compile for every system.
func TestCompilesForOtherSystems(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the program")
	}
	for _, target := range [][2]string{{"linux", "amd64"}, {"windows", "amd64"}, {"darwin", "arm64"}} {
		cmd := exec.Command("go", "build", "-o", os.DevNull, ".")
		cmd.Env = append(os.Environ(), "GOOS="+target[0], "GOARCH="+target[1], "CGO_ENABLED=0")
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s/%s: %s", target[0], target[1], b)
		}
	}
}

// On a Mac, the bundle built is signed ad hoc and starts.
func TestOnAMacTheBundleIsSignedAndStarts(t *testing.T) {
	if runtime.GOOS != "darwin" || testing.Short() {
		t.Skip("needs macOS")
	}
	dir := t.TempDir()
	prog := goBuild(t, dir, runtime.GOARCH)
	if code, _, e := invoke(t, dir, "build", "--binary", filepath.Base(prog), "--name", "Hello", "--identifier", "com.example.hello", "--version", "1.0"); code != 0 {
		t.Fatal(e)
	}
	app := filepath.Join(dir, "dist", "Hello.app")
	if code, _, e := invoke(t, dir, "sign", app, "--adhoc"); code != 0 {
		t.Fatal(e)
	}
	out, err := exec.Command(filepath.Join(app, "Contents", "MacOS", "hello")).Output()
	if err != nil || strings.TrimSpace(string(out)) != "hello from a bundle" {
		t.Fatalf("%q %v", out, err)
	}
	if code, _, e := invoke(t, dir, "dmg", app); code != 0 {
		t.Fatal(e)
	}
}

func TestConfigFieldsAreConsistent(t *testing.T) {
	if err := confdoc.Check(configFields()); err != nil {
		t.Fatal(err)
	}
}

// A build server's steps need no paths: build writes the bundle the
// configuration describes, and zip finds it there.
func TestStepsWithoutPaths(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a program")
	}
	dir := t.TempDir()
	prog := goBuild(t, dir, "arm64")
	cfg := `{"name": "Hello", "identifier": "com.example.hello", "version": "2.0",
  "binaries": [{"path": "` + filepath.Base(prog) + `"}], "out": "release"}`
	if err := os.WriteFile(filepath.Join(dir, "wdyapppack.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, e := invoke(t, dir, "build"); code != 0 {
		t.Fatal(e)
	}
	if code, _, e := invoke(t, dir, "zip"); code != 0 {
		t.Fatal(e)
	}
	sum, err := os.ReadFile(filepath.Join(dir, "release", "Hello-2.0.zip.sha256"))
	if err != nil || !strings.HasSuffix(string(sum), "  Hello-2.0.zip\n") {
		t.Fatalf("%q %v", sum, err)
	}
	if _, err := exec.LookPath("shasum"); err == nil {
		cmd := exec.Command("shasum", "-a", "256", "-c", "Hello-2.0.zip.sha256")
		cmd.Dir = filepath.Join(dir, "release")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("shasum does not accept the checksum file: %s", out)
		}
	}
	if code, out, e := invoke(t, dir, "sign", "--adhoc", "--dry-run"); code != 0 && code != 3 {
		t.Fatalf("exit %d: %s %s", code, out, e)
	}
}
