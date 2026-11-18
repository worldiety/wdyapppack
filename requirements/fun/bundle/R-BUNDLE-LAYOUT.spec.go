// Package bundle holds the requirements about building an application bundle.
package bundle

import "github.com/worldiety/speclink/spec"

// RBundleLayout is the requirement "Bundle layout".
var RBundleLayout = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-LAYOUT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Bundle layout",
	Text:       "A bundle MUST have the layout macOS expects, with an executable program in Contents/MacOS.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "51-bundle-layout"},
		{Extern: "Apple, Bundle Programming Guide: Anatomy of a macOS Application Bundle"},
	},
})
