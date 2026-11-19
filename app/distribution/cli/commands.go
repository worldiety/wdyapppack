// Package clidistribution exposes the distribution context on the command
// line.
package clidistribution

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/cli"
	"github.com/worldiety/wdyapppack/foundation/problem"
)

// Env looks up an environment variable.
type Env func(name string) (string, bool)

// Bundles finds the bundle a configuration file describes. It returns "" when
// the file does not exist or names no bundle.
type Bundles func(configPath string, optional bool) (string, error)

const (
	defaultConfig = "wdyapppack.json"
	macOSOnly     = "macOS only: needs Apple's %s, which come with Xcode or its command line tools (xcode-select --install). On other systems the command explains this and exits with code 3."
)

func configPath(inv *cli.Invocation, env Env) (string, bool) {
	if inv.IsSet("config") {
		return inv.String("config"), false
	}
	if p, ok := env("WDYAPPPACK_CONFIG"); ok && p != "" {
		return p, false
	}
	return defaultConfig, true
}

func settings(inv *cli.Invocation, load distribution.LoadSettings, env Env) (distribution.Settings, error) {
	path, optional := configPath(inv, env)
	got, err := load(inv.Subject, distribution.LoadSettingsRequest{Path: path, Optional: optional, Lookup: env})
	return got.Settings, err
}

// target is the path a command works on: the argument when one is given,
// else the bundle the configuration describes.
func target(inv *cli.Invocation, bundles Bundles, env Env) (string, error) {
	if p := inv.Arg(0); p != "" {
		return p, nil
	}
	path, optional := configPath(inv, env)
	app, err := bundles(path, optional)
	if err != nil {
		return "", err
	}
	if app == "" {
		return "", distribution.NoBundle(path + " does not describe one")
	}
	return app, nil
}

// certificateFlags let a person at a terminal do what a build server does
// through the environment. An option wins over the environment.
func certificateFlags() []cli.Flag {
	return []cli.Flag{
		{Name: "certificate", Type: cli.String, Value: "file.p12",
			Usage: "Sign from a temporary keychain holding this Developer ID certificate (a .p12 with its private key), instead of from your keychain. On a build server, " + distribution.EnvCertificate + " does the same with the file base64 encoded."},
		{Name: "certificate-password", Type: cli.String, Value: "password",
			Usage: "The password of the certificate. Typed here, it ends up in the shell history and is visible in the process list while the command runs; " + distribution.EnvCertificatePassword + " keeps it out of both."},
	}
}

// certificate reads the signing certificate to import into a temporary
// keychain: from the options, else from the environment; nil when there is
// none, and the user's keychain is used.
func certificate(inv *cli.Invocation, env Env) *distribution.Certificate {
	var c distribution.Certificate
	switch {
	case inv.IsSet("certificate") && inv.String("certificate") != "":
		c.File = inv.String("certificate")
	default:
		data, ok := env(distribution.EnvCertificate)
		if !ok || strings.TrimSpace(data) == "" {
			return nil
		}
		c.Base64 = data
	}
	if inv.IsSet("certificate-password") {
		c.Password, c.HasPassword = inv.String("certificate-password"), true
	} else {
		c.Password, c.HasPassword = env(distribution.EnvCertificatePassword)
	}
	return &c
}

func apiKey(env Env) *distribution.APIKey {
	data, ok := env(distribution.EnvAPIKey)
	if !ok || strings.TrimSpace(data) == "" {
		return nil
	}
	id, _ := env(distribution.EnvAPIKeyID)
	issuer, _ := env(distribution.EnvAPIIssuer)
	return &distribution.APIKey{Data: data, KeyID: id, Issuer: issuer}
}

var configFlagWithBundle = cli.Flag{Name: "config", Type: cli.String, Value: "file", Default: defaultConfig + " if it exists",
	Usage: "The configuration file. Without a path argument, the bundle it describes (<out>/<name>.app) is used; its sign, notarize and dmg sections give defaults. Also read from WDYAPPPACK_CONFIG."}

// Sign is the command that signs a bundle.
func Sign(load distribution.LoadSettings, sign distribution.SignBundle, bundles Bundles, env Env) cli.Command {
	return cli.Command{
		Name:    "sign",
		Group:   "Distribute",
		Summary: "sign a bundle, ad hoc or with a Developer ID",
		Description: cli.UseCaseHelp[distribution.SignBundle]() + `

Without a prepared keychain, give the Developer ID certificate as a .p12 file with --certificate and its password with --certificate-password; on a build server, ` + distribution.EnvCertificate + ` holds the file base64 encoded and ` + distribution.EnvCertificatePassword + ` its password. sign then imports it into a temporary keychain of its own and deletes that keychain when it is done, also when it fails or is interrupted. Nothing has to be prepared on the machine. The identity is still named with --identity or "identity" in the sign section, and must be one the certificate contains.`,
		Args: []cli.Arg{{Name: "app", Usage: "the .app bundle to sign; without it, the bundle the configuration describes"}},
		Flags: []cli.Flag{
			configFlagWithBundle,
			{Name: "adhoc", Type: cli.Bool, Usage: "Sign without a certificate. Enough to start the bundle on this Mac; not enough to hand it out."},
			{Name: "identity", Type: cli.String, Value: "name|fingerprint", Config: "sign.identity", Usage: "A Developer ID Application certificate in your keychain: its full name, a unique part of it, or its SHA-1 fingerprint. \"-\" means ad hoc."},
			{Name: "entitlements", Type: cli.String, Value: "file", Config: "sign.entitlements", Usage: "An entitlements property list, applied to the program and the bundle."},
			{Name: "hardened-runtime", Type: cli.Bool, Config: "sign.hardenedRuntime", Default: "on with --identity, off with --adhoc", Usage: "Sign with the hardened runtime, which notarization requires. --hardened-runtime=false switches it off."},
			certificateFlags()[0], certificateFlags()[1],
		},
		Writes: true,
		Examples: []cli.Example{
			cli.Ex("start the bundle on this Mac", `wdyapppack sign "dist/My Tool.app" --adhoc`),
			cli.Ex("sign for distribution, then notarize", `wdyapppack sign "dist/My Tool.app" --identity "Developer ID Application: Example GmbH (TEAM123456)"`),
			cli.Ex("see what would be signed, in which order", `wdyapppack sign "dist/My Tool.app" --adhoc --dry-run`),
			cli.Ex("without a prepared keychain, from a .p12 file", `wdyapppack sign "dist/My Tool.app" --identity "Developer ID Application: Example GmbH" --certificate developer-id.p12 --certificate-password "$P12_PASSWORD"`),
			cli.Ex("on a build server: the bundle and the identity come from wdyapppack.json, the certificate from WDYAPPPACK_CERTIFICATE", `wdyapppack sign`),
		},
		SeeAlso:   []string{"signing", "notarize", "distribution", "ci"},
		Platforms: fmt.Sprintf(macOSOnly, "codesign and security"),
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			s, err := settings(inv, load, env)
			if err != nil {
				return cli.Report{}, err
			}
			app, err := target(inv, bundles, env)
			if err != nil {
				return cli.Report{}, err
			}
			req := distribution.SignRequest{
				App: app, AdHoc: inv.Bool("adhoc"), DryRun: inv.DryRun, Certificate: certificate(inv, env),
				Identity: s.Sign.Identity, Entitlements: s.Sign.Entitlements, HardenedRuntime: s.Sign.HardenedRuntime,
			}
			if inv.IsSet("identity") || inv.Bool("adhoc") {
				req.Identity = inv.String("identity")
			}
			if inv.IsSet("entitlements") {
				req.Entitlements = inv.String("entitlements")
			}
			if inv.IsSet("hardened-runtime") {
				v := inv.Bool("hardened-runtime")
				req.HardenedRuntime = &v
			}
			res, err := sign(inv.Subject, req)
			if err != nil {
				return cli.Report{}, err
			}
			return cli.Report{Text: signText(res), Data: res, Notes: res.Notes, Next: signNext(res)}, nil
		},
	}
}

func signText(r distribution.Signed) string {
	var b strings.Builder
	verb := "Signed"
	if r.DryRun {
		verb = "Would sign (dry run, nothing signed)"
	}
	who := "ad hoc"
	if !r.AdHoc {
		who = "with " + r.Identity
	}
	fmt.Fprintf(&b, "%s %s %s\n", verb, r.Path, who)
	fmt.Fprintf(&b, "  hardened runtime: %s, secure timestamp: %s\n", yesNo(r.HardenedRuntime), yesNo(r.Timestamp))
	if r.TemporaryKeychain {
		fmt.Fprintf(&b, "  certificate: from %s, in a temporary keychain, deleted again\n", distribution.EnvCertificate)
	}
	b.WriteString("\nORDER (inside out)\n")
	for i, o := range r.Order {
		if o == "." {
			o = "the bundle"
		}
		fmt.Fprintf(&b, "  %d. %s\n", i+1, o)
	}
	if r.Verified {
		b.WriteString("\nThe signature verifies (codesign --verify --deep --strict).\n")
	}
	return b.String()
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func signNext(r distribution.Signed) []string {
	app := cli.Quote(r.Path)
	if r.DryRun {
		return []string{"wdyapppack sign " + app + "   (without --dry-run)"}
	}
	if r.AdHoc {
		return []string{"open " + app + "   # starts it on this Mac", "wdyapppack sign " + app + " --identity <Developer ID>   # to hand it out"}
	}
	return []string{"wdyapppack notarize " + app + " --keychain-profile <profile>", "wdyapppack dmg " + app}
}

// Notarize is the command that has a bundle notarized.
func Notarize(load distribution.LoadSettings, notarize distribution.NotarizeArtifact, bundles Bundles, env Env) cli.Command {
	return cli.Command{
		Name:    "notarize",
		Group:   "Distribute",
		Summary: "have Apple notarize it, then staple the ticket",
		Description: cli.UseCaseHelp[distribution.NotarizeArtifact]() + `

On a build server, put an App Store Connect API key into the environment: ` + distribution.EnvAPIKey + ` holds the .p8 file (its text or base64), ` + distribution.EnvAPIKeyID + ` its ID and, for a team key, ` + distribution.EnvAPIIssuer + ` the issuer ID. It is used when no other way is given, and written to a private temporary file only while the command runs. At a terminal, --key, --key-id and --issuer do the same with the .p8 file.`,
		Args: []cli.Arg{{Name: "path", Usage: "the .app, .dmg or .zip to notarize; without it, the bundle the configuration describes"}},
		Flags: []cli.Flag{
			configFlagWithBundle,
			{Name: "keychain-profile", Type: cli.String, Value: "name", Config: "notarize.keychainProfile", Usage: "Credentials stored once with: xcrun notarytool store-credentials <name>. The recommended way."},
			{Name: "key", Type: cli.String, Value: "file.p8", Config: "notarize.key", Usage: "An App Store Connect API key."},
			{Name: "key-id", Type: cli.String, Value: "id", Config: "notarize.keyId", Usage: "The ID of the API key."},
			{Name: "issuer", Type: cli.String, Value: "uuid", Config: "notarize.issuer", Usage: "The issuer ID, for a team API key."},
			{Name: "apple-id", Type: cli.String, Value: "email", Config: "notarize.appleId", Usage: "An Apple ID; needs --team-id and --password-env."},
			{Name: "team-id", Type: cli.String, Value: "id", Config: "notarize.teamId", Usage: "The developer team ID."},
			{Name: "password-env", Type: cli.String, Value: "NAME", Config: "notarize.passwordEnv", Usage: "The environment variable holding the Apple ID's app specific password. That password is never accepted as an option."},
			{Name: "no-staple", Type: cli.Bool, Config: "notarize.staple", Usage: "Do not staple the ticket afterwards."},
		},
		Writes: true,
		Examples: []cli.Example{
			cli.Ex("with a keychain profile (store it once with xcrun notarytool store-credentials)", `wdyapppack notarize "dist/My Tool.app" --keychain-profile wdyapppack`),
			cli.Ex("a disk image, with an API key on a build server", `wdyapppack notarize dist/MyTool-1.2.0.dmg --key AuthKey_ABC123.p8 --key-id ABC123 --issuer 69a6de7e-0000-0000-0000-000000000000`),
			cli.Ex("with an Apple ID; the password is in $AC_PASSWORD", `wdyapppack notarize "dist/My Tool.app" --apple-id dev@example.com --team-id TEAM123456 --password-env AC_PASSWORD`),
			cli.Ex("on a build server: the bundle from wdyapppack.json, the API key from WDYAPPPACK_API_KEY", `wdyapppack notarize`),
		},
		SeeAlso:   []string{"notarization", "sign", "distribution", "ci"},
		Platforms: fmt.Sprintf(macOSOnly, "notarytool and stapler (Xcode 13 or later) and codesign"),
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			s, err := settings(inv, load, env)
			if err != nil {
				return cli.Report{}, err
			}
			n := s.Notarize
			given := false
			for _, f := range []string{"keychain-profile", "key", "key-id", "issuer", "apple-id", "team-id", "password-env"} {
				given = given || inv.IsSet(f)
			}
			if given {
				// Credentials given on the command line replace the file's
				// as a whole, so that two ways never mix.
				n = distribution.NotarizeSettings{Staple: n.Staple}
				n.KeychainProfile, n.Key, n.KeyID, n.Issuer = inv.String("keychain-profile"), inv.String("key"), inv.String("key-id"), inv.String("issuer")
				n.AppleID, n.TeamID, n.PasswordEnv = inv.String("apple-id"), inv.String("team-id"), inv.String("password-env")
			}
			c := distribution.Credentials{KeychainProfile: n.KeychainProfile, Key: n.Key, KeyID: n.KeyID, Issuer: n.Issuer, AppleID: n.AppleID, TeamID: n.TeamID}
			if n.PasswordEnv != "" {
				c.Password, _ = env(n.PasswordEnv)
			}
			staple := n.Staple == nil || *n.Staple
			if inv.IsSet("no-staple") {
				staple = !inv.Bool("no-staple")
			}
			path, err := target(inv, bundles, env)
			if err != nil {
				return cli.Report{}, err
			}
			res, err := notarize(inv.Subject, distribution.NotarizeRequest{
				Path: path, Credentials: c, PasswordEnv: n.PasswordEnv, Staple: staple, APIKey: apiKey(env), DryRun: inv.DryRun,
			})
			if err != nil {
				return cli.Report{}, err
			}
			var b strings.Builder
			if res.DryRun {
				fmt.Fprintf(&b, "Would notarize %s with the %s (dry run, nothing sent):\n", res.Path, res.Credentials)
				for i, s := range res.Steps {
					fmt.Fprintf(&b, "  %d. %s\n", i+1, s)
				}
			} else {
				fmt.Fprintf(&b, "Apple accepted %s (submission %s).\n", res.Path, res.SubmissionID)
				if res.Stapled {
					b.WriteString("The ticket is stapled: Gatekeeper can check it without a network connection.\n")
				}
			}
			next := []string{"spctl --assess --type execute --verbose " + cli.Quote(res.Path) + "   # Gatekeeper's verdict"}
			if strings.HasSuffix(res.Path, ".app") {
				next = append(next, "wdyapppack dmg "+cli.Quote(res.Path), "wdyapppack zip "+cli.Quote(res.Path))
			}
			return cli.Report{Text: b.String(), Data: res, Notes: res.Notes, Next: next}, nil
		},
	}
}

// DMG is the command that makes a disk image.
func DMG(load distribution.LoadSettings, create distribution.CreateDiskImage, bundles Bundles, env Env) cli.Command {
	return cli.Command{
		Name:        "dmg",
		Group:       "Distribute",
		Summary:     "create a disk image to hand out",
		Description: cli.UseCaseHelp[distribution.CreateDiskImage](),
		Args:        []cli.Arg{{Name: "app", Usage: "the .app bundle to put into the image; without it, the bundle the configuration describes"}},
		Flags: []cli.Flag{
			configFlagWithBundle,
			{Name: "out", Type: cli.String, Value: "file.dmg", Default: "<name>-<version>.dmg next to the bundle", Usage: "Where to write the image."},
			{Name: "volume-name", Type: cli.String, Value: "text", Config: "dmg.volumeName", Default: "the bundle's name", Usage: "The name shown in the Finder when the image is opened."},
			{Name: "sign-identity", Type: cli.String, Value: "name|fingerprint", Config: "dmg.sign", Usage: "Also sign the image with this identity. With \"sign\": true in the dmg section, the identity of the sign section is used."},
			certificateFlags()[0], certificateFlags()[1],
		},
		Writes:   true,
		Replaces: true,
		Examples: []cli.Example{
			cli.Ex("an image of a notarized bundle", `wdyapppack dmg "dist/My Tool.app"`),
			cli.Ex("a signed image with its own volume name, to be notarized itself", `wdyapppack dmg "dist/My Tool.app" --volume-name "My Tool 1.2" --sign-identity "Developer ID Application: Example GmbH"`),
		},
		SeeAlso:   []string{"distribution", "notarize", "zip"},
		Platforms: fmt.Sprintf(macOSOnly, "hdiutil (and codesign to sign the image)"),
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			s, err := settings(inv, load, env)
			if err != nil {
				return cli.Report{}, err
			}
			app, err := target(inv, bundles, env)
			if err != nil {
				return cli.Report{}, err
			}
			req := distribution.DiskImageRequest{App: app, Out: inv.String("out"), Volume: s.DMG.VolumeName, Certificate: certificate(inv, env), Force: inv.Force, DryRun: inv.DryRun}
			if s.DMG.Sign {
				req.Identity = s.Sign.Identity
			}
			if inv.IsSet("volume-name") {
				req.Volume = inv.String("volume-name")
			}
			if inv.IsSet("sign-identity") {
				req.Identity = inv.String("sign-identity")
			}
			res, err := create(inv.Subject, req)
			if err != nil {
				return cli.Report{}, err
			}
			verb := "Created"
			if res.DryRun {
				verb = "Would create (dry run)"
			}
			text := fmt.Sprintf("%s %s, volume %q", verb, res.Path, res.Volume)
			if res.Signed {
				text += ", signed"
			}
			text += "\n  checksum: " + res.Checksum
			next := []string{"wdyapppack notarize " + cli.Quote(res.Path) + " --keychain-profile <profile>   # if the bundle inside was not notarized"}
			return cli.Report{Text: text + "\n", Data: res, Notes: res.Notes, Next: next}, nil
		},
	}
}

// Zip is the command that makes a zip archive.
func Zip(create distribution.CreateArchive, bundles Bundles, env Env) cli.Command {
	return cli.Command{
		Name:        "zip",
		Group:       "Distribute",
		Summary:     "create a zip archive that keeps links and permissions",
		Description: cli.UseCaseHelp[distribution.CreateArchive](),
		Args:        []cli.Arg{{Name: "app", Usage: "the .app bundle to archive; without it, the bundle the configuration describes"}},
		Flags: []cli.Flag{
			configFlagWithBundle,
			{Name: "out", Type: cli.String, Value: "file.zip", Default: "<name>-<version>.zip next to the bundle", Usage: "Where to write the archive."},
			{Name: "timestamp", Type: cli.String, Value: "RFC3339", Default: "SOURCE_DATE_EPOCH if set, else the files' own", Usage: "The time recorded for every entry, so that two archives of the same bundle are identical."},
		},
		Writes:   true,
		Replaces: true,
		Examples: []cli.Example{
			cli.Ex("archive a notarized bundle for download", `wdyapppack zip "dist/My Tool.app"`),
		},
		SeeAlso:   []string{"distribution", "dmg"},
		Platforms: "every operating system: the archive is written in Go, like ditto -c -k --keepParent would write it.",
		Run: func(inv *cli.Invocation) (cli.Report, error) {
			var when time.Time
			if s := inv.String("timestamp"); s != "" {
				t, err := time.Parse(time.RFC3339, s)
				if err != nil {
					return cli.Report{}, problem.Misuse(problem.Message{Code: "WAP-OPTION-VALUE", What: fmt.Sprintf("--timestamp %q is not a time in RFC 3339 form.", s), Why: "The value cannot be used as it is written.", How: "Write it like 2026-01-31T12:00:00Z.", See: "zip"})
				}
				when = t
			} else if s, ok := env("SOURCE_DATE_EPOCH"); ok && s != "" {
				if n, err := strconv.ParseInt(s, 10, 64); err == nil {
					when = time.Unix(n, 0).UTC()
				}
			}
			app, err := target(inv, bundles, env)
			if err != nil {
				return cli.Report{}, err
			}
			res, err := create(inv.Subject, distribution.ArchiveRequest{App: app, Out: inv.String("out"), ModTime: when, Force: inv.Force, DryRun: inv.DryRun})
			if err != nil {
				return cli.Report{}, err
			}
			verb := "Created"
			if res.DryRun {
				verb = "Would create (dry run)"
			}
			return cli.Report{Text: verb + " " + res.Path + "\n  checksum: " + res.Checksum + "\n", Data: res, Notes: res.Notes,
				Next: []string{"wdyapppack notarize " + cli.Quote(res.Path) + " --keychain-profile <profile>   # if the bundle inside was not notarized"}}, nil
		},
	}
}
