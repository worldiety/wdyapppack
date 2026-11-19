package distribution

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
)

var _ = spec.For[SignBundle](
	spec.Satisfies(dist.RDistSignAdhoc, dist.RDistSignIdentity, dist.RDistSignOrder, dist.RDistSignVerify,
		dist.RDistPlatform, dist.RDistTemporaryKeychain, dec.RDecTemporaryKeychain,
		cst.RCstAppleToolsForSigning, guide.RGuideDryRun),
	spec.Help(`Signs a bundle with Apple's codesign, from the inside out: libraries and frameworks, then helpers, then further programs, then the program, then the bundle, because signing a container seals the signatures inside it. Afterwards the signature is verified deeply and strictly.

Ad hoc (--adhoc) needs no certificate and is enough to start the bundle on this Mac. To hand it to others, sign it with a Developer ID Application certificate (--identity); the hardened runtime and a secure timestamp are then added, as notarization requires.

On a build server the certificate can come from the environment instead of a prepared keychain. It is then imported into a temporary keychain that exists only while this command runs.`),
)
