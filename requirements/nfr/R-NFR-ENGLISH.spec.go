package nfr

import "github.com/worldiety/speclink/spec"

// RNfrEnglish is the requirement "English".
var RNfrEnglish = spec.Declare(spec.Requirement{
	ID:         "R-NFR-ENGLISH",
	Kind:       spec.NonFunctional,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "English",
	Text:       "The program, its output, its help and its source code MUST be written in English.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "22-english"},
	},
})
