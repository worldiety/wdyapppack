package guide

import "github.com/worldiety/speclink/spec"

// RGuideNoOverwrite is the requirement "Nothing is overwritten unasked".
var RGuideNoOverwrite = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-NO-OVERWRITE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Nothing is overwritten unasked",
	Text:       "wdyapppack MUST NOT replace an existing file or directory unless told to with --force.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "312-nothing-is-overwritten-unasked"},
	},
})
