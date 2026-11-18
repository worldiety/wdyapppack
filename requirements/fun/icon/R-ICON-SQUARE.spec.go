package icon

import "github.com/worldiety/speclink/spec"

// RIconSquare is the requirement "Square images".
var RIconSquare = spec.Declare(spec.Requirement{
	ID:         "R-ICON-SQUARE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Square images",
	Text:       "An icon image MUST be square and at least 512 pixels wide, and an image below 1024 pixels MUST be noted.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "73-square-images"},
	},
})
