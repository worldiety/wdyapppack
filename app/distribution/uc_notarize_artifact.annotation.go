package distribution

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
)

var _ = spec.For[NotarizeArtifact](
	spec.Satisfies(dist.RDistNotarize, dist.RDistNotarizeSecrets, dist.RDistNotarizeLog, dist.RDistStaple, dist.RDistApiKeyEnv,
		dist.RDistPlatform, cst.RCstAppleToolsForSigning),
	spec.Help(`Submits a bundle, disk image or zip archive to Apple's notary service with xcrun notarytool and waits for the verdict, usually a few minutes. A bundle is zipped for the upload. Before anything is sent, the signature is checked: it must be a Developer ID with the hardened runtime and a secure timestamp, or Apple will reject it.

When Apple accepts, the ticket is stapled to the bundle or disk image, so that Gatekeeper can check it offline. When Apple rejects, its log is fetched and its reasons are shown. The Apple ID password is never accepted as an option: use a keychain profile, an API key, or name the environment variable that holds the password.`),
)
