// Package problem carries what went wrong in a form a reader can act on.
//
// Every message states what is wrong, why it matters and how to fix it, and
// names the help topic to read next. A bare error string answers only the
// first of those questions, which is the one the reader could often have
// guessed; the other two are what make a tool explain itself.
//
// The package is shared by the contexts and the command line layer, so that a
// message is written once, where the knowledge is, and rendered by whoever
// talks to the user.
package problem

import (
	"errors"
	"fmt"
	"strings"
)

// Severity tells an error from a note.
type Severity string

const (
	// Error stops the operation.
	Error Severity = "error"
	// Note reports something that does not prevent a working result.
	Note Severity = "note"
)

// Kind says what sort of failure an error is, which decides the exit code.
type Kind int

const (
	// Failed means the operation was attempted and did not succeed, or the
	// input it was given cannot be used.
	Failed Kind = iota
	// Usage means the program was called wrongly: an unknown command or
	// option, a missing value, a malformed configuration.
	Usage
	// Unavailable means something this machine lacks is needed: a tool, or an
	// operating system.
	Unavailable
)

// Message is one thing worth telling the user.
type Message struct {
	// Code is stable and documented, so that scripts and people can look it up.
	Code string `json:"code"`
	// Severity is error or note.
	Severity Severity `json:"severity"`
	// What is wrong, in one sentence naming the thing concerned.
	What string `json:"what"`
	// Why it matters.
	Why string `json:"why,omitempty"`
	// How to fix it; usually a command or a value to change.
	How string `json:"how,omitempty"`
	// See names the help topic that explains the background.
	See string `json:"see,omitempty"`
}

// String renders the message on one line, for logs and test failures.
func (m Message) String() string {
	return fmt.Sprintf("%s [%s]: %s", m.Severity, m.Code, m.What)
}

// AsNote returns the message with the severity of a note.
func (m Message) AsNote() Message { m.Severity = Note; return m }

// AsError returns the message with the severity of an error.
func (m Message) AsError() Message { m.Severity = Error; return m }

// E is a failure that carries messages. It is the only error type the
// contexts return for things a user can fix.
type E struct {
	Kind     Kind
	Messages []Message
}

// Error joins the messages for callers that only want a string.
func (e *E) Error() string {
	parts := make([]string, 0, len(e.Messages))
	for _, m := range e.Messages {
		parts = append(parts, m.String())
	}
	return strings.Join(parts, "; ")
}

// New returns a failure of the given kind made of the given messages. Every
// message is marked as an error.
func New(kind Kind, msgs ...Message) *E {
	out := make([]Message, len(msgs))
	for i, m := range msgs {
		out[i] = m.AsError()
	}
	return &E{Kind: kind, Messages: out}
}

// Fail is shorthand for New(Failed, …).
func Fail(msgs ...Message) *E { return New(Failed, msgs...) }

// Misuse is shorthand for New(Usage, …).
func Misuse(msgs ...Message) *E { return New(Usage, msgs...) }

// Missing is shorthand for New(Unavailable, …).
func Missing(msgs ...Message) *E { return New(Unavailable, msgs...) }

// From extracts the messages of an error. An error that is not an *E becomes a
// single message with the code WAP-INTERNAL, because the user still deserves
// to see it rather than a bare exit code.
func From(err error) *E {
	if err == nil {
		return nil
	}
	var e *E
	if errors.As(err, &e) {
		return e
	}
	return &E{Kind: Failed, Messages: []Message{{
		Code:     "WAP-INTERNAL",
		Severity: Error,
		What:     err.Error(),
		Why:      "This error was not anticipated, so it carries no explanation of its own.",
		How:      "Run the command again with --verbose. If the cause is not clear from the message, report it together with the command you ran.",
		See:      "errors",
	}}}
}

// Collector gathers messages so that every problem is reported at once rather
// than one per run.
type Collector struct {
	errs  []Message
	notes []Message
	kind  Kind
}

// Error records an error. The first error's kind wins, except that a usage
// error outranks a failure: a malformed call is fixed before anything else.
func (c *Collector) Error(kind Kind, m Message) {
	if len(c.errs) == 0 || (kind == Usage && c.kind == Failed) {
		c.kind = kind
	}
	c.errs = append(c.errs, m.AsError())
}

// Note records a note.
func (c *Collector) Note(m Message) { c.notes = append(c.notes, m.AsNote()) }

// Notes returns the notes recorded so far.
func (c *Collector) Notes() []Message { return c.notes }

// Err returns the recorded errors as one failure, nil if there are none.
func (c *Collector) Err() error {
	if len(c.errs) == 0 {
		return nil
	}
	return &E{Kind: c.kind, Messages: c.errs}
}

// Strict turns every note into an error. It is how a build server says that
// nothing may slip through.
func (c *Collector) Strict() {
	for _, n := range c.notes {
		c.Error(Failed, n)
	}
	c.notes = nil
}

// Merge adds the messages of an error to the collector, keeping its kind.
func (c *Collector) Merge(err error) {
	e := From(err)
	if e == nil {
		return
	}
	for _, m := range e.Messages {
		if m.Severity == Note {
			c.Note(m)
			continue
		}
		c.Error(e.Kind, m)
	}
}

// Doc describes one message code for the help topic that lists them.
type Doc struct {
	Code    string `json:"code"`
	Summary string `json:"summary"`
	See     string `json:"see"`
}
