package icon

import "github.com/worldiety/speclink/spec"

// RIconIconset is the requirement "Icon sets".
var RIconIconset = spec.Declare(spec.Requirement{
	ID:         "R-ICON-ICONSET",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Icon sets",
	Text:       "An image MUST be convertible into an .iconset directory.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "75-icon-sets"},
	},
})
