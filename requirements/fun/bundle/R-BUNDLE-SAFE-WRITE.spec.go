package bundle

import "github.com/worldiety/speclink/spec"

// RBundleSafeWrite is the requirement "Safe writing".
var RBundleSafeWrite = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-SAFE-WRITE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Safe writing",
	Text:       "A bundle MUST be assembled next to its destination and moved into place only when it is complete.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "58-safe-writing"},
	},
})
