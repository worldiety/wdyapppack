package bundle

import "github.com/worldiety/speclink/spec"

// RBundleStrict is the requirement "Notes and strict mode".
var RBundleStrict = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-STRICT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Notes and strict mode",
	Text:       "Notes MUST NOT fail a build unless strict mode is requested, which MUST turn every note into an error.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "511-notes-and-strict-mode"},
	},
})
