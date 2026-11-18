// Package config holds the requirements about describing a bundle in a configuration file.
package config

import "github.com/worldiety/speclink/spec"

// RConfigFile is the requirement "Configuration file".
var RConfigFile = spec.Declare(spec.Requirement{
	ID:         "R-CONFIG-FILE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Configuration file",
	Text:       "Everything a build needs MUST be expressible in a JSON configuration file whose relative paths are resolved against the file's directory.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "41-configuration-file"},
	},
})
