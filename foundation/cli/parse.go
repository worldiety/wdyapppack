package cli

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/foundation/suggest"
)

// AllFlags returns every option a command accepts: its own, the write options
// if it writes, and the global ones.
func (c *Command) AllFlags() []Flag {
	out := slices.Clone(c.Flags)
	out = append(out, c.writeFlags()...)
	return append(out, GlobalFlags()...)
}

// writeFlags returns the options for writing this command accepts.
func (c *Command) writeFlags() []Flag {
	var out []Flag
	for _, f := range WriteFlags() {
		if (f.Name == "dry-run" && c.Writes) || (f.Name == "force" && c.Replaces) {
			out = append(out, f)
		}
	}
	return out
}

// parse reads the arguments of a command. Options and positional arguments may
// be mixed; "--" ends the options. Every mistake is collected, so that one run
// reports all of them.
func (a *App) parse(cmd *Command, args []string) (*Invocation, error) {
	flags := cmd.AllFlags()
	byName := map[string]Flag{}
	names := make([]string, 0, len(flags))
	for _, f := range flags {
		byName[f.Name] = f
		names = append(names, f.Name)
	}

	inv := &Invocation{Command: cmd, values: map[string]*value{}, Format: "text"}
	for _, f := range flags {
		inv.values[f.Name] = &value{}
	}

	var errs problem.Collector
	usage := func(m problem.Message) {
		if m.See == "" {
			m.See = cmd.Name
		}
		errs.Error(problem.Usage, m)
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			inv.Args = append(inv.Args, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			inv.Args = append(inv.Args, arg)
			continue
		}

		name, val, hasVal := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if name == "h" {
			name = "help"
		}
		f, ok := byName[name]
		if !ok {
			how := fmt.Sprintf("Run %q to see the options of this command.", a.Name+" "+cmd.Name+" --help")
			if s, ok := Suggest(name, names); ok {
				how = fmt.Sprintf("Did you mean --%s? %s", s, how)
			}
			// The next word is most likely the unknown option's value; reading
			// it as an argument would add a second, misleading error.
			if !hasVal && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			usage(problem.Message{
				Code: "WAP-OPTION-UNKNOWN",
				What: fmt.Sprintf("%s does not know the option --%s.", cmd.Name, name),
				Why:  "An option that is not understood would be silently ignored, and the result would not be what was asked for.",
				How:  how,
			})
			continue
		}

		v := inv.values[name]
		if f.Type == Bool {
			b := true
			if hasVal {
				parsed, err := strconv.ParseBool(val)
				if err != nil {
					usage(problem.Message{
						Code: "WAP-OPTION-VALUE",
						What: fmt.Sprintf("--%s takes true or false, not %q.", name, val),
						Why:  "It is a switch: giving it means true.",
						How:  fmt.Sprintf("Write --%s, or --%s=false.", name, name),
					})
					continue
				}
				b = parsed
			}
			v.set, v.b = true, b
			continue
		}

		if !hasVal {
			if i+1 >= len(args) {
				usage(problem.Message{
					Code: "WAP-OPTION-VALUE",
					What: fmt.Sprintf("--%s needs a value.", name),
					Why:  "It was the last argument, so nothing follows it.",
					How:  fmt.Sprintf("Write --%s <%s>.", name, valueName(f)),
				})
				continue
			}
			i++
			val = args[i]
		}

		if f.Type == Enum && !slices.Contains(f.Enum, val) {
			how := fmt.Sprintf("Use one of: %s.", strings.Join(f.Enum, ", "))
			if s, ok := Suggest(val, f.Enum); ok {
				how = fmt.Sprintf("Did you mean %q? %s", s, how)
			}
			usage(problem.Message{
				Code: "WAP-OPTION-VALUE",
				What: fmt.Sprintf("--%s does not accept %q.", name, val),
				Why:  "Only the listed values have a meaning.",
				How:  how,
			})
			continue
		}

		v.set = true
		if f.Type == List {
			v.list = append(v.list, val)
		} else {
			v.str = val
		}
	}

	inv.Format = inv.values["format"].str
	if inv.Format == "" {
		inv.Format = "text"
	}
	inv.Quiet = inv.values["quiet"].b
	inv.Verbose = inv.values["verbose"].b
	if v, ok := inv.values["dry-run"]; ok {
		inv.DryRun = v.b
	}
	if v, ok := inv.values["force"]; ok {
		inv.Force = v.b
	}

	if inv.values["help"].b {
		// Asking for help is never an error, whatever else is wrong.
		return inv, nil
	}

	required := 0
	for _, arg := range cmd.Args {
		if arg.Required {
			required++
		}
	}
	if len(inv.Args) < required {
		missing := cmd.Args[len(inv.Args)]
		usage(problem.Message{
			Code: "WAP-ARGUMENT-MISSING",
			What: fmt.Sprintf("%s needs the argument <%s>: %s.", cmd.Name, missing.Name, strings.TrimSuffix(missing.Usage, ".")),
			Why:  "Without it the command does not know what to work on.",
			How:  fmt.Sprintf("Write %s %s <%s> [options].", a.Name, cmd.Name, missing.Name),
		})
	}
	if len(inv.Args) > len(cmd.Args) {
		usage(problem.Message{
			Code: "WAP-ARGUMENT-EXTRA",
			What: fmt.Sprintf("%s takes %d argument(s), but %d were given: %s.", cmd.Name, len(cmd.Args), len(inv.Args), strings.Join(inv.Args, " ")),
			Why:  "An argument nobody reads is usually an option whose dashes or value went missing, or a path with a space that needs quotes.",
			How:  fmt.Sprintf("Check the command line against %q.", a.Name+" "+cmd.Name+" --help"),
		})
	}

	return inv, errs.Err()
}

func valueName(f Flag) string {
	if f.Value != "" {
		return f.Value
	}
	return string(f.Type)
}

// Suggest returns the candidate closest to what was typed, if any is close
// enough to be a plausible typo.
func Suggest(typed string, candidates []string) (string, bool) {
	return suggest.Closest(typed, candidates)
}
