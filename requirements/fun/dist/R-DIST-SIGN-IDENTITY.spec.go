package dist

import "github.com/worldiety/speclink/spec"

// RDistSignIdentity is the requirement "Developer ID signing".
var RDistSignIdentity = spec.Declare(spec.Requirement{
	ID:         "R-DIST-SIGN-IDENTITY",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Developer ID signing",
	Text:       "A bundle MUST be signable with a Developer ID given by name or fingerprint, with hardened runtime, secure timestamp and optional entitlements, and a missing identity MUST be answered with the available ones.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "82-developer-id-signing"},
	},
})
