package bundle

import "github.com/worldiety/speclink/spec"

// RBundleHelperArchs is the requirement "Helpers per architecture".
var RBundleHelperArchs = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-HELPER-ARCHS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Helpers per architecture",
	Text:       "A helper MUST be accepted as one program or as one program per architecture, several programs of one helper MUST be merged into one universal helper, and a helper that lacks an architecture of the main program MUST be refused.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "57-resources-libraries-and-helpers"},
	},
})
