package guide

import "github.com/worldiety/speclink/spec"

// RGuideHelpTopics is the requirement "Help topics".
var RGuideHelpTopics = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-HELP-TOPICS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Help topics",
	Text:       "wdyapppack MUST offer help topics on the bundle layout, the configuration, the Info.plist, icons, programs, universal programs, signing, notarization, distribution, categories, recipes, build servers, AI agents, error codes and exit codes.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "32-help-topics"},
	},
})
