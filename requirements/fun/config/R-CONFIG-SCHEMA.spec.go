package config

import "github.com/worldiety/speclink/spec"

// RConfigSchema is the requirement "Schema of the configuration".
var RConfigSchema = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-SCHEMA",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Schema of the configuration",
	Text:       "wdyapppack MUST print a JSON Schema of the configuration file with a description of every field.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "46-schema"},
	},
})
