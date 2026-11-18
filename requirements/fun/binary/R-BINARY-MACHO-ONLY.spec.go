// Package binary holds the requirements about the programs wdyapppack packages.
package binary

import "github.com/worldiety/speclink/spec"

// RBinaryMachoOnly is the requirement "Only macOS programs".
var RBinaryMachoOnly = spec.Declare(spec.Requirement{
	ID:         "R-BINARY-MACHO-ONLY",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Only macOS programs",
	Text:       "A program that is not a Mach-O executable MUST be refused with an explanation of how to build it for macOS.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "61-only-macos-programs"},
	},
})
