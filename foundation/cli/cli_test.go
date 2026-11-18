package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

// sample is a small program with one command of each kind.
func sample() *App {
	a := &App{
		Name:      "tool",
		Tagline:   "does things",
		Intro:     "QUICK START\n  tool make --name x",
		AgentHint: "Use --format json.",
		Groups:    []string{"Create"},
		Subject:   auth.System(),
		Topics: []Topic{
			{Name: "things", Summary: "what things are", Body: func() string { return "Things are things." }},
		},
		Errors: []problem.Doc{{Code: "WAP-THING", Summary: "a thing went wrong", See: "things"}},
		Commands: []Command{{
			Name:        "make",
			Group:       "Create",
			Summary:     "make a thing",
			Description: "Makes a thing.",
			Args:        []Arg{{Name: "dir", Usage: "where to put it"}},
			Flags: []Flag{
				{Name: "name", Type: String, Usage: "the name", Config: "name"},
				{Name: "tag", Type: List, Usage: "a tag"},
				{Name: "loud", Type: Bool, Usage: "be loud"},
				{Name: "shape", Type: Enum, Enum: []string{"round", "square"}, Usage: "the shape"},
			},
			Writes:    true,
			Replaces:  true,
			Platforms: "every operating system",
			Examples:  []Example{{"make a thing", "tool make --name x --tag a --tag b out"}},
			SeeAlso:   []string{"things"},
			Run: func(inv *Invocation) (Report, error) {
				if inv.String("name") == "fail" {
					return Report{}, problem.Fail(problem.Message{Code: "WAP-THING", What: "it failed", Why: "because", How: "do not", See: "things"})
				}
				if inv.String("name") == "missing" {
					return Report{}, problem.Missing(problem.Message{Code: "WAP-THING", What: "no tool"})
				}
				return Report{
					Text:  "made " + inv.String("name") + " " + strings.Join(inv.List("tag"), ",") + " " + inv.Arg(0),
					Data:  map[string]any{"name": inv.String("name"), "dry": inv.DryRun, "set": inv.IsSet("name")},
					Notes: []problem.Message{{Code: "WAP-THING", Severity: problem.Note, What: "small note"}},
					Next:  []string{"tool next"},
				}, nil
			},
		}},
	}
	return a.WithHelp()
}

func run(a *App, args ...string) (int, string, string) {
	var out, errOut bytes.Buffer
	code := a.Run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestNoArgumentsPrintsTheOverview(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"--help"}, {"-h"}} {
		code, out, _ := run(sample(), args...)
		if code != ExitOK {
			t.Fatalf("%v: exit %d", args, code)
		}
		for _, want := range []string{"tool - does things", "QUICK START", "COMMANDS", "make", "TOPICS", "things", "EXIT CODES", "FOR AI AGENTS", "--format"} {
			if !strings.Contains(out, want) {
				t.Errorf("%v: overview lacks %q:\n%s", args, want, out)
			}
		}
	}
	spec.Verified(t, guide.RGuideNoArgs)
}

func TestCommandHelpComesFromTheDeclaration(t *testing.T) {
	a := sample()
	for _, args := range [][]string{{"help", "make"}, {"make", "--help"}, {"make", "-h", "--bogus"}} {
		code, out, _ := run(a, args...)
		if code != ExitOK {
			t.Fatalf("%v: exit %d", args, code)
		}
		for _, f := range a.Commands[0].AllFlags() {
			if !strings.Contains(out, "--"+f.Name) {
				t.Errorf("%v: help lacks option --%s", args, f.Name)
			}
		}
		for _, want := range []string{"USAGE", "DESCRIPTION", "ARGUMENTS", "OPTIONS", "PLATFORMS", "EXAMPLES", "SEE ALSO", "repeatable", "config: name"} {
			if !strings.Contains(out, want) {
				t.Errorf("%v: help lacks %q", args, want)
			}
		}
	}
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, e := range a.Commands[0].Examples {
		if err := a.ParseExample(e.Command); err != nil {
			t.Errorf("example %q does not parse: %v", e.Command, err)
		}
	}
	spec.Verified(t, guide.RGuideCommandHelp, dec.RDecHelpSingleSource)
}

func TestTopicsAreHelp(t *testing.T) {
	code, out, _ := run(sample(), "help", "things")
	if code != ExitOK || !strings.Contains(out, "Things are things.") {
		t.Fatalf("exit %d: %s", code, out)
	}
	spec.Verified(t, guide.RGuideHelpTopics)
}

func TestOptionsAndArgumentsMayBeMixed(t *testing.T) {
	code, out, _ := run(sample(), "make", "out", "--name=x", "--tag", "a", "-tag", "b", "--dry-run")
	if code != ExitOK {
		t.Fatalf("exit %d", code)
	}
	if !strings.HasPrefix(out, "made x a,b out") {
		t.Fatalf("got %q", out)
	}
}

func TestMessagesSayWhatWhyAndHow(t *testing.T) {
	code, _, errOut := run(sample(), "make", "--name", "fail")
	if code != ExitFailed {
		t.Fatalf("exit %d", code)
	}
	for _, want := range []string{"error [WAP-THING]: it failed", "Why: because", "How: do not", "See: tool help things"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("message lacks %q:\n%s", want, errOut)
		}
	}

	// All mistakes of one command line are reported together.
	code, _, errOut = run(sample(), "make", "--nmae", "x", "--shape", "rond", "a", "b")
	if code != ExitUsage {
		t.Fatalf("exit %d", code)
	}
	for _, want := range []string{"WAP-OPTION-UNKNOWN", "WAP-OPTION-VALUE", "WAP-ARGUMENT-EXTRA"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("missing %s in:\n%s", want, errOut)
		}
	}
	spec.Verified(t, guide.RGuideMessages)
}

func TestMistakesGetSuggestions(t *testing.T) {
	_, _, errOut := run(sample(), "mkae")
	if !strings.Contains(errOut, `Did you mean "make"?`) {
		t.Errorf("command: %s", errOut)
	}
	_, _, errOut = run(sample(), "make", "--nmae", "x")
	if !strings.Contains(errOut, "Did you mean --name?") {
		t.Errorf("option: %s", errOut)
	}
	_, _, errOut = run(sample(), "make", "--shape", "rond")
	if !strings.Contains(errOut, `Did you mean "round"?`) {
		t.Errorf("enum: %s", errOut)
	}
	_, _, errOut = run(sample(), "help", "thigns")
	if !strings.Contains(errOut, `Did you mean "things"?`) {
		t.Errorf("topic: %s", errOut)
	}
	if _, ok := Suggest("zzzzzz", []string{"make"}); ok {
		t.Error("an unrelated word got a suggestion")
	}
	spec.Verified(t, guide.RGuideSuggest)
}

func TestJSONForResultsErrorsAndHelp(t *testing.T) {
	code, out, _ := run(sample(), "--format", "json", "make", "--name", "x")
	if code != ExitOK {
		t.Fatalf("exit %d", code)
	}
	var ok struct {
		OK     bool              `json:"ok"`
		Result map[string]any    `json:"result"`
		Notes  []problem.Message `json:"notes"`
		Next   []string          `json:"next"`
	}
	if err := json.Unmarshal([]byte(out), &ok); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if !ok.OK || ok.Result["name"] != "x" || ok.Result["set"] != true || len(ok.Notes) != 1 || len(ok.Next) != 1 {
		t.Fatalf("unexpected result: %s", out)
	}

	code, out, _ = run(sample(), "make", "--name", "fail", "--format=json")
	var bad struct {
		OK       bool              `json:"ok"`
		ExitCode int               `json:"exitCode"`
		Errors   []problem.Message `json:"errors"`
	}
	if err := json.Unmarshal([]byte(out), &bad); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if bad.OK || bad.ExitCode != code || bad.Errors[0].Why == "" {
		t.Fatalf("unexpected failure document: %s", out)
	}

	// Even a usage error is reported in the format that was asked for.
	_, out, _ = run(sample(), "make", "--bogus", "--format", "json")
	if !json.Valid([]byte(out)) {
		t.Fatalf("usage error not JSON: %s", out)
	}

	_, out, _ = run(sample(), "help", "--format", "json")
	var d Description
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Commands) != 2 || len(d.Topics) != 1 || len(d.ExitCodes) != 4 || len(d.Errors) != 1 {
		t.Fatalf("incomplete description: %+v", d)
	}
	spec.Verified(t, guide.RGuideJson)
}

func TestExitCodes(t *testing.T) {
	cases := []struct {
		args []string
		want int
	}{
		{[]string{"make", "--name", "x"}, ExitOK},
		{[]string{"make", "--name", "fail"}, ExitFailed},
		{[]string{"make", "--bogus"}, ExitUsage},
		{[]string{"nothing"}, ExitUsage},
		{[]string{"make", "--name", "missing"}, ExitUnavailable},
		{[]string{"help"}, ExitOK},
	}
	for _, c := range cases {
		if code, _, _ := run(sample(), c.args...); code != c.want {
			t.Errorf("%v: exit %d, want %d", c.args, code, c.want)
		}
	}
	spec.Verified(t, guide.RGuideExitCodes)
}

func TestNextStepsFollowASuccess(t *testing.T) {
	_, out, errOut := run(sample(), "make", "--name", "x")
	if !strings.Contains(out, "next:\n  tool next") {
		t.Fatalf("no next steps: %s", out)
	}
	if !strings.Contains(errOut, "note [WAP-THING]: small note") {
		t.Fatalf("no note: %s", errOut)
	}
	_, out, errOut = run(sample(), "make", "--name", "x", "--quiet")
	if out != "" || errOut != "" {
		t.Fatalf("quiet printed %q %q", out, errOut)
	}
	spec.Verified(t, guide.RGuideNextSteps)
}

// The program never reads standard input: everything it needs is on the
// command line, and what is missing is an error rather than a question.
func TestNeverAsks(t *testing.T) {
	a := sample()
	a.Commands[0].Args[0].Required = true
	code, _, errOut := run(a, "make")
	if code != ExitUsage || !strings.Contains(errOut, "WAP-ARGUMENT-MISSING") {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	spec.Verified(t, nfr.RNfrNonInteractive)
}

func TestWrapKeepsPreformattedLines(t *testing.T) {
	got := Wrap("one two three four\n\n  keep   this\n- item that is long enough", 2, 14)
	want := "  one two\n  three four\n\n    keep   this\n  - item that\n    is long\n    enough\n"
	if got != want {
		t.Fatalf("got\n%q\nwant\n%q", got, want)
	}
}
