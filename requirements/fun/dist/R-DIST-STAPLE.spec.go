package dist

import "github.com/worldiety/speclink/spec"

// RDistStaple is the requirement "Stapling".
var RDistStaple = spec.Declare(spec.Requirement{
	ID:         "R-DIST-STAPLE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Stapling",
	Text:       "After a successful notarization the ticket MUST be stapled to the bundle or disk image.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "88-stapling"},
	},
})
