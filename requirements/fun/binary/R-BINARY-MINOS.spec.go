package binary

import "github.com/worldiety/speclink/spec"

// RBinaryMinos is the requirement "Minimum macOS version".
var RBinaryMinos = spec.Declare(spec.Requirement{
	ID:         "R-BINARY-MINOS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Minimum macOS version",
	Text:       "The minimum macOS version MUST default to the one the program states, else to 11.0, and a version lower than the program's MUST be refused.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "63-minimum-macos-version"},
	},
})
