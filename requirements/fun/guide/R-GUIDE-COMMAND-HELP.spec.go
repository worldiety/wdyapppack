package guide

import "github.com/worldiety/speclink/spec"

// RGuideCommandHelp is the requirement "Help for every command".
var RGuideCommandHelp = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-COMMAND-HELP",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Help for every command",
	Text:       "Every command MUST explain its purpose, each option with its type, default and meaning, examples, the platforms it runs on and related topics.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "33-help-for-every-command"},
	},
})
