// Package dist holds the requirements about handing a bundle to other people: signing, notarization, disk images and archives.
package dist

import "github.com/worldiety/speclink/spec"

// RDistSignAdhoc is the requirement "Ad hoc signing".
var RDistSignAdhoc = spec.Declare(spec.Requirement{
	ID:         "R-DIST-SIGN-ADHOC",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Ad hoc signing",
	Text:       "A bundle MUST be signable ad hoc, without a certificate.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "81-ad-hoc-signing"},
	},
})
