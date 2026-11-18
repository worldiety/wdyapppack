// Package notarytool implements the distribution context's Notary port with
// Apple's notarytool and stapler, run through xcrun.
package notarytool

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/run"
)

// Tools names the programs; tests put fakes here.
type Tools struct {
	Xcrun string
	OS    string
}

// DefaultTools are Apple's tools on this machine.
func DefaultTools() Tools { return Tools{Xcrun: "xcrun", OS: runtime.GOOS} }

// Notary is notarytool.
type Notary struct {
	tools Tools
	run   *run.Runner
}

// New returns the notary.
func New(r *run.Runner, t Tools) Notary { return Notary{tools: t, run: r} }

// Ready reports whether notarytool can run here.
func (n Notary) Ready() error {
	if n.tools.OS != "darwin" {
		return &distribution.Unavailable{Tool: "xcrun notarytool", OS: n.tools.OS}
	}
	if _, err := exec.LookPath(n.tools.Xcrun); err != nil {
		return &distribution.Unavailable{Tool: "xcrun", Install: "Install Xcode's command line tools with: xcode-select --install"}
	}
	if _, err := n.run.Run(nil, n.tools.Xcrun, "--find", "notarytool"); err != nil {
		return &distribution.Unavailable{Tool: "notarytool", Install: "notarytool comes with Xcode 13 or later; install or update Xcode, or its command line tools with: xcode-select --install"}
	}
	return nil
}

func credentials(c distribution.Credentials) (args, secrets []string) {
	switch {
	case c.KeychainProfile != "":
		return []string{"--keychain-profile", c.KeychainProfile}, nil
	case c.Key != "":
		args = []string{"--key", c.Key, "--key-id", c.KeyID}
		if c.Issuer != "" {
			args = append(args, "--issuer", c.Issuer)
		}
		return args, nil
	}
	return []string{"--apple-id", c.AppleID, "--team-id", c.TeamID, "--password", c.Password}, []string{c.Password}
}

// Submit uploads a file and waits for the verdict.
func (n Notary) Submit(path string, c distribution.Credentials) (distribution.Submission, error) {
	cred, secrets := credentials(c)
	args := append([]string{"notarytool", "submit", path, "--wait", "--output-format", "json"}, cred...)
	res, err := n.run.Run(secrets, n.tools.Xcrun, args...)
	var sub distribution.Submission
	if jerr := json.Unmarshal(res.Stdout, &sub); jerr == nil && sub.ID != "" {
		// A rejected submission may end with an error code and still report
		// its verdict; the verdict is what counts.
		return sub, nil
	}
	if err != nil {
		return sub, err
	}
	return sub, fmt.Errorf("notarytool gave no verdict: %s", res.Stdout)
}

// Log fetches the service's report on a submission.
func (n Notary) Log(id string, c distribution.Credentials) ([]byte, error) {
	cred, secrets := credentials(c)
	res, err := n.run.Run(secrets, n.tools.Xcrun, append([]string{"notarytool", "log", id}, cred...)...)
	if err != nil {
		return nil, err
	}
	if len(res.Stdout) == 0 {
		return nil, errors.New("notarytool printed no log")
	}
	return res.Stdout, nil
}

// Staple attaches the ticket.
func (n Notary) Staple(path string) error {
	_, err := n.run.Run(nil, n.tools.Xcrun, "stapler", "staple", path)
	return err
}

var _ distribution.Notary = Notary{}
