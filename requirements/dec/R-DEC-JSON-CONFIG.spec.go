// Package dec holds the technical rulings that shape how wdyapppack is built, each with what it costs.
package dec

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/cst"
)

// RDecJsonConfig is the requirement "The configuration file is JSON".
var RDecJsonConfig = spec.Declare(spec.Requirement{
	ID:           "R-DEC-JSON-CONFIG",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "The configuration file is JSON",
	Text:         "The configuration file MUST be written in JSON.",
	Rationale:    "JSON is read and written by the standard library, while YAML or TOML would need a dependency that the standard library constraint rules out; JSON is also what editors complete and check from a JSON Schema.",
	Consequences: "JSON has no comments, so a configuration cannot explain itself in place: the explanation has to live in the help and the schema. Trailing commas and missing quotes are syntax errors users will meet, so the error has to point at the line.",
	DerivedFrom:  []spec.Requirement{cst.RCstStdlibOnly},
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "41-configuration-file"},
	},
})
