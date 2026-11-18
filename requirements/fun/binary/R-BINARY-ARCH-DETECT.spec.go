package binary

import "github.com/worldiety/speclink/spec"

// RBinaryArchDetect is the requirement "Architectures".
var RBinaryArchDetect = spec.Declare(spec.Requirement{
	ID:         "R-BINARY-ARCH-DETECT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Architectures",
	Text:       "The architectures of a program MUST be recognised, and a stated architecture that does not match the program or is given twice MUST be refused.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "62-architectures"},
	},
})
