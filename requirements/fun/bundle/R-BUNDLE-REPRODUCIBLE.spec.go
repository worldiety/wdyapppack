package bundle

import "github.com/worldiety/speclink/spec"

// RBundleReproducible is the requirement "Reproducible output".
var RBundleReproducible = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-REPRODUCIBLE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Reproducible output",
	Text:       "The same input MUST produce a byte-identical bundle when the file time is fixed.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "59-reproducible-output"},
	},
})
