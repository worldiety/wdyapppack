package nfr

import "github.com/worldiety/speclink/spec"

// RNfrCrossPlatformCore is the requirement "Core works on every operating system".
var RNfrCrossPlatformCore = spec.Declare(spec.Requirement{
	ID:         "R-NFR-CROSS-PLATFORM-CORE",
	Kind:       spec.NonFunctional,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Core works on every operating system",
	Text:       "Building a bundle, converting an icon, merging a universal program and creating a zip archive MUST work on every operating system Go supports.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "25-works-on-every-operating-system"},
	},
})
