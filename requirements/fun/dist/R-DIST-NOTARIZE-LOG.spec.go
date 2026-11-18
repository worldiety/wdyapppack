package dist

import "github.com/worldiety/speclink/spec"

// RDistNotarizeLog is the requirement "Rejections".
var RDistNotarizeLog = spec.Declare(spec.Requirement{
	ID:         "R-DIST-NOTARIZE-LOG",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Rejections",
	Text:       "When Apple rejects a submission, its log MUST be fetched and its reasons shown.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "87-rejections"},
	},
})
