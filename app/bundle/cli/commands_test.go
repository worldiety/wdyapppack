package clibundle

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/app/bundle"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/cli"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
)

func noEnv(string) (string, bool) { return "", false }

// run builds with a configuration file that sets everything, and returns what
// reached the use case.
func run(t *testing.T, args ...string) bundle.BuildRequest {
	t.Helper()
	fromFile := bundle.Config{
		Name: "File", Identifier: "com.example.file", Version: "1.0", Agent: true,
		Binaries:    []bundle.Binary{{Path: "file-bin"}},
		Environment: map[string]string{"A": "file", "B": "file"},
		Plist:       map[string]any{"K": "file"},
		Icon:        &bundle.Icon{Path: "file.png", Shape: "squircle"},
	}
	load := func(auth.Subject, bundle.LoadConfigRequest) (bundle.LoadedConfig, error) {
		return bundle.LoadedConfig{Config: fromFile, Found: true}, nil
	}
	var got bundle.BuildRequest
	build := func(_ auth.Subject, r bundle.BuildRequest) (bundle.BuildResult, error) {
		got = r
		return bundle.BuildResult{Path: "dist/x.app", Name: "x"}, nil
	}
	app := (&cli.App{Name: "wdyapppack", Subject: auth.System(), Commands: []cli.Command{Build(load, build, noEnv)}}).WithHelp()
	var out, errOut bytes.Buffer
	if code := app.Run(append([]string{"build"}, args...), &out, &errOut); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	return got
}

func TestOptionsOverrideTheFileOnlyWhenGiven(t *testing.T) {
	c := run(t).Config
	if c.Name != "File" || !c.Agent || c.Binaries[0].Path != "file-bin" || c.Icon.Shape != "squircle" {
		t.Fatalf("the file was overridden without an option: %+v", c)
	}

	c = run(t, "--name", "Flag", "--agent=false", "--binary", "arm64=a", "--binary", "b",
		"--env", "B=flag", "--plist", "N:int=3", "--icon", "flag.png").Config
	if c.Name != "Flag" || c.Agent {
		t.Fatalf("single values not overridden: %+v", c)
	}
	if len(c.Binaries) != 2 || c.Binaries[0].Arch != "arm64" || c.Binaries[1].Path != "b" {
		t.Fatalf("the list was not replaced: %+v", c.Binaries)
	}
	if c.Environment["A"] != "file" || c.Environment["B"] != "flag" {
		t.Fatalf("keys not merged: %v", c.Environment)
	}
	if c.Plist["K"] != "file" || c.Plist["N"] != json.Number("3") {
		t.Fatalf("plist not merged: %v", c.Plist)
	}
	if c.Icon.Path != "flag.png" || c.Icon.Shape != "squircle" {
		t.Fatalf("icon %+v", c.Icon)
	}
	spec.Verified(t, config.RConfigOverride)
}

func TestPlistValues(t *testing.T) {
	cases := map[string]any{
		"K=v":          "v",
		"K:string=a=b": "a=b",
		"K:bool=true":  true,
		"K:int=7":      json.Number("7"),
		"K:real=2":     json.Number("2.0"),
		"K:array=a,b":  []any{"a", "b"},
	}
	for in, want := range cases {
		k, v, err := ParsePlistValue(in)
		if err != nil || k != "K" {
			t.Fatalf("%s: %v", in, err)
		}
		if a, ok := want.([]any); ok {
			if len(v.([]any)) != len(a) {
				t.Errorf("%s → %v", in, v)
			}
			continue
		}
		if v != want {
			t.Errorf("%s → %#v", in, v)
		}
	}
	for _, bad := range []string{"K", "K:bool=maybe", "K:int=x", "K:blob=1"} {
		if _, _, err := ParsePlistValue(bad); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}

func TestBinaryPrefixIsOnlyAnArchitectureWhenItIsOne(t *testing.T) {
	if b := ParseBinary("amd64=dist/x"); b.Arch != "amd64" || b.Path != "dist/x" {
		t.Fatal(b)
	}
	if b := ParseBinary("dist/a=b"); b.Arch != "" || b.Path != "dist/a=b" {
		t.Fatal(b)
	}
}

func TestTopicsRender(t *testing.T) {
	for _, topic := range Topics() {
		body := topic.Body()
		if len(body) < 200 {
			t.Errorf("%s is nearly empty", topic.Name)
		}
		if strings.Contains(body, "%USAGE%") {
			t.Errorf("%s has an unfilled placeholder", topic.Name)
		}
		for _, l := range strings.Split(body, "\n") {
			if len([]rune(l)) > cli.Width && !strings.HasPrefix(l, "  ") {
				t.Errorf("%s: line longer than %d: %q", topic.Name, cli.Width, l)
			}
		}
	}
}
