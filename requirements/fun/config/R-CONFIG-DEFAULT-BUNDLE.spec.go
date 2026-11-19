package config

import "github.com/worldiety/speclink/spec"

// RConfigDefaultBundle is the requirement "The bundle from the configuration".
var RConfigDefaultBundle = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-DEFAULT-BUNDLE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "The bundle from the configuration",
	Text:       "Signing, notarizing and wrapping MUST use the bundle the configuration file describes when no bundle is named.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "47-the-bundle-from-the-configuration"},
	},
})
