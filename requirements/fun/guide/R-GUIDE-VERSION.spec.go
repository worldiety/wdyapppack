package guide

import "github.com/worldiety/speclink/spec"

// RGuideVersion is the requirement "Version".
var RGuideVersion = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-VERSION",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Version",
	Text:       "wdyapppack MUST report its own version.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "310-version"},
	},
})
