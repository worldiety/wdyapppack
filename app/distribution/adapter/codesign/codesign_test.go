package codesign

import (
	"errors"
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
	"github.com/worldiety/wdyapppack/requirements/cst"
)

func fakes(t *testing.T) (Signer, runtest.Tool, runtest.Tool) {
	dir := t.TempDir()
	cs := runtest.Fake(t, dir, "codesign")
	sec := runtest.Fake(t, dir, "security")
	return New(&run.Runner{}, Tools{Codesign: cs.Path, Security: sec.Path, OS: "darwin"}), cs, sec
}

// Signing is done by Apple's codesign, with exactly the options asked for.
func TestCodesignIsAskedToSign(t *testing.T) {
	s, cs, _ := fakes(t)
	if err := s.Ready(); err != nil {
		t.Fatal(err)
	}
	if err := s.Sign("My.app", distribution.SignOptions{Identity: "ABC", HardenedRuntime: true, Timestamp: true, Entitlements: "e.plist"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Sign("lib.dylib", distribution.SignOptions{Identity: "-"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Verify("My.app"); err != nil {
		t.Fatal(err)
	}
	calls := cs.Calls(t)
	want := [][]string{
		{"--force", "--sign", "ABC", "--options", "runtime", "--timestamp", "--entitlements", "e.plist", "My.app"},
		{"--force", "--sign", "-", "--timestamp=none", "lib.dylib"},
		{"--verify", "--deep", "--strict", "--verbose=2", "My.app"},
	}
	for i := range want {
		if !slices.Equal(calls[i], want[i]) {
			t.Errorf("call %d: %q, want %q", i, calls[i], want[i])
		}
	}
	spec.Verified(t, cst.RCstAppleToolsForSigning)
}

func TestIdentitiesAreRead(t *testing.T) {
	s, _, sec := fakes(t)
	sec.Says(t, `  1) ABCDEF0123456789ABCDEF0123456789ABCDEF01 "Developer ID Application: Example GmbH (TEAM123456)"
  2) 1111111111111111111111111111111111111111 "Apple Development: someone@example.com (ABC)"
     2 valid identities found
`)
	ids, err := s.Identities()
	if err != nil || len(ids) != 2 || ids[0].Name != "Developer ID Application: Example GmbH (TEAM123456)" || ids[1].Hash != "1111111111111111111111111111111111111111" {
		t.Fatalf("%+v %v", ids, err)
	}
}

func TestSignatureDescription(t *testing.T) {
	sig := parseDescription(`Executable=/x/My.app/Contents/MacOS/my
Identifier=com.example.my
Format=app bundle with Mach-O universal (x86_64 arm64)
CodeDirectory v=20500 size=1234 flags=0x10000(runtime) hashes=10+7 location=embedded
Authority=Developer ID Application: Example GmbH (TEAM123456)
Authority=Developer ID Certification Authority
Authority=Apple Root CA
Timestamp=1 Jan 2026 at 12:00:00
TeamIdentifier=TEAM123456
`)
	if !sig.Signed || sig.AdHoc || !sig.HardenedRuntime || !sig.Timestamp || sig.TeamID != "TEAM123456" || len(sig.Authority) != 3 {
		t.Fatalf("%+v", sig)
	}
	adhoc := parseDescription("CodeDirectory v=20400 size=1 flags=0x2(adhoc) hashes=1\nSignature=adhoc\nTeamIdentifier=not set\n")
	if !adhoc.AdHoc || adhoc.TeamID != "" || adhoc.HardenedRuntime {
		t.Fatalf("%+v", adhoc)
	}

	s, cs, _ := fakes(t)
	cs.Complains(t, "My.app: code object is not signed at all").Fails(t, 1)
	got, err := s.Describe("My.app")
	if err != nil || got.Signed {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestFailuresCarryCodesignsWords(t *testing.T) {
	s, cs, _ := fakes(t)
	cs.Complains(t, "errSecInternalComponent").Fails(t, 1)
	err := s.Sign("My.app", distribution.SignOptions{Identity: "-"})
	if err == nil || !strings.Contains(err.Error(), "errSecInternalComponent") {
		t.Fatalf("got %v", err)
	}
}

func TestNotReadyElsewhere(t *testing.T) {
	var u *distribution.Unavailable
	err := New(&run.Runner{}, Tools{Codesign: "codesign", Security: "security", OS: "linux"}).Ready()
	if !errors.As(err, &u) || u.OS != "linux" {
		t.Fatalf("got %v", err)
	}
	err = New(&run.Runner{}, Tools{Codesign: "/nonexistent/codesign", Security: "/nonexistent/security", OS: "darwin"}).Ready()
	if !errors.As(err, &u) || u.Install == "" {
		t.Fatalf("got %v", err)
	}
}

// On a Mac, the real codesign signs a real program ad hoc.
func TestRealAdHocSignature(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("codesign exists only on macOS")
	}
	if _, err := exec.LookPath("codesign"); err != nil {
		t.Skip("codesign not installed")
	}
	src, err := exec.LookPath("true")
	if err != nil {
		t.Skip(err)
	}
	data, err := os.ReadFile(src)
	if err != nil {
		t.Skip(err)
	}
	p := filepath.Join(t.TempDir(), "prog")
	if err := os.WriteFile(p, data, 0o755); err != nil {
		t.Fatal(err)
	}
	s := New(&run.Runner{}, DefaultTools())
	if err := s.Sign(p, distribution.SignOptions{Identity: "-"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Verify(p); err != nil {
		t.Fatal(err)
	}
	sig, err := s.Describe(p)
	if err != nil || !sig.Signed || !sig.AdHoc {
		t.Fatalf("%+v %v", sig, err)
	}
}
