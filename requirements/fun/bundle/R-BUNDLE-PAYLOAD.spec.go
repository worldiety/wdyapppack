package bundle

import "github.com/worldiety/speclink/spec"

// RBundlePayload is the requirement "Resources, libraries and helpers".
var RBundlePayload = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-PAYLOAD",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Resources, libraries and helpers",
	Text:       "Resources, frameworks and helpers MUST be copied into their directories in the bundle, and helpers MUST be checked like the main program.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "57-resources-libraries-and-helpers"},
	},
})
