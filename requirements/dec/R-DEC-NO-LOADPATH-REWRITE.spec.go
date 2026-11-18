package dec

import "github.com/worldiety/speclink/spec"

// RDecNoLoadpathRewrite is the requirement "Library load paths are not rewritten".
var RDecNoLoadpathRewrite = spec.Declare(spec.Requirement{
	ID:           "R-DEC-NO-LOADPATH-REWRITE",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Library load paths are not rewritten",
	Text:         "wdyapppack MUST NOT rewrite the library load paths of a program; it MUST only report libraries loaded from outside the system.",
	Rationale:    "Rewriting load commands changes the program, invalidates its signature and is only safe with install_name_tool, which exists only on macOS. A packaging tool that edits the program it packages becomes responsible for that program's correctness.",
	Consequences: "A program linked against libraries from a package manager has to be fixed by whoever builds it before its bundle runs on other Macs; wdyapppack can only point at the problem.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "65-libraries-outside-the-system"},
	},
})
