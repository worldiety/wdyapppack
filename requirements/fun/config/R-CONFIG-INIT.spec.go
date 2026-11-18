package config

import "github.com/worldiety/speclink/spec"

// RConfigInit is the requirement "Starter configuration".
var RConfigInit = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-INIT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Starter configuration",
	Text:       "wdyapppack MUST write a complete starter configuration, filled from the program where possible and with placeholders that a build refuses everywhere else.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "45-starter-configuration"},
	},
})
