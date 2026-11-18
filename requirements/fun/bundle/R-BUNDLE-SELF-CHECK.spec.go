package bundle

import "github.com/worldiety/speclink/spec"

// RBundleSelfCheck is the requirement "Self-check".
var RBundleSelfCheck = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-SELF-CHECK",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Self-check",
	Text:       "After writing, the bundle MUST be read back and checked for a complete Info.plist, an executable program and the icon it names.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "510-self-check"},
	},
})
