package dist

import "github.com/worldiety/speclink/spec"

// RDistNotarize is the requirement "Notarization".
var RDistNotarize = spec.Declare(spec.Requirement{
	ID:         "R-DIST-NOTARIZE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Notarization",
	Text:       "A bundle, disk image or zip archive signed with a Developer ID MUST be submittable to Apple for notarization, waiting for the result, and an unsuitable signature MUST be refused before anything is sent.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "85-notarization"},
		{Extern: "Apple, Notarizing macOS software before distribution"},
	},
})
