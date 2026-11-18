package config

import "github.com/worldiety/speclink/spec"

// RConfigEnv is the requirement "Environment variables".
var RConfigEnv = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-ENV",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Environment variables",
	Text:       "Text values in the configuration MAY refer to environment variables as ${NAME}, and a reference to an unset variable MUST be an error.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "44-environment-variables"},
	},
})
