package guide

import "github.com/worldiety/speclink/spec"

// RGuideRequirements is the requirement "Requirements in the help".
var RGuideRequirements = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-REQUIREMENTS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Requirements in the help",
	Text:       "The requirements MUST be compiled into the program and listed by a help topic.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "39-why-it-behaves-as-it-does"},
	},
})
