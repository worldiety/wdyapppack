package bundle

import "github.com/worldiety/speclink/spec"

// RBundlePlistDefaults is the requirement "Default metadata".
var RBundlePlistDefaults = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-PLIST-DEFAULTS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Default metadata",
	Text:       "The Info.plist MUST contain every key an application needs, with defaults, together with the optional keys the user sets.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "53-default-metadata"},
		{Extern: "Apple, Information Property List Key Reference"},
	},
})
