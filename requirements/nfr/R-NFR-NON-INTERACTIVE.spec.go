package nfr

import "github.com/worldiety/speclink/spec"

// RNfrNonInteractive is the requirement "Never interactive".
var RNfrNonInteractive = spec.Declare(spec.Requirement{
	ID:         "R-NFR-NON-INTERACTIVE",
	Kind:       spec.NonFunctional,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Never interactive",
	Text:       "wdyapppack MUST NOT ask questions or wait for input; when information is missing it MUST stop and name what is missing and the alternatives.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "23-never-interactive"},
	},
})
