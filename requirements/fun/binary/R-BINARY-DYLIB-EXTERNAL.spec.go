package binary

import "github.com/worldiety/speclink/spec"

// RBinaryDylibExternal is the requirement "Libraries outside the system".
var RBinaryDylibExternal = spec.Declare(spec.Requirement{
	ID:         "R-BINARY-DYLIB-EXTERNAL",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Libraries outside the system",
	Text:       "Every library a program loads from outside the system MUST be noted with an explanation of how to bundle it.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "65-libraries-outside-the-system"},
	},
})
