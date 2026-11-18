package cli

import (
	"strings"
)

// Wrap formats prose for a terminal.
//
// Paragraphs are separated by empty lines and are filled to the width. A line
// starting with two spaces is preformatted — a command, a tree, a table — and
// is kept exactly, only indented. A line starting with "- " is a list item and
// wraps with a hanging indent.
func Wrap(text string, indent, width int) string {
	pad := strings.Repeat(" ", indent)
	var b strings.Builder
	var para []string
	bullet := false

	flush := func() {
		if len(para) == 0 {
			return
		}
		words := strings.Fields(strings.Join(para, " "))
		first, rest := pad, pad
		if bullet {
			rest = pad + "  "
		}
		line := first
		empty := true
		for _, w := range words {
			if !empty && len(line)+1+len(w) > width {
				b.WriteString(line + "\n")
				line, empty = rest, true
			}
			if !empty {
				line += " "
			}
			line += w
			empty = false
		}
		b.WriteString(line + "\n")
		para, bullet = nil, false
	}

	for _, l := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		switch {
		case strings.TrimSpace(l) == "":
			flush()
			b.WriteString("\n")
		case strings.HasPrefix(l, "  "):
			flush()
			b.WriteString(strings.TrimRight(pad+l, " ") + "\n")
		case strings.HasPrefix(l, "- "):
			flush()
			para, bullet = []string{l}, true
		default:
			para = append(para, l)
		}
	}
	flush()
	return b.String()
}

// hanging renders a label followed by a wrapped explanation: the label on its
// own line when it is long, beside the text when it is short.
func hanging(label, text string, indent, textIndent int) string {
	pad := strings.Repeat(" ", indent)
	if indent+len(label)+2 <= textIndent {
		wrapped := Wrap(text, textIndent, Width)
		return pad + label + wrapped[indent+len(label):]
	}
	return pad + label + "\n" + Wrap(text, textIndent, Width)
}
