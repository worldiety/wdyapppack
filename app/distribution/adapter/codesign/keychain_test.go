package codesign

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/run"
	"github.com/worldiety/wdyapppack/foundation/run/runtest"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
)

// security is asked for exactly the steps a build server script takes, the
// passwords never show in the log, and everything is gone afterwards.
func TestTemporaryKeychainSteps(t *testing.T) {
	dir := t.TempDir()
	sec := runtest.Fake(t, dir, "security")
	cs := runtest.Fake(t, dir, "codesign")
	var log bytes.Buffer
	s := New(&run.Runner{Log: &log, Verbose: func() bool { return true }}, Tools{Codesign: cs.Path, Security: sec.Path, OS: "darwin"})

	kc, closeKC, err := s.OpenKeychain([]byte("P12"), "cert-secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(kc.Path, "signing.keychain-db") {
		t.Fatal(kc.Path)
	}
	if err := s.Sign("My.app", distribution.SignOptions{Identity: "ABC", Keychain: kc.Path}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Identities(kc); err != nil {
		t.Fatal(err)
	}
	// The fake does not create the keychain file; make it exist so that
	// cleanup has to delete it.
	if err := os.WriteFile(kc.Path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := closeKC(); err != nil {
		t.Fatal(err)
	}

	var verbs []string
	for _, c := range sec.Calls(t) {
		verbs = append(verbs, c[0])
	}
	want := []string{"create-keychain", "set-keychain-settings", "unlock-keychain", "import", "set-key-partition-list", "find-identity", "delete-keychain"}
	if !slices.Equal(verbs, want) {
		t.Fatalf("steps %v", verbs)
	}
	importCall := sec.Calls(t)[3]
	if !slices.Contains(importCall, "cert-secret") || !slices.Contains(importCall, "/usr/bin/codesign") {
		t.Fatalf("%q", importCall)
	}
	if c := cs.Calls(t)[0]; c[len(c)-3] != "--keychain" || c[len(c)-2] != kc.Path {
		t.Fatalf("codesign not told the keychain: %q", c)
	}
	if strings.Contains(log.String(), "cert-secret") {
		t.Fatalf("the certificate password was shown:\n%s", log.String())
	}
	if _, err := os.Stat(filepath.Dir(kc.Path)); !os.IsNotExist(err) {
		t.Fatal("the temporary directory with the certificate is still there")
	}
	spec.Verified(t, dist.RDistTemporaryKeychain)
}

func TestTemporaryKeychainIsRemovedOnInterrupt(t *testing.T) {
	dir := t.TempDir()
	sec := runtest.Fake(t, dir, "security")
	s := New(&run.Runner{}, Tools{Codesign: "codesign", Security: sec.Path, OS: "darwin"})
	kc, _, err := s.OpenKeychain([]byte("P12"), "")
	if err != nil {
		t.Fatal(err)
	}
	run.Interrupted()
	if _, err := os.Stat(filepath.Dir(kc.Path)); !os.IsNotExist(err) {
		t.Fatal("an interrupt left the certificate behind")
	}
}

func TestAFailedImportLeavesNothingBehind(t *testing.T) {
	dir := t.TempDir()
	sec := runtest.Fake(t, dir, "security").Complains(t, "security: SecKeychainItemImport: MAC verification failed during PKCS12 import (wrong password?)").Fails(t, 1)
	s := New(&run.Runner{}, Tools{Codesign: "codesign", Security: sec.Path, OS: "darwin"})
	before, _ := filepath.Glob(filepath.Join(os.TempDir(), "wdyapppack-keychain-*"))
	_, _, err := s.OpenKeychain([]byte("P12"), "s3cr3t-pw")
	if err == nil || !strings.Contains(err.Error(), "wrong password") {
		t.Fatalf("got %v", err)
	}
	after, _ := filepath.Glob(filepath.Join(os.TempDir(), "wdyapppack-keychain-*"))
	if len(after) > len(before) {
		t.Fatalf("left behind: %v", after)
	}
}

// On a Mac, a real temporary keychain takes a real certificate and is
// deleted again. The certificate is self-signed, so it is no valid signing
// identity; what is checked is the keychain, not the signature.
func TestRealTemporaryKeychain(t *testing.T) {
	if runtime.GOOS != "darwin" || testing.Short() {
		t.Skip("needs macOS")
	}
	openssl, err := exec.LookPath("openssl")
	if err != nil {
		t.Skip("openssl is needed to make a test certificate")
	}
	dir := t.TempDir()
	key, crt, p12 := filepath.Join(dir, "k.pem"), filepath.Join(dir, "c.pem"), filepath.Join(dir, "c.p12")
	gen := [][]string{
		{"req", "-x509", "-newkey", "rsa:2048", "-nodes", "-keyout", key, "-out", crt, "-days", "1",
			"-subj", "/CN=wdyapppack test", "-addext", "extendedKeyUsage=codeSigning"},
		{"pkcs12", "-export", "-inkey", key, "-in", crt, "-out", p12, "-passout", "pass:test", "-legacy"},
	}
	for _, args := range gen {
		if out, err := exec.Command(openssl, args...).CombinedOutput(); err != nil {
			// Older LibreSSL has no -legacy and writes the old format anyway.
			if args[0] == "pkcs12" {
				out, err = exec.Command(openssl, args[:len(args)-1]...).CombinedOutput()
			}
			if err != nil {
				t.Skipf("cannot make a test certificate: %s", out)
			}
		}
	}
	data, err := os.ReadFile(p12)
	if err != nil {
		t.Fatal(err)
	}

	s := New(&run.Runner{}, DefaultTools())
	kc, closeKC, err := s.OpenKeychain(data, "test")
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("security", "find-certificate", "-c", "wdyapppack test", kc.Path).CombinedOutput()
	if err != nil {
		closeKC()
		t.Fatalf("the certificate is not in the keychain: %s", out)
	}
	if err := closeKC(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(kc.Path); !os.IsNotExist(err) {
		t.Fatal("the keychain is still there")
	}
	list, _ := exec.Command("security", "list-keychains", "-d", "user").CombinedOutput()
	if strings.Contains(string(list), kc.Path) {
		t.Fatalf("the keychain is still in the search list:\n%s", list)
	}
}
