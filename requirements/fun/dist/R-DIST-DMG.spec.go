package dist

import "github.com/worldiety/speclink/spec"

// RDistDmg is the requirement "Disk images".
var RDistDmg = spec.Declare(spec.Requirement{
	ID:         "R-DIST-DMG",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Disk images",
	Text:       "A bundle MUST be wrappable into a disk image with the highest compression the format offers that contains a link to the Applications folder.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "89-disk-images"},
	},
})
