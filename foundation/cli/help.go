package cli

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// Width is the line width of the help. Eighty columns fit every terminal and
// every code review.
const Width = 80

// WithHelp adds the help command. It is a method rather than a field because
// the help command describes the app it belongs to.
func (a *App) WithHelp() *App {
	a.Commands = append(a.Commands, Command{
		Name:    "help",
		Group:   "Reference",
		Summary: "show help for a command or a topic",
		Description: fmt.Sprintf(`Without an argument, prints the overview: what %[1]s does, a quick start, every command and every topic. With a command or a topic, prints its help.

The same help is available as JSON with --format json: without an argument, the complete description of every command, option, topic, error code, exit code and configuration field; with an argument, that one command or topic.`, a.Name),
		Args:      []Arg{{Name: "command|topic", Usage: "the command or topic to explain"}},
		Platforms: "every operating system",
		Examples: []Example{
			{"the overview", a.Name + " help"},
			{"everything an agent needs, machine readable", a.Name + " help --format json"},
			{"one command", a.Name + " help build"},
			{"one topic", a.Name + " help config"},
		},
		Run: func(inv *Invocation) (Report, error) {
			name := inv.Arg(0)
			if name == "" {
				if inv.Format == "json" {
					return Report{Raw: true, Data: a.Describe()}, nil
				}
				return Report{Raw: true, Text: a.Overview()}, nil
			}
			return a.helpOf(inv, name)
		},
	})
	if a.topic("agents") != nil {
		a.Commands[len(a.Commands)-1].SeeAlso = []string{"agents"}
	}
	return a
}

func (a *App) printHelpOf(inv *Invocation, name string) int {
	r, err := a.helpOf(inv, name)
	if err != nil {
		return a.fail(inv.Stdout, inv.Stderr, inv.Format, inv, err)
	}
	a.printReport(inv, r)
	return ExitOK
}

func (a *App) helpOf(inv *Invocation, name string) (Report, error) {
	if c := a.command(name); c != nil {
		if inv.Format == "json" {
			return Report{Raw: true, Data: a.describeCommand(c)}, nil
		}
		return Report{Raw: true, Text: a.CommandHelp(c)}, nil
	}
	if t := a.topic(name); t != nil {
		if inv.Format == "json" {
			return Report{Raw: true, Data: map[string]any{"name": t.Name, "summary": t.Summary, "body": t.Body()}}, nil
		}
		return Report{Raw: true, Text: a.TopicHelp(t)}, nil
	}

	var all []string
	for _, c := range a.Commands {
		all = append(all, c.Name)
	}
	for _, t := range a.Topics {
		all = append(all, t.Name)
	}
	how := fmt.Sprintf("Run %q to see every command and topic.", a.Name+" help")
	if s, ok := Suggest(name, all); ok {
		how = fmt.Sprintf("Did you mean %q? Run: %s help %s", s, a.Name, s)
	}
	return Report{}, problem.Misuse(problem.Message{
		Code: "WAP-TOPIC-UNKNOWN",
		What: fmt.Sprintf("There is no command or help topic called %q.", name),
		Why:  "help explains the commands and topics listed in the overview.",
		How:  how,
	})
}

// Overview is the text printed without arguments.
func (a *App) Overview() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s - %s\n", a.Name, a.Tagline)
	if a.Intro != "" {
		b.WriteString("\n" + strings.TrimRight(a.Intro, "\n") + "\n")
	}

	b.WriteString("\nCOMMANDS\n")
	width := 0
	for _, c := range a.Commands {
		width = max(width, len(c.Name))
	}
	groupWidth := 0
	for _, g := range a.groups() {
		groupWidth = max(groupWidth, len(g))
	}
	for _, g := range a.groups() {
		first := true
		for _, c := range a.Commands {
			if c.Group != g {
				continue
			}
			label := ""
			if first {
				label = g
				first = false
			}
			line := fmt.Sprintf("  %-*s  %-*s  %s", groupWidth, label, width, c.Name, c.Summary)
			if mark := platformMark(c.Platforms); mark != "" {
				line += "  [" + mark + "]"
			}
			b.WriteString(line + "\n")
		}
	}
	fmt.Fprintf(&b, "\n  Run %q for everything about one command.\n", a.Name+" help <command>")

	b.WriteString("\nTOPICS\n")
	tw := 0
	for _, t := range a.Topics {
		tw = max(tw, len(t.Name))
	}
	for _, t := range a.Topics {
		fmt.Fprintf(&b, "  %-*s  %s\n", tw, t.Name, t.Summary)
	}
	fmt.Fprintf(&b, "\n  Run %q to read one.\n", a.Name+" help <topic>")

	if a.Outro != "" {
		b.WriteString("\n" + strings.TrimRight(a.Outro, "\n") + "\n")
	}

	b.WriteString("\nOPTIONS FOR EVERY COMMAND\n")
	writeFlags(&b, GlobalFlags())
	b.WriteString("\nOPTIONS FOR COMMANDS THAT WRITE\n")
	writeFlags(&b, WriteFlags())

	b.WriteString("\nEXIT CODES\n")
	for _, e := range ExitCodes() {
		b.WriteString(hanging(fmt.Sprintf("%d", e.Code), e.Meaning, 2, 6))
	}

	if a.AgentHint != "" {
		b.WriteString("\nFOR AI AGENTS\n")
		b.WriteString(Wrap(a.AgentHint, 2, Width))
	}
	return b.String()
}

// platformMark shortens the platform note for the overview. Commands that run
// everywhere carry no mark.
func platformMark(p string) string {
	if p == "" || strings.HasPrefix(p, "every") {
		return ""
	}
	mark, _, _ := strings.Cut(p, ":")
	return strings.TrimSuffix(strings.TrimSpace(mark), " only")
}

func (a *App) groups() []string {
	out := slices.Clone(a.Groups)
	for _, c := range a.Commands {
		if !slices.Contains(out, c.Group) {
			out = append(out, c.Group)
		}
	}
	return out
}

// CommandHelp is the help of one command.
func (a *App) CommandHelp(c *Command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s - %s\n", a.Name, c.Name, c.Summary)

	b.WriteString("\nUSAGE\n")
	b.WriteString("  " + a.synopsis(c) + "\n")

	if c.Description != "" {
		b.WriteString("\nDESCRIPTION\n")
		b.WriteString(Wrap(c.Description, 2, Width))
	}

	if len(c.Args) > 0 {
		b.WriteString("\nARGUMENTS\n")
		for _, arg := range c.Args {
			usage := arg.Usage
			if arg.Required {
				usage += " (required)"
			}
			b.WriteString(hanging("<"+arg.Name+">", usage, 2, 6))
		}
	}

	if len(c.Flags) > 0 {
		b.WriteString("\nOPTIONS\n")
		writeFlags(&b, c.Flags)
	}
	if w := c.writeFlags(); len(w) > 0 {
		b.WriteString("\nOPTIONS FOR WRITING\n")
		writeFlags(&b, w)
	}
	b.WriteString("\nOPTIONS FOR EVERY COMMAND\n")
	writeFlags(&b, GlobalFlags())

	b.WriteString("\nPLATFORMS\n")
	b.WriteString(Wrap(c.Platforms, 2, Width))

	if len(c.Examples) > 0 {
		b.WriteString("\nEXAMPLES\n")
		for i, e := range c.Examples {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(Wrap("# "+e.Explanation, 2, Width))
			b.WriteString("  " + e.Command + "\n")
		}
	}

	if len(c.SeeAlso) > 0 {
		b.WriteString("\nSEE ALSO\n")
		for _, s := range c.SeeAlso {
			summary := ""
			if t := a.topic(s); t != nil {
				summary = t.Summary
			} else if o := a.command(s); o != nil {
				summary = o.Summary
			}
			b.WriteString(strings.TrimRight(fmt.Sprintf("  %s help %-14s %s", a.Name, s, summary), " ") + "\n")
		}
	}
	return b.String()
}

func (a *App) synopsis(c *Command) string {
	parts := []string{a.Name, c.Name}
	for _, arg := range c.Args {
		if arg.Required {
			parts = append(parts, "<"+arg.Name+">")
		} else {
			parts = append(parts, "[<"+arg.Name+">]")
		}
	}
	if len(c.AllFlags()) > 0 {
		parts = append(parts, "[options]")
	}
	return strings.Join(parts, " ")
}

func writeFlags(b *strings.Builder, flags []Flag) {
	for _, f := range flags {
		head := "--" + f.Name
		if f.Type != Bool {
			head += " <" + valueName(f) + ">"
		}
		var notes []string
		if f.Type == List {
			notes = append(notes, "repeatable")
		}
		if f.Default != "" {
			notes = append(notes, "default: "+f.Default)
		}
		if f.Config != "" {
			notes = append(notes, "config: "+f.Config)
		}
		usage := f.Usage
		if len(notes) > 0 {
			usage += " (" + strings.Join(notes, "; ") + ")"
		}
		b.WriteString("  " + head + "\n")
		b.WriteString(Wrap(usage, 6, Width))
	}
}

// TopicHelp is the text of one topic.
func (a *App) TopicHelp(t *Topic) string {
	return fmt.Sprintf("%s help %s - %s\n\n%s", a.Name, t.Name, t.Summary, ensureNewline(t.Body()))
}

// Description is the machine readable form of the whole help.
type Description struct {
	Version     int               `json:"version"`
	Name        string            `json:"name"`
	Summary     string            `json:"summary"`
	Usage       string            `json:"usage"`
	Commands    []CommandDesc     `json:"commands"`
	GlobalFlags []Flag            `json:"globalFlags"`
	WriteFlags  []Flag            `json:"writeFlags"`
	Topics      []TopicDesc       `json:"topics"`
	ExitCodes   []ExitCode        `json:"exitCodes"`
	Errors      []problem.Doc     `json:"errors"`
	Extra       map[string]any    `json:"extra,omitempty"`
	Output      map[string]string `json:"output"`
}

// CommandDesc describes one command for machines.
type CommandDesc struct {
	Command
	Usage string `json:"usage"`
}

// TopicDesc describes one topic for machines.
type TopicDesc struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Body    string `json:"body"`
}

// Describe returns the whole help as data.
func (a *App) Describe() Description {
	d := Description{
		Version:     1,
		Name:        a.Name,
		Summary:     a.Tagline,
		Usage:       a.Name + " <command> [arguments] [options]",
		GlobalFlags: GlobalFlags(),
		WriteFlags:  WriteFlags(),
		ExitCodes:   ExitCodes(),
		Errors:      a.Errors,
		Extra:       a.Extra,
		Output: map[string]string{
			"success": `{"version":1,"ok":true,"command":…,"dryRun":…,"result":…,"notes":[message…],"next":[command line…]}`,
			"failure": `{"version":1,"ok":false,"command":…,"exitCode":…,"errors":[message…]}`,
			"message": `{"code":…,"severity":"error"|"note","what":…,"why":…,"how":…,"see":help topic}`,
		},
	}
	for i := range a.Commands {
		d.Commands = append(d.Commands, a.describeCommand(&a.Commands[i]))
	}
	for _, t := range a.Topics {
		d.Topics = append(d.Topics, TopicDesc{Name: t.Name, Summary: t.Summary, Body: t.Body()})
	}
	slices.SortFunc(d.Errors, func(x, y problem.Doc) int { return cmp.Compare(x.Code, y.Code) })
	return d
}

func (a *App) describeCommand(c *Command) CommandDesc {
	return CommandDesc{Command: *c, Usage: a.synopsis(c)}
}

// ErrorsTopic renders the list of message codes.
func (a *App) ErrorsTopic() string {
	docs := slices.Clone(a.Errors)
	slices.SortFunc(docs, func(x, y problem.Doc) int { return cmp.Compare(x.Code, y.Code) })
	var b strings.Builder
	b.WriteString(Wrap(fmt.Sprintf("Every error and note %s prints carries a code. The code is stable: scripts may test for it. Each message also says why it matters, how to fix it, and which help topic explains the background.", a.Name), 0, Width))
	b.WriteString("\n")
	for _, d := range docs {
		b.WriteString(d.Code + "\n")
		text := d.Summary
		if d.See != "" {
			text += " See: " + a.Name + " help " + d.See
		}
		b.WriteString(Wrap(text, 4, Width))
	}
	return b.String()
}

// ExitCodesTopic renders the exit codes.
func (a *App) ExitCodesTopic() string {
	var b strings.Builder
	b.WriteString(Wrap(fmt.Sprintf("%s ends with one of these exit codes. They are part of its interface: scripts may rely on them.", a.Name), 0, Width))
	b.WriteString("\n")
	for _, e := range ExitCodes() {
		b.WriteString(hanging(fmt.Sprintf("%d", e.Code), e.Meaning, 2, 6))
	}
	return b.String()
}

// Validate checks the declarations against each other: every topic named in
// SeeAlso exists, every example names a real command, no option is declared
// twice. It is run by the tests, so that a broken reference never ships.
func (a *App) Validate() error {
	var errs []string
	seenCmd := map[string]bool{}
	for _, c := range a.Commands {
		if seenCmd[c.Name] {
			errs = append(errs, "command declared twice: "+c.Name)
		}
		seenCmd[c.Name] = true
		if c.Summary == "" || c.Description == "" || c.Platforms == "" {
			errs = append(errs, "command "+c.Name+" lacks a summary, description or platform note")
		}
		seenFlag := map[string]bool{}
		for _, f := range c.AllFlags() {
			if seenFlag[f.Name] {
				errs = append(errs, "command "+c.Name+" declares --"+f.Name+" twice")
			}
			seenFlag[f.Name] = true
			if f.Usage == "" {
				errs = append(errs, "command "+c.Name+" leaves --"+f.Name+" unexplained")
			}
			if f.Type == Enum && len(f.Enum) == 0 {
				errs = append(errs, "command "+c.Name+" declares enum --"+f.Name+" without values")
			}
		}
		for _, s := range c.SeeAlso {
			if a.topic(s) == nil && a.command(s) == nil {
				errs = append(errs, "command "+c.Name+" refers to unknown topic "+s)
			}
		}
		if len(c.Examples) == 0 {
			errs = append(errs, "command "+c.Name+" has no example")
		}
		for _, e := range c.Examples {
			f := strings.Fields(e.Command)
			if len(f) < 2 || f[0] != a.Name || a.command(f[1]) == nil {
				errs = append(errs, "example of "+c.Name+" does not start with a command: "+e.Command)
			}
		}
	}
	for _, t := range a.Topics {
		if seenCmd[t.Name] {
			errs = append(errs, "topic and command share the name "+t.Name)
		}
	}
	for _, d := range a.Errors {
		if d.See != "" && a.topic(d.See) == nil && a.command(d.See) == nil {
			errs = append(errs, "error "+d.Code+" refers to unknown topic "+d.See)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

// ParseExample parses an example command line as the program would, so that a
// test can prove every documented example is accepted. Quoted words are kept
// together.
func (a *App) ParseExample(line string) error {
	words := splitWords(line)
	// What follows a redirection, a pipe or a comment belongs to the shell.
	for i, w := range words {
		if w == ">" || w == ">>" || w == "|" || strings.HasPrefix(w, "#") {
			words = words[:i]
			break
		}
	}
	if len(words) < 2 {
		return fmt.Errorf("too short: %s", line)
	}
	cmd := a.command(words[1])
	if cmd == nil {
		return fmt.Errorf("unknown command in: %s", line)
	}
	_, err := a.parse(cmd, words[2:])
	return err
}

// splitWords splits a command line at spaces, keeping quoted text together.
func splitWords(line string) []string {
	var out []string
	var cur strings.Builder
	inQuote := rune(0)
	has := false
	for _, r := range line {
		switch {
		case inQuote != 0 && r == inQuote:
			inQuote = 0
		case inQuote == 0 && (r == '"' || r == '\''):
			inQuote, has = r, true
		case inQuote == 0 && r == ' ':
			if has {
				out = append(out, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
			has = true
		}
	}
	if has {
		out = append(out, cur.String())
	}
	return out
}
