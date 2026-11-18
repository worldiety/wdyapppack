package dec

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

// RDecPureGoIcns is the requirement "Icons are produced in Go".
var RDecPureGoIcns = spec.Declare(spec.Requirement{
	ID:           "R-DEC-PURE-GO-ICNS",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Icons are produced in Go",
	Text:         "Icons MUST be resized and encoded into .icns by wdyapppack itself, without sips or iconutil.",
	Rationale:    "Apple's image tools exist only on macOS, and icons must be produced on Linux build servers too. The .icns container is a simple list of tagged entries, and current macOS accepts PNG data in every entry it reads.",
	Consequences: "Resampling quality is ours to get right and to test. Icon formats Apple adds later need code here, and asset catalogues (Assets.car) and Icon Composer files cannot be produced at all.",
	DerivedFrom:  []spec.Requirement{nfr.RNfrCrossPlatformCore, cst.RCstStdlibOnly},
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "72-sizes"},
	},
})
