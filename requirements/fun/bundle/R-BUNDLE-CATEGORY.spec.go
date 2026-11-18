package bundle

import "github.com/worldiety/speclink/spec"

// RBundleCategory is the requirement "Application category".
var RBundleCategory = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-CATEGORY",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Application category",
	Text:       "The application category MUST be one that Apple defines, given in full or by its short name.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "54-application-category"},
		{Extern: "Apple, Information Property List Key Reference"},
	},
})
