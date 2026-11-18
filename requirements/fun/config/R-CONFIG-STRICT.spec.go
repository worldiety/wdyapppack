package config

import "github.com/worldiety/speclink/spec"

// RConfigStrict is the requirement "Unknown fields are refused".
var RConfigStrict = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-STRICT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Unknown fields are refused",
	Text:       "A configuration file with an unknown field MUST be refused.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "43-unknown-fields-are-refused"},
	},
})
