package guide

import "github.com/worldiety/speclink/spec"

// RGuideExitCodes is the requirement "Exit codes".
var RGuideExitCodes = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-EXIT-CODES",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Exit codes",
	Text:       "wdyapppack MUST exit with 0 on success, 1 on a failed operation, 2 on a usage error and 3 when a required tool or platform is missing.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "37-exit-codes"},
	},
})
