// Package codesign implements the distribution context's Signer port with
// Apple's codesign and security tools.
package codesign

import (
	"os/exec"
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

// Identities lists valid code signing identities.
func (s Signer) Identities() ([]distribution.Identity, error) {
	res, err := s.run.Run(nil, s.tools.Security, "find-identity", "-v", "-p", "codesigning")
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
