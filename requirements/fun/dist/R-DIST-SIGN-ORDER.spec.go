package dist

import "github.com/worldiety/speclink/spec"

// RDistSignOrder is the requirement "Signing order".
var RDistSignOrder = spec.Declare(spec.Requirement{
	ID:         "R-DIST-SIGN-ORDER",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Signing order",
	Text:       "Code in a bundle MUST be signed from the inside out: frameworks, then helpers, then the program, then the bundle.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "83-signing-order"},
	},
})
