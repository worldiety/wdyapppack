package dist

import "github.com/worldiety/speclink/spec"

// RDistNotarizeSecrets is the requirement "Secrets".
var RDistNotarizeSecrets = spec.Declare(spec.Requirement{
	ID:         "R-DIST-NOTARIZE-SECRETS",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Secrets",
	Text:       "Certificates and keys MUST be given as files or environment variables, never as option content, and passwords MUST come from the keychain or environment variables, except that the password of a signing certificate MAY be given as an option, with the help warning that it then ends up in the shell history.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "86-secrets"},
	},
})
