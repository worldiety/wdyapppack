package distribution

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
)

func sign(ws Workspace, s Signer, req SignRequest) (Signed, error) {
	return NewSignBundle(ws, s)(auth.System(), req)
}

func TestAdHocSigning(t *testing.T) {
	s := &fakeSigner{}
	res, err := sign(bundleWS(t), s, SignRequest{App: app, AdHoc: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range s.opts {
		if o.Identity != "-" || o.Timestamp || o.HardenedRuntime {
			t.Fatalf("not ad hoc: %+v", o)
		}
	}
	if !res.Verified || !res.AdHoc {
		t.Fatalf("%+v", res)
	}
	spec.Verified(t, dist.RDistSignAdhoc)
}

func TestDeveloperIDSigning(t *testing.T) {
	ws := bundleWS(t)
	ws.put("app.entitlements", []byte("<plist/>"))
	s := &fakeSigner{ids: []Identity{devID, {Hash: "1111111111111111111111111111111111111111", Name: "Apple Development: Someone"}}}

	// By a unique part of the name.
	res, err := sign(ws, s, SignRequest{App: app, Identity: "Example GmbH", Entitlements: "app.entitlements"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Identity != devID.Name || !res.HardenedRuntime || !res.Timestamp {
		t.Fatalf("%+v", res)
	}
	for i, o := range s.opts {
		if o.Identity != devID.Hash || !o.HardenedRuntime || !o.Timestamp {
			t.Fatalf("%+v", o)
		}
		target := s.signed[i]
		withEnt := target == app || strings.HasSuffix(target, "MacOS/mytool")
		if (o.Entitlements != "") != withEnt {
			t.Errorf("entitlements on %s: %q", target, o.Entitlements)
		}
	}

	// By fingerprint, in any case.
	if _, err := sign(ws, &fakeSigner{ids: s.ids}, SignRequest{App: app, Identity: strings.ToLower(devID.Hash)}); err != nil {
		t.Fatal(err)
	}

	// Not found: the available ones are listed.
	_, err = sign(ws, &fakeSigner{ids: s.ids}, SignRequest{App: app, Identity: "Nobody"})
	wantCode(t, err, "WAP-IDENTITY-NOT-FOUND")
	if !strings.Contains(err.Error(), "") || !strings.Contains(problem.From(err).Messages[0].What, devID.Name) {
		t.Fatalf("available identities not listed: %v", err)
	}

	// Ambiguous.
	_, err = sign(ws, &fakeSigner{ids: s.ids}, SignRequest{App: app, Identity: "e"})
	wantCode(t, err, "WAP-IDENTITY-AMBIGUOUS")

	// Neither or both.
	_, err = sign(ws, &fakeSigner{}, SignRequest{App: app})
	wantCode(t, err, "WAP-SIGN-MODE")
	_, err = sign(ws, &fakeSigner{}, SignRequest{App: app, AdHoc: true, Identity: "Example"})
	wantCode(t, err, "WAP-SIGN-MODE")

	// The hardened runtime can be switched off.
	off := false
	s = &fakeSigner{ids: s.ids}
	if _, err := sign(ws, s, SignRequest{App: app, Identity: "Example", HardenedRuntime: &off}); err != nil || s.opts[0].HardenedRuntime {
		t.Fatalf("%v %+v", err, s.opts)
	}
	spec.Verified(t, dist.RDistSignIdentity)
}

func TestSigningOrderIsInsideOut(t *testing.T) {
	s := &fakeSigner{}
	res, err := sign(bundleWS(t), s, SignRequest{App: app, AdHoc: true})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"Contents/Frameworks/Bar.framework",
		"Contents/Frameworks/libfoo.dylib",
		"Contents/Helpers/helper",
		"Contents/MacOS/tool2",
		"Contents/MacOS/mytool",
		".",
	}
	if !slices.Equal(res.Order, want) {
		t.Fatalf("order %v", res.Order)
	}
	if s.signed[len(s.signed)-1] != app || s.signed[0] != filepath.Join(app, "Contents/Frameworks/Bar.framework") {
		t.Fatalf("signed %v", s.signed)
	}
	spec.Verified(t, dist.RDistSignOrder)
}

func TestSignatureIsVerified(t *testing.T) {
	s := &fakeSigner{verifyErr: errString("a sealed resource is missing or invalid")}
	_, err := sign(bundleWS(t), s, SignRequest{App: app, AdHoc: true})
	wantCode(t, err, "WAP-SIGN-VERIFY")
	if !strings.Contains(err.Error(), "sealed resource") {
		t.Fatalf("codesign's words are lost: %v", err)
	}
	spec.Verified(t, dist.RDistSignVerify)
}

type errString string

func (e errString) Error() string { return string(e) }

func notarizeWith(ws Workspace, s Signer, n Notary, req NotarizeRequest) (Notarized, error) {
	return NewNotarizeArtifact(ws, s, n)(auth.System(), req)
}

var goodSig = Signature{Signed: true, Authority: []string{devID.Name, "Developer ID Certification Authority", "Apple Root CA"}, HardenedRuntime: true, Timestamp: true}

func TestNotarization(t *testing.T) {
	ws := bundleWS(t)
	n := &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	res, err := notarizeWith(ws, &fakeSigner{sig: goodSig}, n, NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p"}, Staple: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "Accepted" || len(ws.zips) != 1 || !strings.HasSuffix(n.submitted[0], "My Tool.zip") {
		t.Fatalf("the bundle was not zipped for upload: %+v %v %v", res, ws.zips, n.submitted)
	}

	// A bundle that is not signed for notarization is refused before upload.
	for _, sig := range []Signature{{}, {Signed: true, AdHoc: true}, {Signed: true, Authority: []string{"Apple Development: X"}, HardenedRuntime: true, Timestamp: true},
		{Signed: true, Authority: goodSig.Authority, Timestamp: true}, {Signed: true, Authority: goodSig.Authority, HardenedRuntime: true}} {
		n := &fakeNotary{verdict: Submission{ID: "x", Status: "Accepted"}}
		_, err := notarizeWith(ws, &fakeSigner{sig: sig}, n, NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p"}})
		wantCode(t, err, "WAP-NOTARIZE-UNSIGNED")
		if len(n.submitted) != 0 {
			t.Fatal("an unsuitable bundle was uploaded")
		}
	}

	// Only .app, .dmg and .zip.
	ws.put("dist/x.pkg", []byte("x"))
	_, err = notarizeWith(ws, &fakeSigner{}, &fakeNotary{}, NotarizeRequest{Path: "dist/x.pkg", Credentials: Credentials{KeychainProfile: "p"}})
	wantCode(t, err, "WAP-NOTARIZE-FORMAT")

	// Exactly one way to reach the service.
	_, err = notarizeWith(ws, &fakeSigner{}, &fakeNotary{}, NotarizeRequest{Path: app})
	wantCode(t, err, "WAP-NOTARIZE-CREDENTIALS")
	_, err = notarizeWith(ws, &fakeSigner{}, &fakeNotary{}, NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p", AppleID: "a"}})
	wantCode(t, err, "WAP-NOTARIZE-CREDENTIALS")

	// A dry run sends nothing.
	n = &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	res, err = notarizeWith(ws, &fakeSigner{sig: goodSig}, n, NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p"}, Staple: true, DryRun: true})
	if err != nil || len(n.submitted) != 0 || len(res.Steps) != 3 {
		t.Fatalf("%v %v %+v", err, n.submitted, res)
	}
	spec.Verified(t, dist.RDistNotarize, guide.RGuideDryRun)
}

func TestPasswordsComeFromTheEnvironment(t *testing.T) {
	ws := bundleWS(t)
	req := NotarizeRequest{Path: app, Credentials: Credentials{AppleID: "dev@example.com", TeamID: "TEAM"}}
	_, err := notarizeWith(ws, &fakeSigner{sig: goodSig}, &fakeNotary{}, req)
	wantCode(t, err, "WAP-NOTARIZE-CREDENTIALS")

	req.PasswordEnv = "AC_PASSWORD"
	_, err = notarizeWith(ws, &fakeSigner{sig: goodSig}, &fakeNotary{}, req)
	wantCode(t, err, "WAP-NOTARIZE-PASSWORD")
	if !strings.Contains(err.Error(), "AC_PASSWORD") {
		t.Fatal("the variable is not named")
	}

	req.Credentials.Password = "secret"
	n := &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	if _, err := notarizeWith(ws, &fakeSigner{sig: goodSig}, n, req); err != nil {
		t.Fatal(err)
	}
	if n.creds[0].Password != "secret" {
		t.Fatal("the password did not reach the service")
	}
	spec.Verified(t, dist.RDistNotarizeSecrets)
}

func TestRejectionsShowApplesReasons(t *testing.T) {
	log, _ := json.Marshal(map[string]any{
		"statusSummary": "Archive contains critical validation errors",
		"issues": []map[string]string{
			{"severity": "error", "path": "My Tool.zip/My Tool.app/Contents/MacOS/mytool", "message": "The executable does not have the hardened runtime enabled.", "architecture": "arm64"},
		},
	})
	n := &fakeNotary{verdict: Submission{ID: "abc", Status: "Invalid"}, log: log}
	_, err := notarizeWith(bundleWS(t), &fakeSigner{sig: goodSig}, n, NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p"}, Staple: true})
	wantCode(t, err, "WAP-NOTARIZE-REJECTED")
	what := problem.From(err).Messages[0].What
	if !strings.Contains(what, "hardened runtime") || !strings.Contains(what, "critical validation errors") {
		t.Fatalf("reasons missing: %s", what)
	}
	if len(n.stapled) != 0 {
		t.Fatal("a rejected submission was stapled")
	}
	spec.Verified(t, dist.RDistNotarizeLog)
}

func TestStapling(t *testing.T) {
	ws := bundleWS(t)
	n := &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	res, err := notarizeWith(ws, &fakeSigner{sig: goodSig}, n, NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p"}, Staple: true})
	if err != nil || !res.Stapled || n.stapled[0] != app {
		t.Fatalf("%v %+v %v", err, res, n.stapled)
	}

	ws.put("dist/My Tool.dmg", []byte("dmg"))
	n = &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	if _, err := notarizeWith(ws, &fakeSigner{}, n, NotarizeRequest{Path: "dist/My Tool.dmg", Credentials: Credentials{KeychainProfile: "p"}, Staple: true}); err != nil || n.stapled[0] != "dist/My Tool.dmg" {
		t.Fatalf("%v %v", err, n.stapled)
	}

	ws.put("dist/My Tool.zip", []byte("zip"))
	n = &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	res, err = notarizeWith(ws, &fakeSigner{}, n, NotarizeRequest{Path: "dist/My Tool.zip", Credentials: Credentials{KeychainProfile: "p"}, Staple: true})
	if err != nil || len(n.stapled) != 0 || res.Notes[0].Code != "WAP-STAPLE-ZIP" {
		t.Fatalf("%v %v %+v", err, n.stapled, res)
	}
	spec.Verified(t, dist.RDistStaple)
}

func TestDiskImage(t *testing.T) {
	ws := bundleWS(t)
	img := &fakeImager{}
	s := &fakeSigner{ids: []Identity{devID}}
	res, err := NewCreateDiskImage(ws, img, s)(auth.System(), DiskImageRequest{App: app, Identity: "Example"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Path != "dist/MyTool-1.2.0.dmg" || img.specs[0].Volume != "My Tool" || img.specs[0].App != app {
		t.Fatalf("%+v %+v", res, img.specs)
	}
	if s.signed[0] != res.Path || !s.opts[0].Timestamp {
		t.Fatal("the image was not signed")
	}
	ws.put(res.Path, []byte("dmg"))
	_, err = NewCreateDiskImage(ws, img, s)(auth.System(), DiskImageRequest{App: app})
	wantCode(t, err, "WAP-ARTIFACT-EXISTS")
	spec.Verified(t, dist.RDistDmg)
}

func TestZipArchive(t *testing.T) {
	ws := bundleWS(t)
	res, err := NewCreateArchive(ws)(auth.System(), ArchiveRequest{App: app})
	if err != nil || res.Path != "dist/MyTool-1.2.0.zip" || ws.zips[0] != app+" -> dist/MyTool-1.2.0.zip" {
		t.Fatalf("%v %+v %v", err, res, ws.zips)
	}
	_, err = NewCreateArchive(ws)(auth.System(), ArchiveRequest{App: app})
	wantCode(t, err, "WAP-ARTIFACT-EXISTS")
	_, err = NewCreateArchive(ws)(auth.System(), ArchiveRequest{App: "dist/nothing.app"})
	wantCode(t, err, "WAP-PATH-NOT-FOUND")
	spec.Verified(t, dist.RDistZip)
}

func TestWhereAppleToolsAreMissingItSaysSo(t *testing.T) {
	ws := bundleWS(t)
	linux := &Unavailable{Tool: "codesign", OS: "linux"}
	_, err := sign(ws, &fakeSigner{unavailable: linux}, SignRequest{App: app, AdHoc: true})
	wantCode(t, err, "WAP-PLATFORM-MACOS")
	if exitKind(err) != problem.Unavailable {
		t.Fatal("not reported as unavailable")
	}
	if !strings.Contains(problem.From(err).Messages[0].How, "Mac") {
		t.Fatal("does not say where to run it")
	}

	missing := &Unavailable{Tool: "hdiutil"}
	_, err = NewCreateDiskImage(ws, &fakeImager{unavailable: missing}, &fakeSigner{})(auth.System(), DiskImageRequest{App: app})
	wantCode(t, err, "WAP-TOOL-MISSING")

	_, err = notarizeWith(ws, &fakeSigner{}, &fakeNotary{unavailable: &Unavailable{Tool: "xcrun notarytool", OS: "windows"}},
		NotarizeRequest{Path: app, Credentials: Credentials{KeychainProfile: "p"}})
	wantCode(t, err, "WAP-PLATFORM-MACOS")

	// A zip archive needs nothing from Apple.
	if _, err := NewCreateArchive(ws)(auth.System(), ArchiveRequest{App: app}); err != nil {
		t.Fatal(err)
	}
	spec.Verified(t, dist.RDistPlatform)
}

func TestSettingsFromTheConfigurationFile(t *testing.T) {
	ws := newWS()
	ws.put("proj/wdyapppack.json", []byte(`{
  "name": "ignored here",
  "sign": {"identity": "${ID}", "entitlements": "app.entitlements", "hardenedRuntime": true},
  "notarize": {"keychainProfile": "p"},
  "dmg": {"volumeName": "Vol", "sign": true}
}`))
	load := NewLoadSettings(ws)
	lookup := func(k string) (string, bool) {
		if k == "ID" {
			return "Example", true
		}
		return "", false
	}
	got, err := load(auth.System(), LoadSettingsRequest{Path: "proj/wdyapppack.json", Lookup: lookup})
	if err != nil {
		t.Fatal(err)
	}
	s := got.Settings
	if s.Sign.Identity != "Example" || s.Sign.Entitlements != "proj/app.entitlements" || !*s.Sign.HardenedRuntime || s.Notarize.KeychainProfile != "p" || !s.DMG.Sign {
		t.Fatalf("%+v", s)
	}
	_, err = load(auth.System(), LoadSettingsRequest{Path: "proj/wdyapppack.json"})
	wantCode(t, err, "WAP-SETTINGS-ENV-UNSET")

	ws.put("bad.json", []byte(`{"sign": {"identty": "x"}}`))
	_, err = load(auth.System(), LoadSettingsRequest{Path: "bad.json"})
	wantCode(t, err, "WAP-SETTINGS-INVALID")
	if !strings.Contains(err.Error(), "") || !strings.Contains(problem.From(err).Messages[0].How, `"identity"`) {
		t.Fatalf("no suggestion: %v", problem.From(err).Messages[0].How)
	}
	if got, err := load(auth.System(), LoadSettingsRequest{Optional: true}); err != nil || got.Found {
		t.Fatalf("%v %v", got, err)
	}
	spec.Verified(t, config.RConfigFile, config.RConfigStrict, config.RConfigEnv)
}

func TestEveryMessageCodeIsExplained(t *testing.T) {
	for code, e := range explanations() {
		if e.summary == "" || e.why == "" || e.how == "" || e.see == "" {
			t.Errorf("%s is not fully explained", code)
		}
	}
}
