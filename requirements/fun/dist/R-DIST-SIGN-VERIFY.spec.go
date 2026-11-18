package dist

import "github.com/worldiety/speclink/spec"

// RDistSignVerify is the requirement "Signature verification".
var RDistSignVerify = spec.Declare(spec.Requirement{
	ID:         "R-DIST-SIGN-VERIFY",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Signature verification",
	Text:       "A signature MUST be verified after signing, and a signature that does not verify MUST fail the command.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "84-signature-verification"},
	},
})
