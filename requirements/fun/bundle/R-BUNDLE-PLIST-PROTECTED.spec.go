package bundle

import "github.com/worldiety/speclink/spec"

// RBundlePlistProtected is the requirement "Protected keys".
var RBundlePlistProtected = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-PLIST-PROTECTED",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Protected keys",
	Text:       "Keys derived from the bundle itself MUST NOT be overwritten by additional metadata.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "56-protected-keys"},
	},
})
