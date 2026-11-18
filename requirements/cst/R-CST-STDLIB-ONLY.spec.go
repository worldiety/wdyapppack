// Package cst holds the constraints wdyapppack is built under.
package cst

import "github.com/worldiety/speclink/spec"

// RCstStdlibOnly is the requirement "Standard library only".
var RCstStdlibOnly = spec.Declare(spec.Requirement{
	ID:         "R-CST-STDLIB-ONLY",
	Kind:       spec.Constraint,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Standard library only",
	Text:       "wdyapppack MUST depend on nothing but the Go standard library and the speclink specification package.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "24-standard-library-only"},
	},
})
