package distribution

import (
	"cmp"
	"slices"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

type explanation struct{ summary, why, how, see string }

// explanations lists every message of this context; the help topic on error
// codes is generated from it.
func explanations() map[string]explanation {
	return map[string]explanation{
		"WAP-PLATFORM-MACOS": {"the command needs macOS, and this is another operating system",
			"Code signing, notarization and disk images are done by Apple's tools, which exist only on macOS.",
			"Run this step on a Mac, for example in a macOS job of your build server. build, icon, universal and zip work everywhere, so the bundle can be built elsewhere and only signed on the Mac.", "distribution"},
		"WAP-TOOL-MISSING": {"a tool from Apple that the command needs was not found",
			"These tools come with Xcode or its command line tools.",
			"Install the command line tools with: xcode-select --install", "distribution"},
		"WAP-PATH-NOT-FOUND": {"the bundle or file named does not exist",
			"There is nothing to work on.",
			"Check the path; build writes bundles to dist/ unless told otherwise.", "distribution"},
		"WAP-NOT-A-BUNDLE": {"the path is not an application bundle",
			"Only a directory ending in .app with a Contents/Info.plist can be signed or packaged as an application.",
			"Give the .app directory written by: wdyapppack build", "bundle"},
		"WAP-SIGN-MODE": {"neither or both of --adhoc and --identity were given",
			"A bundle is signed either ad hoc, which needs no certificate and is enough for this Mac, or with a Developer ID, which is needed to hand it to others.",
			"Give --adhoc, or --identity with your Developer ID Application certificate, or set \"identity\" in the sign section of the configuration.", "signing"},
		"WAP-IDENTITY-NOT-FOUND": {"no signing identity matches the one given",
			"codesign can only sign with a certificate whose private key is in your keychain.",
			"Use one of the identities listed, by name or fingerprint. A Developer ID certificate is created in your Apple developer account and installed with Xcode.", "signing"},
		"WAP-IDENTITY-AMBIGUOUS": {"several signing identities match the one given",
			"codesign would not know which to use.",
			"Give the full name, or the fingerprint of the one you mean.", "signing"},
		"WAP-ENTITLEMENTS-MISSING": {"the entitlements file does not exist",
			"Entitlements grant capabilities the hardened runtime otherwise forbids; a missing file would sign without them.",
			"Check the path. Entitlements are an XML property list, usually named <app>.entitlements.", "signing"},
		"WAP-SIGN-FAILED": {"codesign refused to sign",
			"Its own message, shown here, names the reason.",
			"Common causes: the keychain is locked (unlock it or run from a logged in session), the certificate expired, or a file inside the bundle was changed while signing.", "signing"},
		"WAP-SIGN-VERIFY": {"the signature does not verify",
			"A bundle whose signature does not verify is refused by Gatekeeper and by the notary service.",
			"Sign again; if it persists, check codesign's message for the file it names.", "signing"},
		"WAP-NOTARIZE-CREDENTIALS": {"no way, or more than one way, to reach the notary service was given",
			"notarytool needs exactly one: a keychain profile, an App Store Connect API key, or an Apple ID with an app specific password.",
			"The simplest is a keychain profile: run once xcrun notarytool store-credentials <profile>, then pass --keychain-profile <profile>.", "notarization"},
		"WAP-NOTARIZE-PASSWORD": {"the environment variable that should hold the password is not set",
			"Passwords are never taken as options, because those end up in the shell history and the process list.",
			"Set the variable named by --password-env to the app specific password of your Apple ID, or use a keychain profile.", "notarization"},
		"WAP-NOTARIZE-UNSIGNED": {"the bundle is not signed the way notarization requires",
			"Apple notarizes only code signed with a Developer ID, with the hardened runtime and a secure timestamp; anything else is rejected after a long upload.",
			"Sign it first: wdyapppack sign <app> --identity \"Developer ID Application: …\"", "notarization"},
		"WAP-NOTARIZE-FORMAT": {"only a .app, .dmg or .zip can be notarized",
			"The notary service accepts disk images, zip archives and installer packages; a bundle is zipped for upload.",
			"Give the .app, a .dmg made by wdyapppack dmg, or a .zip made by wdyapppack zip.", "notarization"},
		"WAP-NOTARIZE-FAILED": {"the submission to the notary service failed",
			"notarytool's own message, shown here, names the reason; often the credentials or the network.",
			"Check the credentials with: xcrun notarytool history --keychain-profile <profile>", "notarization"},
		"WAP-NOTARIZE-REJECTED": {"Apple rejected the submission",
			"The service found something it does not accept; its reasons are listed.",
			"Fix what the reasons name — most often a file signed without the hardened runtime or timestamp — sign again and resubmit.", "notarization"},
		"WAP-STAPLE-FAILED": {"the ticket could not be stapled",
			"Without the stapled ticket, macOS has to ask Apple online the first time the application is opened.",
			"Try again later with: xcrun stapler staple <path>; the notarization itself succeeded.", "notarization"},
		"WAP-STAPLE-ZIP": {"a zip archive cannot carry a ticket",
			"The ticket is attached to the bundle or disk image, never to an archive; the notarization is still recorded by Apple.",
			"To hand out a zip with a stapled bundle: notarize the .app (which zips it for upload and staples it), then run wdyapppack zip on it.", "notarization"},
		"WAP-DMG-FAILED": {"hdiutil could not create the disk image",
			"Its own message, shown here, names the reason.",
			"Check that there is enough disk space and that no other image of the same name is mounted.", "distribution"},
		"WAP-ZIP-FAILED": {"the zip archive could not be written",
			"The operating system refused to create or move the file.",
			"Check that the output directory is writable and the disk is not full.", "distribution"},
		"WAP-ARTIFACT-EXISTS": {"the disk image or archive exists already",
			"Nothing is replaced unless you ask for it.",
			"Add --force to replace it, or give another --out.", "distribution"},
		"WAP-SETTINGS-MISSING": {"the configuration file named does not exist",
			"A file named with --config must exist; only the default wdyapppack.json may be absent.",
			"Check the path, or write one with: wdyapppack init", "config"},
		"WAP-SETTINGS-INVALID": {"a section sign, notarize or dmg of the configuration file cannot be read",
			"The file is not valid JSON, or a section holds an unknown field or a value of the wrong type.",
			"Fix the section; every field is listed by: wdyapppack help config", "config"},
		"WAP-SETTINGS-ENV-UNSET": {"a section of the configuration refers to an environment variable that is not set",
			"Leaving ${NAME} in place would sign or submit with a wrong value.",
			"Set the variable, or write the value into the file.", "config"},
	}
}

func msg(code, what string) problem.Message {
	e, ok := explanations()[code]
	if !ok {
		panic("distribution: message code " + code + " is not explained")
	}
	return problem.Message{Code: code, Severity: problem.Error, What: what, Why: e.why, How: e.how, See: e.see}
}

func msgHow(code, what, how string) problem.Message {
	m := msg(code, what)
	m.How = how
	return m
}

// Codes documents every message of this context.
func Codes() []problem.Doc {
	var out []problem.Doc
	for code, e := range explanations() {
		out = append(out, problem.Doc{Code: code, Summary: e.summary, See: e.see})
	}
	slices.SortFunc(out, func(a, b problem.Doc) int { return cmp.Compare(a.Code, b.Code) })
	return out
}
