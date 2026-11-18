// Package nfr holds the quality requirements: how wdyapppack behaves rather than what it does.
package nfr

import "github.com/worldiety/speclink/spec"

// RNfrSelfExplaining is the requirement "Self-explaining".
var RNfrSelfExplaining = spec.Declare(spec.Requirement{
	ID:         "R-NFR-SELF-EXPLAINING",
	Kind:       spec.NonFunctional,
	Discipline: spec.Mixed,
	Status:     spec.Normative,
	Title:      "Self-explaining",
	Text:       "wdyapppack MUST give a person or an AI agent everything needed to use it from the program itself, without documentation outside the program.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "21-self-explaining"},
	},
})
