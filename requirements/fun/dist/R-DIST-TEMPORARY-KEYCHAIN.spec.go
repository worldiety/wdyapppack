package dist

import "github.com/worldiety/speclink/spec"

// RDistTemporaryKeychain is the requirement "Signing on a build server".
var RDistTemporaryKeychain = spec.Declare(spec.Requirement{
	ID:         "R-DIST-TEMPORARY-KEYCHAIN",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Signing on a build server",
	Text:       "When a PKCS#12 certificate is given, base64 encoded in WDYAPPPACK_CERTIFICATE or as a file by option, with its password from WDYAPPPACK_CERTIFICATE_PASSWORD or by option, every command that signs MUST sign from a temporary keychain holding it, which the command creates and deletes itself, also when it fails or is interrupted, and the named identity MUST be one the certificate contains.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "812-signing-on-a-build-server"},
	},
})
