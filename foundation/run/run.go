// Package run starts external programs and reports what it starts.
//
// The adapters to Apple's tools share it, so that --verbose shows every
// command exactly as it was run, with secrets replaced, and so that a failure
// carries the tool's own words.
package run

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner starts programs.
type Runner struct {
	// Log receives every command line when Verbose reports true.
	Log io.Writer
	// Verbose is asked before every command; nil means quiet.
	Verbose func() bool
}

// Result is what a program printed.
type Result struct {
	Stdout, Stderr []byte
}

// Error is a program that ended unsuccessfully.
type Error struct {
	Command string
	Err     error
	Output  string
}

func (e *Error) Error() string {
	if e.Output != "" {
		return e.Output
	}
	return fmt.Sprintf("%s: %v", e.Command, e.Err)
}

// Run starts a program and waits for it. Values in secrets are replaced by
// *** wherever the command line is shown.
func (r *Runner) Run(secrets []string, name string, args ...string) (Result, error) {
	line := Line(secrets, name, args...)
	if r != nil && r.Log != nil && r.Verbose != nil && r.Verbose() {
		fmt.Fprintf(r.Log, "  $ %s\n", line)
	}
	cmd := exec.Command(name, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	res := Result{Stdout: out.Bytes(), Stderr: errOut.Bytes()}
	if err != nil {
		text := strings.TrimSpace(strings.TrimSpace(errOut.String()) + "\n" + strings.TrimSpace(out.String()))
		for _, s := range secrets {
			if s != "" {
				text = strings.ReplaceAll(text, s, "***")
			}
		}
		return res, &Error{Command: line, Err: err, Output: text}
	}
	return res, nil
}

// Line renders a command line for people, quoting where a shell would need
// it and hiding secrets.
func Line(secrets []string, name string, args ...string) string {
	parts := []string{quote(name)}
	for _, a := range args {
		for _, s := range secrets {
			if s != "" && a == s {
				a = "***"
			}
		}
		parts = append(parts, quote(a))
	}
	return strings.Join(parts, " ")
}

func quote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\"'$`\\*?[]{}()<>|&;#") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
