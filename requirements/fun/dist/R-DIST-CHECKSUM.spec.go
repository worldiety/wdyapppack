package dist

import "github.com/worldiety/speclink/spec"

// RDistChecksum is the requirement "Checksums".
var RDistChecksum = spec.Declare(spec.Requirement{
	ID:         "R-DIST-CHECKSUM",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Checksums",
	Text:       "Next to every disk image and zip archive a SHA-256 checksum file MUST be written in the format of shasum -a 256.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "814-checksums"},
	},
})
