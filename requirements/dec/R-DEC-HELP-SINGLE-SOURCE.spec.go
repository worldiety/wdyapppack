package dec

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

// RDecHelpSingleSource is the requirement "Help and options come from one declaration".
var RDecHelpSingleSource = spec.Declare(spec.Requirement{
	ID:           "R-DEC-HELP-SINGLE-SOURCE",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Help and options come from one declaration",
	Text:         "A command's options, its help text and its JSON help MUST be generated from one declaration.",
	Rationale:    "Help written beside the code drifts away from it, and a self-explaining tool whose explanation is wrong is worse than one with none, because it is believed.",
	Consequences: "Every command has to be declared through the foundation's command layer rather than with the flag package directly, and an option's help is limited to what fits its declaration; longer explanations go into help topics.",
	DerivedFrom:  []spec.Requirement{nfr.RNfrSelfExplaining},
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "33-help-for-every-command"},
	},
})
