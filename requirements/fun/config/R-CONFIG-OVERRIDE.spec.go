package config

import "github.com/worldiety/speclink/spec"

// RConfigOverride is the requirement "Options override the file".
var RConfigOverride = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-OVERRIDE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Options override the file",
	Text:       "Every configuration value MUST also be settable as an option, and an option MUST win over the file.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "42-options-override-the-file"},
	},
})
