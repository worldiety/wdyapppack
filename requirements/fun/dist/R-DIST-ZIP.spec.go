package dist

import "github.com/worldiety/speclink/spec"

// RDistZip is the requirement "Zip archives".
var RDistZip = spec.Declare(spec.Requirement{
	ID:         "R-DIST-ZIP",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Zip archives",
	Text:       "A bundle MUST be wrappable, on every operating system, into a zip archive that keeps symbolic links and executable permissions.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "810-zip-archives"},
	},
})
