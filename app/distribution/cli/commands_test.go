package clidistribution

import (
	"bytes"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/cli"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
)

func env(vars map[string]string) Env {
	return func(k string) (string, bool) { v, ok := vars[k]; return v, ok }
}

func noSettings(auth.Subject, distribution.LoadSettingsRequest) (distribution.LoadedSettings, error) {
	return distribution.LoadedSettings{}, nil
}

// Without a path, every command works on the bundle the configuration
// describes; a path given wins.
func TestTheBundleComesFromTheConfiguration(t *testing.T) {
	var asked []string
	bundles := func(path string, optional bool) (string, error) {
		asked = append(asked, path)
		return "dist/From Config.app", nil
	}
	var got []string
	sign := func(_ auth.Subject, r distribution.SignRequest) (distribution.Signed, error) {
		got = append(got, r.App)
		return distribution.Signed{Path: r.App, AdHoc: true}, nil
	}
	notarize := func(_ auth.Subject, r distribution.NotarizeRequest) (distribution.Notarized, error) {
		got = append(got, r.Path)
		return distribution.Notarized{Path: r.Path, DryRun: true}, nil
	}
	dmg := func(_ auth.Subject, r distribution.DiskImageRequest) (distribution.DiskImage, error) {
		got = append(got, r.App)
		return distribution.DiskImage{Path: "x.dmg"}, nil
	}
	zip := func(_ auth.Subject, r distribution.ArchiveRequest) (distribution.Archive, error) {
		got = append(got, r.App)
		return distribution.Archive{Path: "x.zip"}, nil
	}
	e := env(map[string]string{})
	app := (&cli.App{Name: "wdyapppack", Subject: auth.System(), Commands: []cli.Command{
		Sign(noSettings, sign, bundles, e), Notarize(noSettings, notarize, bundles, e),
		DMG(noSettings, dmg, bundles, e), Zip(zip, bundles, e),
	}}).WithHelp()
	run := func(args ...string) {
		t.Helper()
		var out, errOut bytes.Buffer
		if code := app.Run(args, &out, &errOut); code != 0 {
			t.Fatalf("%v: exit %d: %s", args, code, errOut.String())
		}
	}
	run("sign", "--adhoc")
	run("notarize", "--keychain-profile", "p", "--dry-run")
	run("dmg")
	run("zip", "--config", "other.json")
	run("sign", "given.app", "--adhoc")
	want := []string{"dist/From Config.app", "dist/From Config.app", "dist/From Config.app", "dist/From Config.app", "given.app"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v", got)
		}
	}
	if asked[3] != "other.json" {
		t.Fatalf("the configuration named was not asked: %v", asked)
	}

	// No configuration, no path: the error says so.
	none := func(string, bool) (string, error) { return "", nil }
	app = (&cli.App{Name: "wdyapppack", Subject: auth.System(), Commands: []cli.Command{Sign(noSettings, sign, none, e)}}).WithHelp()
	var out, errOut bytes.Buffer
	if code := app.Run([]string{"sign", "--adhoc"}, &out, &errOut); code != cli.ExitUsage || !bytes.Contains(errOut.Bytes(), []byte("WAP-NO-BUNDLE")) {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	spec.Verified(t, config.RConfigDefaultBundle)
}

// The secrets a build server sets reach the use cases; nothing else is needed.
func TestSecretsComeFromTheEnvironment(t *testing.T) {
	vars := map[string]string{
		distribution.EnvCertificate: "UDEy", distribution.EnvCertificatePassword: "",
		distribution.EnvAPIKey: "key", distribution.EnvAPIKeyID: "ID", distribution.EnvAPIIssuer: "ISS",
	}
	var signReq distribution.SignRequest
	var notReq distribution.NotarizeRequest
	sign := func(_ auth.Subject, r distribution.SignRequest) (distribution.Signed, error) {
		signReq = r
		return distribution.Signed{}, nil
	}
	notarize := func(_ auth.Subject, r distribution.NotarizeRequest) (distribution.Notarized, error) {
		notReq = r
		return distribution.Notarized{}, nil
	}
	bundles := func(string, bool) (string, error) { return "a.app", nil }
	app := (&cli.App{Name: "wdyapppack", Subject: auth.System(), Commands: []cli.Command{
		Sign(noSettings, sign, bundles, env(vars)), Notarize(noSettings, notarize, bundles, env(vars)),
	}}).WithHelp()
	var out, errOut bytes.Buffer
	app.Run([]string{"sign", "--identity", "Example"}, &out, &errOut)
	app.Run([]string{"notarize"}, &out, &errOut)
	c := signReq.Certificate
	if c == nil || c.Base64 != "UDEy" || !c.HasPassword || c.Password != "" {
		t.Fatalf("%+v", c)
	}

	// Options win over the environment, each on its own.
	app.Run([]string{"sign", "--identity", "Example", "--certificate", "my.p12", "--certificate-password", "typed"}, &out, &errOut)
	c = signReq.Certificate
	if c.File != "my.p12" || c.Base64 != "" || c.Password != "typed" {
		t.Fatalf("%+v", c)
	}
	app.Run([]string{"sign", "--identity", "Example", "--certificate", "my.p12"}, &out, &errOut)
	if c = signReq.Certificate; c.File != "my.p12" || !c.HasPassword || c.Password != "" {
		t.Fatalf("the password from the environment was not used: %+v", c)
	}
	help := strings.Join(strings.Fields(app.CommandHelp(&app.Commands[0])), " ")
	if !strings.Contains(help, "shell history") || !strings.Contains(help, distribution.EnvCertificatePassword) {
		t.Fatal("the help does not warn about a typed password")
	}
	k := notReq.APIKey
	if k == nil || k.Data != "key" || k.KeyID != "ID" || k.Issuer != "ISS" {
		t.Fatalf("%+v", k)
	}
	spec.Verified(t, dist.RDistTemporaryKeychain, dist.RDistNotarizeSecrets)
}
