package dist

import "github.com/worldiety/speclink/spec"

// RDistApiKeyEnv is the requirement "Notarization key from the environment".
var RDistApiKeyEnv = spec.Declare(spec.Requirement{
	ID:         "R-DIST-API-KEY-ENV",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Notarization key from the environment",
	Text:       "When no other way to reach the notary service is given, the App Store Connect API key MUST be taken from WDYAPPPACK_API_KEY, base64 encoded or as text, with WDYAPPPACK_API_KEY_ID and optionally WDYAPPPACK_API_ISSUER, and written to a private temporary file only while the command runs.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "813-notarization-key-from-the-environment"},
	},
})
