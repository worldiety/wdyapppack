package bundle

import "github.com/worldiety/speclink/spec"

// RBundlePlistCustom is the requirement "Additional metadata".
var RBundlePlistCustom = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-PLIST-CUSTOM",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Additional metadata",
	Text:       "Any further Info.plist key MUST be settable with its type, as an option or from a partial property list file.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "55-additional-metadata"},
	},
})
