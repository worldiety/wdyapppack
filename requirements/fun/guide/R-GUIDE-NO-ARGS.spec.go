// Package guide holds the requirements about how wdyapppack explains itself and talks to its user.
package guide

import "github.com/worldiety/speclink/spec"

// RGuideNoArgs is the requirement "Help without arguments".
var RGuideNoArgs = spec.Declare(spec.Requirement{
	ID:         "R-GUIDE-NO-ARGS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Help without arguments",
	Text:       "Running wdyapppack without arguments or with the command help MUST print the complete overview and exit successfully.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "31-help-without-arguments"},
	},
})
