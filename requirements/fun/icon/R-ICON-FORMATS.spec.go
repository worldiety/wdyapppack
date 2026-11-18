// Package icon holds the requirements about the application icon.
package icon

import "github.com/worldiety/speclink/spec"

// RIconFormats is the requirement "Accepted images".
var RIconFormats = spec.Declare(spec.Requirement{
	ID:         "R-ICON-FORMATS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Accepted images",
	Text:       "An icon MUST be accepted as a PNG or JPEG image, an .icns file or an .iconset directory.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "71-accepted-images"},
	},
})
