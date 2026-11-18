package guide

import "github.com/worldiety/speclink/spec"

// RGuideJson is the requirement "Machine readable output".
var RGuideJson = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-JSON",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Machine readable output",
	Text:       "Every command MUST be able to report its result, errors and notes as JSON, and the complete help MUST be available as JSON.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "36-machine-readable-output"},
	},
})
