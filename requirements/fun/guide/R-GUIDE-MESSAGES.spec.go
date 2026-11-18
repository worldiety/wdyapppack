package guide

import "github.com/worldiety/speclink/spec"

// RGuideMessages is the requirement "Messages that say what to do".
var RGuideMessages = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-MESSAGES",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Messages that say what to do",
	Text:       "Every error and note MUST state what is wrong, why it matters and how to fix it, carry a stable code and name a help topic, and all problems found MUST be reported together.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "34-messages-that-say-what-to-do"},
	},
})
