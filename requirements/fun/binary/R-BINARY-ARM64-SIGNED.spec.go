package binary

import "github.com/worldiety/speclink/spec"

// RBinaryArm64Signed is the requirement "Signature on Apple silicon".
var RBinaryArm64Signed = spec.Declare(spec.Requirement{
	ID:         "R-BINARY-ARM64-SIGNED",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Signature on Apple silicon",
	Text:       "A program for Apple silicon without a signature MUST be noted with an explanation of how to sign it.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "64-signature-on-apple-silicon"},
	},
})
