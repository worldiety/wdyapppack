package guide

import "github.com/worldiety/speclink/spec"

// RGuideNextSteps is the requirement "Next steps".
var RGuideNextSteps = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-NEXT-STEPS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Next steps",
	Text:       "After a successful command wdyapppack MUST name the commands that usually follow.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "38-next-steps"},
	},
})
