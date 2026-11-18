// Package runtest puts fake command line tools on disk for tests.
//
// A fake records every call in <tool>.log, one line per call with the
// arguments separated by tabs, prints <tool>.out if it exists, and exits with
// the number in <tool>.exit, 0 by default. That is enough to check what an
// adapter asks Apple's tools to do, on any machine with a POSIX shell.
package runtest

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const script = `#!/bin/sh
me="$0"
line=""
for a in "$@"; do line="$line$a	"; done
printf '%s\n' "$line" >> "$me.log"
[ -f "$me.err" ] && cat "$me.err" >&2
[ -f "$me.out" ] && cat "$me.out"
code=0
[ -f "$me.exit" ] && code=$(cat "$me.exit")
exit $code
`

// Tool is a fake program.
type Tool struct {
	Path string
}

// Fake writes a fake tool into dir.
func Fake(t *testing.T, dir, name string) Tool {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake tools are shell scripts")
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return Tool{Path: p}
}

// Says makes the tool print this on standard output.
func (f Tool) Says(t *testing.T, out string) Tool {
	t.Helper()
	if err := os.WriteFile(f.Path+".out", []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

// Complains makes the tool print this on standard error.
func (f Tool) Complains(t *testing.T, msg string) Tool {
	t.Helper()
	if err := os.WriteFile(f.Path+".err", []byte(msg), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

// Fails makes the tool exit with this code.
func (f Tool) Fails(t *testing.T, code int) Tool {
	t.Helper()
	if err := os.WriteFile(f.Path+".exit", []byte(strings.TrimSpace(itoa(code))), 0o644); err != nil {
		t.Fatal(err)
	}
	return f
}

// Calls returns the argument lists of every call so far.
func (f Tool) Calls(t *testing.T) [][]string {
	t.Helper()
	data, err := os.ReadFile(f.Path + ".log")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out [][]string
	for _, l := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		out = append(out, strings.Split(strings.TrimSuffix(l, "\t"), "\t"))
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
