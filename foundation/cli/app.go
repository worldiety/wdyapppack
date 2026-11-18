package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/problem"
)

// Exit codes. They are part of the interface: scripts branch on them.
const (
	ExitOK          = 0
	ExitFailed      = 1
	ExitUsage       = 2
	ExitUnavailable = 3
)

// ExitCode documents one exit code.
type ExitCode struct {
	Code    int    `json:"code"`
	Meaning string `json:"meaning"`
}

// ExitCodes returns the documented exit codes.
func ExitCodes() []ExitCode {
	return []ExitCode{
		{ExitOK, "success; also for help, which is a successful run rather than an error"},
		{ExitFailed, "the operation failed, or its input cannot be used (for example a program that is not for macOS)"},
		{ExitUsage, "the program was called wrongly: unknown command or option, missing or malformed value, invalid configuration file"},
		{ExitUnavailable, "a required tool or operating system is missing here (for example codesign on Linux)"},
	}
}

func exitCodeOf(k problem.Kind) int {
	switch k {
	case problem.Usage:
		return ExitUsage
	case problem.Unavailable:
		return ExitUnavailable
	}
	return ExitFailed
}

// App is the whole program as the user sees it.
type App struct {
	// Name is the program name, e.g. "wdyapppack".
	Name string
	// Tagline is the one sentence under the name.
	Tagline string
	// Intro is printed after the tagline in the overview: a quick start and
	// whatever a newcomer needs before the list of commands.
	Intro string
	// Outro is printed after the list of topics.
	Outro string
	// AgentHint is printed last, under FOR AI AGENTS.
	AgentHint string
	// Groups orders the command groups in the overview.
	Groups []string
	// Commands and Topics are everything there is.
	Commands []Command
	Topics   []Topic
	// Errors documents every message code, for the errors topic.
	Errors []problem.Doc
	// Extra is added to the JSON description under its keys, e.g. the fields
	// of the configuration file.
	Extra map[string]any
	// Subject is who every command acts as.
	Subject auth.Subject
	// Before, when set, sees every invocation after its arguments were
	// parsed and before it runs, e.g. to pass --verbose on to adapters.
	Before func(inv *Invocation)
}

// GlobalFlags are accepted by every command.
func GlobalFlags() []Flag {
	return []Flag{
		{Name: "format", Type: Enum, Value: "text|json", Default: "text", Enum: []string{"text", "json"},
			Usage: "Output format. json prints one JSON document on standard output, with the result or the errors, for scripts and AI agents."},
		{Name: "quiet", Type: Bool, Usage: "Print nothing but errors (text format only)."},
		{Name: "verbose", Type: Bool, Usage: "Report every step, including the exact external commands run, on standard error."},
		{Name: "help", Type: Bool, Usage: "Show the help of the command instead of running it. -h works too."},
	}
}

// WriteFlags are added to the commands that write something: --dry-run to
// every one, --force to those that create an output that may exist.
func WriteFlags() []Flag {
	return []Flag{
		{Name: "dry-run", Type: Bool, Usage: "Show what would be written and which external tools would be called, and write nothing."},
		{Name: "force", Type: Bool, Usage: "Replace an existing output. Without it, nothing that exists is ever replaced. Accepted by the commands that create an output."},
	}
}

// Run executes the program with the given arguments, without the program name,
// and returns the exit code.
func (a *App) Run(args []string, stdout, stderr io.Writer) int {
	// Global options may come before the command; the first word that is not
	// an option is the command.
	cmdIndex := -1
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			cmdIndex = i
			break
		}
		if isGlobalValueFlag(arg) && !strings.Contains(arg, "=") {
			i++ // --format json: the next word is the value, not the command
		}
	}

	var name string
	var rest []string
	if cmdIndex >= 0 {
		name = args[cmdIndex]
		rest = append(slices.Clone(args[:cmdIndex]), args[cmdIndex+1:]...)
	} else {
		name = "help"
		rest = args
	}
	if name == "help" && cmdIndex < 0 {
		// --version without a command is answered by the version command
		// when there is one.
		for _, r := range rest {
			if (r == "--version" || r == "-version") && a.command("version") != nil {
				name = "version"
				rest = slices.DeleteFunc(slices.Clone(rest), func(s string) bool { return s == "--version" || s == "-version" })
				break
			}
		}
	}

	cmd := a.command(name)
	if cmd == nil {
		format := sniffFormat(args)
		msg := problem.Message{
			Code:     "WAP-COMMAND-UNKNOWN",
			Severity: problem.Error,
			What:     fmt.Sprintf("%q is not a command of %s.", name, a.Name),
			Why:      "Only the commands listed in the help exist.",
			How:      a.suggestCommand(name),
			See:      "",
		}
		return a.fail(stdout, stderr, format, nil, problem.Misuse(msg))
	}

	inv, err := a.parse(cmd, rest)
	if inv == nil {
		inv = &Invocation{Command: cmd, Format: sniffFormat(rest)}
	}
	inv.Stdout, inv.Stderr, inv.Program, inv.Subject = stdout, stderr, a.Name, a.Subject
	if err != nil {
		return a.fail(stdout, stderr, inv.Format, inv, err)
	}

	if inv.Bool("help") && cmd.Name != "help" {
		return a.printHelpOf(inv, cmd.Name)
	}

	if a.Before != nil {
		a.Before(inv)
	}
	report, err := cmd.Run(inv)
	if err != nil {
		return a.fail(stdout, stderr, inv.Format, inv, err)
	}
	a.printReport(inv, report)
	return ExitOK
}

func isGlobalValueFlag(arg string) bool {
	n := strings.TrimLeft(arg, "-")
	return n == "format"
}

// sniffFormat finds --format json before parsing succeeded, so that even a
// usage error is reported in the format the caller asked for.
func sniffFormat(args []string) string {
	for i, a := range args {
		switch a {
		case "--format=json", "-format=json":
			return "json"
		case "--format", "-format":
			if i+1 < len(args) && args[i+1] == "json" {
				return "json"
			}
		}
	}
	return "text"
}

func (a *App) command(name string) *Command {
	for i := range a.Commands {
		if a.Commands[i].Name == name {
			return &a.Commands[i]
		}
	}
	return nil
}

func (a *App) topic(name string) *Topic {
	for i := range a.Topics {
		if a.Topics[i].Name == name {
			return &a.Topics[i]
		}
	}
	return nil
}

func (a *App) suggestCommand(name string) string {
	names := make([]string, 0, len(a.Commands))
	for _, c := range a.Commands {
		names = append(names, c.Name)
	}
	if s, ok := Suggest(name, names); ok {
		return fmt.Sprintf("Did you mean %q? Run: %s %s --help", s, a.Name, s)
	}
	return fmt.Sprintf("Run %q to see every command.", a.Name+" help")
}

// fail prints the messages of an error and returns its exit code.
func (a *App) fail(stdout, stderr io.Writer, format string, inv *Invocation, err error) int {
	e := problem.From(err)
	code := exitCodeOf(e.Kind)
	if format == "json" {
		env := map[string]any{
			"version":  1,
			"ok":       false,
			"exitCode": code,
			"errors":   e.Messages,
		}
		if inv != nil && inv.Command != nil {
			env["command"] = inv.Command.Name
		}
		writeJSON(stdout, env)
		return code
	}
	for i, m := range e.Messages {
		if i > 0 {
			writef(stderr, "\n")
		}
		writeMessage(stderr, a.Name, m)
	}
	return code
}

func (a *App) printReport(inv *Invocation, r Report) {
	if inv.Format == "json" {
		if r.Raw {
			writeJSON(inv.Stdout, r.Data)
			return
		}
		env := map[string]any{
			"version": 1,
			"ok":      true,
			"command": inv.Command.Name,
			"dryRun":  inv.DryRun,
			"result":  r.Data,
			"notes":   nonNil(r.Notes),
			"next":    nonNil(r.Next),
		}
		writeJSON(inv.Stdout, env)
		return
	}
	if r.Raw {
		writef(inv.Stdout, "%s", ensureNewline(r.Text))
		return
	}
	if !inv.Quiet && r.Text != "" {
		writef(inv.Stdout, "%s", ensureNewline(r.Text))
	}
	if !inv.Quiet {
		for _, n := range r.Notes {
			writef(inv.Stderr, "\n")
			writeMessage(inv.Stderr, a.Name, n)
		}
	}
	if !inv.Quiet && len(r.Next) > 0 {
		writef(inv.Stdout, "\nnext:\n")
		for _, n := range r.Next {
			writef(inv.Stdout, "  %s\n", n)
		}
	}
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func ensureNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func writeJSON(w io.Writer, v any) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// writeMessage renders a message in the three part form.
func writeMessage(w io.Writer, program string, m problem.Message) {
	writef(w, "%s [%s]: %s\n", m.Severity, m.Code, m.What)
	if m.Why != "" {
		writef(w, "%s", Wrap("Why: "+m.Why, 4, 80))
	}
	if m.How != "" {
		writef(w, "%s", Wrap("How: "+m.How, 4, 80))
	}
	if m.See != "" {
		writef(w, "    See: %s help %s\n", program, m.See)
	}
}
