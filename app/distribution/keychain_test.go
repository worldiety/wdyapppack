package distribution

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
)

func cert(p12 string) *Certificate {
	return &Certificate{Base64: base64.StdEncoding.EncodeToString([]byte(p12)), Password: "pw", HasPassword: true}
}

// On a build server, sign imports the certificate from the environment into a
// keychain of its own, signs from it, and deletes it again — also when it
// fails.
func TestSigningFromACertificateInTheEnvironment(t *testing.T) {
	ws := bundleWS(t)
	s := &fakeSigner{certIDs: []Identity{devID}}
	res, err := sign(ws, s, SignRequest{App: app, Identity: "Example GmbH", Certificate: cert("P12DATA")})
	if err != nil {
		t.Fatal(err)
	}
	if s.opened != 1 || s.closed != 1 || string(s.imported[0]) != "P12DATA" || s.passwords[0] != "pw" {
		t.Fatalf("opened %d closed %d imported %q", s.opened, s.closed, s.imported)
	}
	for _, o := range s.opts {
		if o.Keychain != "/tmp/kc/signing.keychain-db" {
			t.Fatalf("not signed from the temporary keychain: %+v", o)
		}
	}
	if !res.TemporaryKeychain {
		t.Fatal("the result does not say so")
	}

	// Base64 with line breaks, as base64 -i writes it, is accepted.
	wrapped := &Certificate{Base64: "UDEy\nREFUQQ==\n", Password: "", HasPassword: true}
	s = &fakeSigner{certIDs: []Identity{devID}}
	if _, err := sign(ws, s, SignRequest{App: app, Identity: "Example", Certificate: wrapped}); err != nil || string(s.imported[0]) != "P12DATA" {
		t.Fatalf("%v %q", err, s.imported)
	}

	// The identity must be in the certificate; the keychain is removed anyway.
	s = &fakeSigner{certIDs: []Identity{devID}, ids: []Identity{{Hash: "2222222222222222222222222222222222222222", Name: "Developer ID Application: Other"}}}
	_, err = sign(ws, s, SignRequest{App: app, Identity: "Other", Certificate: cert("P12")})
	wantCode(t, err, "WAP-IDENTITY-NOT-FOUND")
	if !strings.Contains(err.Error(), EnvCertificate) || s.closed != 1 {
		t.Fatalf("%v, closed %d", err, s.closed)
	}

	// A failed signature still removes the keychain.
	s = &fakeSigner{certIDs: []Identity{devID}, verifyErr: errString("broken")}
	_, _ = sign(ws, s, SignRequest{App: app, Identity: "Example", Certificate: cert("P12")})
	if s.opened != s.closed {
		t.Fatalf("opened %d, closed %d", s.opened, s.closed)
	}

	// Missing password, broken base64, a certificate that does not import.
	_, err = sign(ws, &fakeSigner{}, SignRequest{App: app, Identity: "Example", Certificate: &Certificate{Base64: "UDEy"}})
	wantCode(t, err, "WAP-CERTIFICATE-PASSWORD")
	_, err = sign(ws, &fakeSigner{}, SignRequest{App: app, Identity: "Example", Certificate: &Certificate{Base64: "%%%", HasPassword: true}})
	wantCode(t, err, "WAP-CERTIFICATE-INVALID")
	_, err = sign(ws, &fakeSigner{importErr: errString("MAC verification failed during PKCS12 import (wrong password?)")}, SignRequest{App: app, Identity: "Example", Certificate: cert("P12")})
	wantCode(t, err, "WAP-CERTIFICATE-INVALID")

	// At a terminal, the certificate is a file.
	ws.put("developer-id.p12", []byte("FROMFILE"))
	s = &fakeSigner{certIDs: []Identity{devID}}
	if _, err := sign(ws, s, SignRequest{App: app, Identity: "Example", Certificate: &Certificate{File: "developer-id.p12", Password: "typed", HasPassword: true}}); err != nil {
		t.Fatal(err)
	}
	if string(s.imported[0]) != "FROMFILE" || s.passwords[0] != "typed" || s.closed != 1 {
		t.Fatalf("%q %q closed %d", s.imported, s.passwords, s.closed)
	}
	_, err = sign(ws, &fakeSigner{}, SignRequest{App: app, Identity: "Example", Certificate: &Certificate{File: "missing.p12", HasPassword: true}})
	wantCode(t, err, "WAP-CERTIFICATE-INVALID")
	_, err = sign(ws, &fakeSigner{}, SignRequest{App: app, Identity: "Example", Certificate: &Certificate{File: "developer-id.p12"}})
	wantCode(t, err, "WAP-CERTIFICATE-PASSWORD")
	if !strings.Contains(err.Error(), "developer-id.p12") {
		t.Fatalf("the source is not named: %v", err)
	}

	// Ad hoc needs no certificate, and does not touch one that is there.
	s = &fakeSigner{}
	if _, err := sign(ws, s, SignRequest{App: app, AdHoc: true, Certificate: cert("P12")}); err != nil || s.opened != 0 {
		t.Fatalf("%v, opened %d", err, s.opened)
	}

	// The disk image is signed from a keychain of its own too.
	s = &fakeSigner{certIDs: []Identity{devID}}
	if _, err := NewCreateDiskImage(ws, &writingImager{fakeImager: &fakeImager{}, ws: ws}, s)(auth.System(), DiskImageRequest{App: app, Identity: "Example", Certificate: cert("P12")}); err != nil {
		t.Fatal(err)
	}
	if s.opened != 1 || s.closed != 1 || s.opts[0].Keychain == "" {
		t.Fatalf("opened %d closed %d %+v", s.opened, s.closed, s.opts)
	}
	spec.Verified(t, dist.RDistTemporaryKeychain, dec.RDecTemporaryKeychain)
}

const pem = "-----BEGIN PRIVATE KEY-----\nMIGTAgEAMBMGByqGSM49\n-----END PRIVATE KEY-----\n"

func TestAPIKeyFromTheEnvironment(t *testing.T) {
	for name, data := range map[string]string{"text": pem, "base64": base64.StdEncoding.EncodeToString([]byte(pem))} {
		ws := bundleWS(t)
		n := &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
		res, err := notarizeWith(ws, &fakeSigner{sig: goodSig}, n, NotarizeRequest{
			Path: app, Staple: true, APIKey: &APIKey{Data: data, KeyID: "ABC123", Issuer: "iss"},
		})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		c := n.creds[0]
		if c.KeyID != "ABC123" || c.Issuer != "iss" || c.Key != "tmp/AuthKey_ABC123.p8" {
			t.Fatalf("%s: %+v", name, c)
		}
		var key writtenFile
		for _, w := range ws.written {
			if strings.HasSuffix(w.path, ".p8") {
				key = w
			}
		}
		if string(key.data) != pem || key.perm != 0o600 {
			t.Fatalf("%s: key written as %q with %o", name, key.data, key.perm)
		}
		if _, err := ws.Read(key.path); err == nil {
			t.Fatalf("%s: the key file outlived the command", name)
		}
		if !strings.Contains(res.Credentials, EnvAPIKey) || strings.Contains(res.Credentials, "PRIVATE") {
			t.Fatalf("%s: %q", name, res.Credentials)
		}
	}

	// Credentials given explicitly win; the environment is not even read.
	n := &fakeNotary{verdict: Submission{ID: "abc", Status: "Accepted"}}
	if _, err := notarizeWith(bundleWS(t), &fakeSigner{sig: goodSig}, n, NotarizeRequest{
		Path: app, Credentials: Credentials{KeychainProfile: "p"}, APIKey: &APIKey{Data: "garbage"},
	}); err != nil || n.creds[0].KeychainProfile != "p" {
		t.Fatalf("%v %+v", err, n.creds)
	}

	_, err := notarizeWith(bundleWS(t), &fakeSigner{sig: goodSig}, &fakeNotary{}, NotarizeRequest{Path: app, APIKey: &APIKey{Data: pem}})
	wantCode(t, err, "WAP-API-KEY-INVALID")
	_, err = notarizeWith(bundleWS(t), &fakeSigner{sig: goodSig}, &fakeNotary{}, NotarizeRequest{Path: app, APIKey: &APIKey{Data: "bm90IGEga2V5", KeyID: "X"}})
	wantCode(t, err, "WAP-API-KEY-INVALID")

	// A dry run writes no key.
	ws := bundleWS(t)
	if _, err := notarizeWith(ws, &fakeSigner{sig: goodSig}, &fakeNotary{}, NotarizeRequest{Path: app, APIKey: &APIKey{Data: pem, KeyID: "X"}, DryRun: true}); err != nil || len(ws.written) != 0 {
		t.Fatalf("%v %v", err, ws.written)
	}
	spec.Verified(t, dist.RDistApiKeyEnv, dist.RDistNotarizeSecrets)
}

func TestChecksumsNextToImagesAndArchives(t *testing.T) {
	ws := bundleWS(t)
	zipRes, err := NewCreateArchive(ws)(auth.System(), ArchiveRequest{App: app})
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("zip"))
	want := hex.EncodeToString(sum[:]) + "  MyTool-1.2.0.zip\n"
	if got, _ := ws.Read("dist/MyTool-1.2.0.zip.sha256"); string(got) != want || zipRes.SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("%q", got)
	}

	ws = bundleWS(t)
	img := &fakeImager{}
	imager := &writingImager{fakeImager: img, ws: ws}
	res, err := NewCreateDiskImage(ws, imager, &fakeSigner{})(auth.System(), DiskImageRequest{App: app})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := ws.Read(res.Checksum); !strings.HasSuffix(string(got), "  MyTool-1.2.0.dmg\n") {
		t.Fatalf("%q", got)
	}

	// An existing checksum file is not replaced unasked.
	ws = bundleWS(t)
	ws.put("dist/MyTool-1.2.0.zip.sha256", []byte("old"))
	_, err = NewCreateArchive(ws)(auth.System(), ArchiveRequest{App: app})
	wantCode(t, err, "WAP-ARTIFACT-EXISTS")
	spec.Verified(t, dist.RDistChecksum)
}

// writingImager writes the image into the workspace, as hdiutil would.
type writingImager struct {
	*fakeImager
	ws *memWS
}

func (w *writingImager) Create(s DiskImageSpec) error {
	w.ws.put(s.Out, []byte("dmg"))
	return w.fakeImager.Create(s)
}
