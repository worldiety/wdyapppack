package cli

import (
	"reflect"
	"strings"

	"github.com/worldiety/speclink/spec"
)

// UseCaseHelp returns the end user text bound to a use case with spec.Help.
//
// The text is written once, in the annotation beside the use case, and is
// shown both in the derived specification and in the command's help. Reading
// it back through the registry rather than repeating it is what keeps the two
// from drifting apart.
func UseCaseHelp[T any]() string {
	t := reflect.TypeFor[T]()
	name := t.PkgPath() + "." + t.Name()
	var parts []string
	for _, e := range spec.Entries() {
		if e.Target != name {
			continue
		}
		for _, a := range e.Assertions {
			if a.Kind == "help" && a.Text != "" {
				parts = append(parts, strings.TrimSpace(a.Text))
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

// Quote returns a path as it has to be typed in a shell: unchanged when that
// is safe, in single quotes otherwise.
func Quote(s string) string {
	if s != "" && strings.IndexFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("-_./=:@+", r))
	}) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
