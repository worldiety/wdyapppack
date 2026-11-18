package dist

import "github.com/worldiety/speclink/spec"

// RDistNotarizeSecrets is the requirement "Secrets".
var RDistNotarizeSecrets = spec.Declare(spec.Requirement{
	ID:         "R-DIST-NOTARIZE-SECRETS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Secrets",
	Text:       "A password MUST NOT be accepted as an option value; it MUST come from the keychain or from an environment variable whose name is given.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "86-secrets"},
	},
})
