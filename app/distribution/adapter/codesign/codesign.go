// Package codesign implements the distribution context's Signer port with
// Apple's codesign and security tools.
package codesign

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/run"
)

// Tools names the programs; tests put fakes here.
type Tools struct {
	Codesign string
	Security string
	// OS is the operating system the tools are expected on.
	OS string
}

// DefaultTools are Apple's tools on this machine.
func DefaultTools() Tools { return Tools{Codesign: "codesign", Security: "security", OS: runtime.GOOS} }

// Signer is codesign.
type Signer struct {
	tools Tools
	run   *run.Runner
}

// New returns the signer.
func New(r *run.Runner, t Tools) Signer { return Signer{tools: t, run: r} }

const install = "Install Xcode's command line tools with: xcode-select --install"

// Ready reports whether codesign can run here.
func (s Signer) Ready() error {
	if s.tools.OS != "darwin" {
		return &distribution.Unavailable{Tool: "codesign", OS: s.tools.OS}
	}
	for _, t := range []string{s.tools.Codesign, s.tools.Security} {
		if _, err := exec.LookPath(t); err != nil {
			return &distribution.Unavailable{Tool: t, Install: install}
		}
	}
	return nil
}

var identityLine = regexp.MustCompile(`^\s*\d+\)\s+([0-9A-Fa-f]{40})\s+"(.*)"\s*$`)

// OpenKeychain creates a keychain in a private temporary directory, imports
// the PKCS#12 certificate into it and allows codesign to use the key without
// asking, exactly as a build server script would. The keychain gets a random
// password nobody needs to know: it is deleted again when the function
// returned is called, or when the process is interrupted.
func (s Signer) OpenKeychain(p12 []byte, password string) (distribution.Keychain, func() error, error) {
	dir, err := os.MkdirTemp("", "wdyapppack-keychain-*")
	if err != nil {
		return distribution.Keychain{}, nil, err
	}
	secret := make([]byte, 24)
	if _, err := rand.Read(secret); err != nil {
		os.RemoveAll(dir)
		return distribution.Keychain{}, nil, err
	}
	kcPassword := hex.EncodeToString(secret)
	keychain := filepath.Join(dir, "signing.keychain-db")
	certFile := filepath.Join(dir, "certificate.p12")

	var once bool
	cleanup := func() error {
		if once {
			return nil
		}
		once = true
		var errs []error
		if _, err := os.Stat(keychain); err == nil {
			if _, err := s.run.Run(nil, s.tools.Security, "delete-keychain", keychain); err != nil {
				errs = append(errs, err)
			}
		}
		if err := os.RemoveAll(dir); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}
	unregister := run.OnInterrupt(func() { _ = cleanup() })
	closeKC := func() error { unregister(); return cleanup() }

	secrets := []string{kcPassword, password}
	steps := [][]string{
		{"create-keychain", "-p", kcPassword, keychain},
		{"set-keychain-settings", "-lut", "21600", keychain},
		{"unlock-keychain", "-p", kcPassword, keychain},
		{"import", certFile, "-k", keychain, "-P", password, "-T", "/usr/bin/codesign", "-T", "/usr/bin/security"},
		{"set-key-partition-list", "-S", "apple-tool:,apple:,codesign:", "-s", "-k", kcPassword, keychain},
	}
	if err := os.WriteFile(certFile, p12, 0o600); err != nil {
		_ = closeKC()
		return distribution.Keychain{}, nil, err
	}
	for _, args := range steps {
		if _, err := s.run.Run(secrets, s.tools.Security, args...); err != nil {
			_ = closeKC()
			return distribution.Keychain{}, nil, err
		}
	}
	// The certificate is in the keychain now; the file is not needed.
	_ = os.Remove(certFile)
	return distribution.Keychain{Path: keychain}, closeKC, nil
}

// Identities lists valid code signing identities, of one keychain or of the
// user's keychains.
func (s Signer) Identities(k distribution.Keychain) ([]distribution.Identity, error) {
	args := []string{"find-identity", "-v", "-p", "codesigning"}
	if k.Path != "" {
		args = append(args, k.Path)
	}
	res, err := s.run.Run(nil, s.tools.Security, args...)
	if err != nil {
		return nil, err
	}
	var out []distribution.Identity
	for _, l := range strings.Split(string(res.Stdout), "\n") {
		if m := identityLine.FindStringSubmatch(l); m != nil {
			out = append(out, distribution.Identity{Hash: strings.ToUpper(m[1]), Name: m[2]})
		}
	}
	return out, nil
}

// Sign signs one piece of code.
func (s Signer) Sign(path string, o distribution.SignOptions) error {
	args := []string{"--force", "--sign", o.Identity}
	if o.HardenedRuntime {
		args = append(args, "--options", "runtime")
	}
	if o.Timestamp {
		args = append(args, "--timestamp")
	} else {
		args = append(args, "--timestamp=none")
	}
	if o.Entitlements != "" {
		args = append(args, "--entitlements", o.Entitlements)
	}
	if o.Keychain != "" {
		args = append(args, "--keychain", o.Keychain)
	}
	args = append(args, path)
	_, err := s.run.Run(nil, s.tools.Codesign, args...)
	return err
}

// Verify checks a signature deeply and strictly.
func (s Signer) Verify(path string) error {
	_, err := s.run.Run(nil, s.tools.Codesign, "--verify", "--deep", "--strict", "--verbose=2", path)
	return err
}

// Describe reads a signature.
func (s Signer) Describe(path string) (distribution.Signature, error) {
	res, err := s.run.Run(nil, s.tools.Codesign, "--display", "--verbose=4", path)
	text := string(res.Stderr) + string(res.Stdout)
	if err != nil {
		if strings.Contains(text, "not signed at all") {
			return distribution.Signature{}, nil
		}
		return distribution.Signature{}, err
	}
	return parseDescription(text), nil
}

// parseDescription reads codesign's description of a signature.
func parseDescription(text string) distribution.Signature {
	sig := distribution.Signature{Signed: true}
	for _, l := range strings.Split(text, "\n") {
		k, v, _ := strings.Cut(strings.TrimSpace(l), "=")
		switch k {
		case "Authority":
			sig.Authority = append(sig.Authority, v)
		case "TeamIdentifier":
			if v != "not set" {
				sig.TeamID = v
			}
		case "Signature":
			if v == "adhoc" {
				sig.AdHoc = true
			}
		case "Timestamp":
			sig.Timestamp = true
		}
		if strings.Contains(l, "flags=") && strings.Contains(l, "runtime") {
			sig.HardenedRuntime = true
		}
		if strings.Contains(l, "flags=") && strings.Contains(l, "adhoc") {
			sig.AdHoc = true
		}
	}
	return sig
}

var _ distribution.Signer = Signer{}
