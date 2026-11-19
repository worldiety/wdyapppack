// Package topology names the world around wdyapppack and every way across its
// boundary.
//
// This is the one part of the model that is declared rather than read from the
// code: nothing in a Go module states that a person or an agent runs it, that
// the file system belongs to whoever runs it, or that Apple's notary service is
// somebody else's system. It is then held against the code: every adapter must
// be named by a channel here, and every channel must name a package that
// exists.
package topology

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/app/bundle"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/dec"
	rbundle "github.com/worldiety/wdyapppack/requirements/fun/bundle"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

// Operator is whoever runs wdyapppack: a person at a terminal, a build
// server's script, or an AI agent. They are one actor because the program
// treats them alike and explains itself to each the same way.
var Operator = spec.Actor{
	ID:        "operator",
	Name:      "Operator",
	Role:      "A person, a build script or an AI agent who packages a program, signs it and hands it out. Acts with the rights of their operating system user.",
	Satisfies: []spec.Requirement{nfr.RNfrSelfExplaining, nfr.RNfrNonInteractive, dec.RDecCliRunsAsSystem},
}

// FileSystem holds the programs, images and configuration read, and the
// bundles, icons and archives written.
var FileSystem = spec.Foreign{
	ID:   "filesystem",
	Name: "File system",
	Role: "The local file system. What is written there outlives the process; access rights and backups belong to whoever runs wdyapppack.",
}

// AppleToolchain is the set of Apple's tools that come with Xcode or its
// command line tools.
var AppleToolchain = spec.Foreign{
	ID:        "apple-toolchain",
	Name:      "Apple toolchain",
	Role:      "codesign, security and hdiutil on macOS: they sign code with certificates from the user's keychain and create disk images. Present only on macOS.",
	Satisfies: []spec.Requirement{cst.RCstAppleToolsForSigning},
}

// AppleNotary is Apple's notary service, reached through notarytool.
var AppleNotary = spec.Foreign{
	ID:        "apple-notary",
	Name:      "Apple notary service",
	Role:      "Apple's online service that scans signed software and issues the ticket Gatekeeper checks. Reached through xcrun notarytool; the ticket is attached with xcrun stapler.",
	Satisfies: []spec.Requirement{dist.RDistNotarize},
}

// BundleCommands is how the operator creates bundles, and what they hand over:
// the configuration file.
var BundleCommands = spec.Channel{
	From:     "operator",
	To:       "app/bundle/cli",
	Label:    "init, build, icon, universal",
	Protocol: "Command line, arguments and wdyapppack.json; text or JSON on standard output",
	Data:     "Program paths, metadata, icon images, the configuration file",
	Auth:     "none of its own; the rights of the operating system user who runs it",
	Crypto:   "none, local",
	Contract: bundle.Config{},
	Satisfies: []spec.Requirement{
		config.RConfigFile, config.RConfigOverride, guide.RGuideNoArgs, guide.RGuideJson,
		rbundle.RBundleLayout,
	},
}

// DistributionCommands is how the operator signs, notarizes and packages.
var DistributionCommands = spec.Channel{
	From:      "operator",
	To:        "app/distribution/cli",
	Label:     "sign, notarize, dmg, zip",
	Protocol:  "Command line, arguments and the sign, notarize and dmg sections of wdyapppack.json",
	Data:      "Bundle paths, signing identities, notarization credentials; on a build server a .p12 certificate and an App Store Connect API key from fixed environment variables (never options)",
	Auth:      "none of its own; the rights of the operating system user, and the keychain they can open or the certificate the environment hands over",
	Crypto:    "none, local",
	Satisfies: []spec.Requirement{dist.RDistNotarizeSecrets, guide.RGuideJson},
}

// BundleFiles is where bundles, icons and universal programs are read and
// written.
var BundleFiles = spec.Channel{
	From:      "app/bundle/adapter/fs",
	To:        "filesystem",
	Label:     "Programs, images and bundles",
	Protocol:  "File system",
	Data:      "Programs, images and property lists read; bundles, icons, universal programs and configuration files written, each moved into place whole",
	Auth:      "Rights of the process user",
	Crypto:    "none, local",
	Satisfies: []spec.Requirement{rbundle.RBundleSafeWrite, guide.RGuideNoOverwrite},
}

// DistributionFiles is where bundles are read and archives written.
var DistributionFiles = spec.Channel{
	From:      "app/distribution/adapter/fs",
	To:        "filesystem",
	Label:     "Bundles and archives",
	Protocol:  "File system",
	Data:      "Bundles read; zip archives written",
	Auth:      "Rights of the process user",
	Crypto:    "none, local",
	Satisfies: []spec.Requirement{dist.RDistZip},
}

// Codesign is the way to Apple's code signing tools.
var Codesign = spec.Channel{
	From:      "app/distribution/adapter/codesign",
	To:        "apple-toolchain",
	Label:     "codesign and security",
	Protocol:  "Process execution",
	Data:      "Paths of code to sign and verify, identity fingerprints, entitlements, a .p12 certificate for a temporary keychain; signatures and identity lists back",
	Auth:      "The user's keychain holds the private keys; on a build server a temporary keychain with a random password, created and deleted by the command",
	Crypto:    "Signatures by codesign; secure timestamps fetched by codesign from Apple over TLS",
	Satisfies: []spec.Requirement{dist.RDistSignAdhoc, dist.RDistSignIdentity, dist.RDistSignVerify, dist.RDistTemporaryKeychain, cst.RCstAppleToolsForSigning},
}

// DiskImages is the way to hdiutil.
var DiskImages = spec.Channel{
	From:      "app/distribution/adapter/hdiutil",
	To:        "apple-toolchain",
	Label:     "hdiutil",
	Protocol:  "Process execution",
	Data:      "A staging directory with the bundle and a link to /Applications; a compressed disk image back",
	Auth:      "Rights of the process user",
	Crypto:    "none, local",
	Satisfies: []spec.Requirement{dist.RDistDmg},
}

// Notarization is the way to Apple's notary service.
var Notarization = spec.Channel{
	From:      "app/distribution/adapter/notarytool",
	To:        "apple-notary",
	Label:     "notarytool and stapler",
	Protocol:  "Process execution of xcrun notarytool, which speaks HTTPS to Apple",
	Data:      "A zip archive or disk image of signed software; a verdict, a log of issues and a ticket back",
	Auth:      "A keychain profile, an App Store Connect API key (from a file or the environment, written to a private temporary file), or an Apple ID with an app specific password",
	Crypto:    "TLS between notarytool and Apple",
	Satisfies: []spec.Requirement{dist.RDistNotarize, dist.RDistNotarizeLog, dist.RDistStaple},
}
