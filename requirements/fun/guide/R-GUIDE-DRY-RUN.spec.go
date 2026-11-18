package guide

import "github.com/worldiety/speclink/spec"

// RGuideDryRun is the requirement "Dry run".
var RGuideDryRun = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-DRY-RUN",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Dry run",
	Text:       "Every command that writes MUST offer a dry run that shows what would be written or called and writes nothing.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "311-trying-without-writing"},
	},
})
