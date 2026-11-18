package guide

import "github.com/worldiety/speclink/spec"

// RGuideSuggest is the requirement "Suggestions for mistakes".
var RGuideSuggest = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-SUGGEST",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Suggestions for mistakes",
	Text:       "A mistyped command, option, help topic, configuration field or category MUST be answered with the closest valid name.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "35-suggestions-for-mistakes"},
	},
})
