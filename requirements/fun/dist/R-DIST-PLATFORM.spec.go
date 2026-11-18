package dist

import "github.com/worldiety/speclink/spec"

// RDistPlatform is the requirement "Commands that need macOS".
var RDistPlatform = spec.Declare(spec.Requirement{
	ID:         "R-DIST-PLATFORM",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Commands that need macOS",
	Text:       "A command that needs Apple's tools MUST explain, wherever they are missing, what is missing and where to run the command instead.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "811-platform"},
	},
})
