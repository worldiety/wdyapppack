package icon

import "github.com/worldiety/speclink/spec"

// RIconShape is the requirement "Shape for macOS 26".
var RIconShape = spec.Declare(spec.Requirement{
	ID:         "R-ICON-SHAPE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Shape for macOS 26",
	Text:       "On request an icon image MUST be cut into Apple's rounded square with the margin of Apple's grid; by default it MUST be used unchanged, with a note when its corners are not transparent.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "74-shape-for-macos-26"},
	},
})
