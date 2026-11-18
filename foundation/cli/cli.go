// Package cli is the command line layer: it declares commands, parses their
// arguments, renders their help and reports their results.
//
// The one idea behind it is that a command is declared once. The options it
// accepts, the text that explains them and the machine readable description an
// agent reads are all generated from that declaration, so the help can never
// describe an option the command does not have, or miss one it does.
//
// It deliberately does not use the standard library's flag package for
// parsing. That package stops at the first argument that is not a flag and
// reports mistakes in its own words; a tool that has to explain itself needs to
// accept paths before or after the options and answer a typo with the option
// that was meant. Both are a few dozen lines here.
package cli

import (
	"io"

	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/problem"
)

// FlagType names what an option takes.
type FlagType string

const (
	// String takes one text value; given twice, the last one wins.
	String FlagType = "string"
	// Bool takes no value, or an explicit true or false after "=".
	Bool FlagType = "bool"
	// List takes one value and may be given several times.
	List FlagType = "list"
	// Enum takes one of a fixed set of values.
	Enum FlagType = "enum"
)

// Flag declares one option of a command.
type Flag struct {
	// Name without dashes, e.g. "binary". It is accepted as -binary and --binary.
	Name string `json:"name"`
	// Type decides how the value is read.
	Type FlagType `json:"type"`
	// Value names the value in the help, e.g. "path" or "[arch=]path".
	Value string `json:"value,omitempty"`
	// Default is shown in the help. It is documentation only: the command
	// applies its defaults itself, because some depend on other inputs.
	Default string `json:"default,omitempty"`
	// Usage explains the option in one or two sentences.
	Usage string `json:"usage"`
	// Enum lists the accepted values of an Enum option.
	Enum []string `json:"enum,omitempty"`
	// Config names the field of the configuration file this option overrides.
	Config string `json:"config,omitempty"`
}

// Arg declares a positional argument.
type Arg struct {
	Name     string `json:"name"`
	Usage    string `json:"usage"`
	Required bool   `json:"required"`
}

// Example is a complete command line with what it does.
type Example struct {
	// Explanation says what the example achieves, in a few words.
	Explanation string `json:"explanation"`
	// Command is the full command line, starting with the program name.
	Command string `json:"command"`
}

// Command declares one thing the program can do.
type Command struct {
	Name string `json:"name"`
	// Group orders the overview, e.g. "Create" or "Distribute".
	Group string `json:"group"`
	// Summary is one line for the overview.
	Summary string `json:"summary"`
	// Description is the full explanation. Paragraphs are separated by an
	// empty line; lines indented by two spaces are printed as they are.
	Description string `json:"description"`
	Args        []Arg  `json:"arguments,omitempty"`
	Flags       []Flag `json:"flags,omitempty"`
	// Writes adds --dry-run: every command that writes something, or calls a
	// tool that does, can be tried first.
	Writes bool `json:"writes"`
	// Replaces adds --force: a command that creates an output never replaces
	// an existing one unless told to.
	Replaces bool      `json:"replaces"`
	Examples []Example `json:"examples,omitempty"`
	// SeeAlso names help topics or commands.
	SeeAlso []string `json:"seeAlso,omitempty"`
	// Platforms says where the command runs, e.g. "every operating system"
	// or "macOS only: needs codesign".
	Platforms string `json:"platforms"`
	// Run does the work. It is not part of the description.
	Run func(inv *Invocation) (Report, error) `json:"-"`
}

// Topic is background knowledge that belongs to no single command.
type Topic struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	// Body returns the text. It is a function so that topics generated from
	// data, such as the list of error codes, are built when they are read.
	Body func() string `json:"-"`
}

// Report is what a command hands back on success.
type Report struct {
	// Text is printed as it is in text mode.
	Text string
	// Data is the result in JSON mode. Nil means no result beyond success.
	Data any
	// Notes are things worth knowing that did not prevent the result.
	Notes []problem.Message
	// Next lists the commands that usually follow, as complete command lines.
	Next []string
	// Raw means Text is the whole output in text mode and Data the whole
	// output in JSON mode, with no envelope, notes or next steps. The help
	// and the schema are raw: they are documents, not results.
	Raw bool
}

// Invocation is one run of a command, with everything parsed.
type Invocation struct {
	// Subject is who is acting. A command line tool acts as the operating
	// system user who started it.
	Subject auth.Subject
	// Command is the declaration being run.
	Command *Command
	// Program is the name the program was called by, for messages.
	Program string
	// Args are the positional arguments.
	Args []string
	// Format is "text" or "json".
	Format string
	// Quiet suppresses everything but errors in text mode.
	Quiet bool
	// Verbose asks for every step to be reported.
	Verbose bool
	// DryRun asks to show what would happen without doing it.
	DryRun bool
	// Force allows replacing what exists.
	Force bool
	// Stdout and Stderr are where output goes.
	Stdout, Stderr io.Writer

	values map[string]*value
}

type value struct {
	set  bool
	str  string
	list []string
	b    bool
}

// String returns the value of a string or enum option, "" when not given.
func (i *Invocation) String(name string) string { return i.lookup(name).str }

// Bool returns the value of a bool option.
func (i *Invocation) Bool(name string) bool { return i.lookup(name).b }

// List returns every value of a list option, in the order given.
func (i *Invocation) List(name string) []string { return i.lookup(name).list }

// IsSet reports whether the option was given on the command line. This is how
// an option overrides the configuration file only when it was actually given.
func (i *Invocation) IsSet(name string) bool { return i.lookup(name).set }

// Arg returns the n-th positional argument, "" if there is none.
func (i *Invocation) Arg(n int) string {
	if n < len(i.Args) {
		return i.Args[n]
	}
	return ""
}

func (i *Invocation) lookup(name string) *value {
	if v, ok := i.values[name]; ok {
		return v
	}
	panic("cli: option --" + name + " is not declared by command " + i.Command.Name)
}

// Logf reports a step when --verbose was given. It writes to standard error so
// that standard output stays the result.
func (i *Invocation) Logf(format string, args ...any) {
	if !i.Verbose {
		return
	}
	writef(i.Stderr, "  "+format+"\n", args...)
}

// Ex is shorthand for an example.
func Ex(explanation, command string) Example {
	return Example{Explanation: explanation, Command: command}
}
