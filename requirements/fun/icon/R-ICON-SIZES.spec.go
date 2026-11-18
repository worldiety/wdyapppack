package icon

import "github.com/worldiety/speclink/spec"

// RIconSizes is the requirement "Icon sizes".
var RIconSizes = spec.Declare(spec.Requirement{
	ID:         "R-ICON-SIZES",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Icon sizes",
	Text:       "An image MUST be turned into an .icns file containing every size from 16 to 1024 pixels.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "72-sizes"},
	},
})
