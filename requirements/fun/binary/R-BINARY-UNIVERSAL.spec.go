package binary

import "github.com/worldiety/speclink/spec"

// RBinaryUniversal is the requirement "Universal programs".
var RBinaryUniversal = spec.Declare(spec.Requirement{
	ID:         "R-BINARY-UNIVERSAL",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Universal programs",
	Text:       "Programs for different architectures MUST be mergeable into one universal program that keeps their signatures, on their own and during a build.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "66-universal-programs"},
	},
})
