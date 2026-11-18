package distribution

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// NotarizeRequest asks for a bundle, disk image or archive to be notarized.
type NotarizeRequest struct {
	Path        string
	Credentials Credentials
	// PasswordEnv names the variable the password was taken from, for
	// messages; the password itself is in Credentials.
	PasswordEnv string
	// Staple attaches the ticket afterwards.
	Staple bool
	DryRun bool
}

// Notarized describes the outcome.
type Notarized struct {
	Path         string `json:"path"`
	SubmissionID string `json:"submissionId,omitempty"`
	Status       string `json:"status,omitempty"`
	Stapled      bool   `json:"stapled"`
	DryRun       bool   `json:"dryRun"`
	// Steps describes what is done, in order.
	Steps []string `json:"steps"`
	Report
}

func notarize(ws Workspace, signer Signer, notary Notary, req NotarizeRequest) (Notarized, error) {
	if err := checkCredentials(req); err != nil {
		return Notarized{}, err
	}
	kind := strings.ToLower(filepath.Ext(strings.TrimRight(req.Path, "/")))
	if kind != ".app" && kind != ".dmg" && kind != ".zip" {
		return Notarized{}, problem.Misuse(msg("WAP-NOTARIZE-FORMAT", fmt.Sprintf("%s is neither a .app, a .dmg nor a .zip.", req.Path)))
	}
	if st, err := ws.Stat(req.Path); err != nil || !st.Exists {
		return Notarized{}, problem.Fail(msg("WAP-PATH-NOT-FOUND", fmt.Sprintf("%s does not exist.", req.Path)))
	}
	if kind == ".app" {
		if _, err := readBundle(ws, req.Path); err != nil {
			return Notarized{}, err
		}
	}
	if err := ready(signer, notary); err != nil {
		return Notarized{}, err
	}

	var notes problem.Collector
	if kind == ".app" {
		sig, err := signer.Describe(req.Path)
		if err != nil {
			return Notarized{}, problem.Fail(msg("WAP-NOTARIZE-UNSIGNED", fmt.Sprintf("The signature of %s cannot be read: %v", req.Path, err)))
		}
		if reason := unsuitable(sig); reason != "" {
			return Notarized{}, problem.Fail(msg("WAP-NOTARIZE-UNSIGNED", fmt.Sprintf("%s %s.", req.Path, reason)))
		}
	}
	staple := req.Staple && kind != ".zip"
	if req.Staple && kind == ".zip" {
		notes.Note(msg("WAP-STAPLE-ZIP", fmt.Sprintf("%s is a zip archive; no ticket is stapled.", req.Path)))
	}

	res := Notarized{Path: req.Path, DryRun: req.DryRun, Report: Report{Notes: notes.Notes()}}
	upload := req.Path
	if kind == ".app" {
		res.Steps = append(res.Steps, "zip "+req.Path+" for the upload")
	}
	res.Steps = append(res.Steps, "submit to Apple's notary service and wait for the verdict (usually a few minutes)")
	if staple {
		res.Steps = append(res.Steps, "staple the ticket to "+req.Path)
	}
	if req.DryRun {
		return res, nil
	}

	if kind == ".app" {
		dir, cleanup, err := ws.TempDir()
		if err != nil {
			return Notarized{}, problem.Fail(msg("WAP-ZIP-FAILED", fmt.Sprintf("No scratch directory: %v.", err)))
		}
		defer cleanup()
		upload = filepath.Join(dir, strings.TrimSuffix(filepath.Base(req.Path), ".app")+".zip")
		if err := ws.Zip(req.Path, upload, ZipOptions{}); err != nil {
			return Notarized{}, problem.Fail(msg("WAP-ZIP-FAILED", fmt.Sprintf("%s could not be zipped for the upload: %v.", req.Path, err)))
		}
	}

	sub, err := notary.Submit(upload, req.Credentials)
	if err != nil {
		return Notarized{}, problem.Fail(msg("WAP-NOTARIZE-FAILED", fmt.Sprintf("The submission of %s failed: %v", req.Path, strings.TrimSpace(err.Error()))))
	}
	res.SubmissionID, res.Status = sub.ID, sub.Status
	if !strings.EqualFold(sub.Status, "Accepted") {
		return Notarized{}, rejected(notary, req, sub)
	}
	if staple {
		if err := notary.Staple(req.Path); err != nil {
			notes.Note(msg("WAP-STAPLE-FAILED", fmt.Sprintf("Stapling %s failed: %v", req.Path, strings.TrimSpace(err.Error()))))
		} else {
			res.Stapled = true
		}
	}
	res.Notes = notes.Notes()
	return res, nil
}

// unsuitable says why a signature will be rejected, "" if it will not.
func unsuitable(sig Signature) string {
	switch {
	case !sig.Signed:
		return "is not signed"
	case sig.AdHoc:
		return "is signed ad hoc, and Apple notarizes only code signed with a Developer ID"
	case !developerID(sig):
		return "is not signed with a Developer ID Application certificate (signed by: " + strings.Join(sig.Authority, ", ") + ")"
	case !sig.HardenedRuntime:
		return "is signed without the hardened runtime"
	case !sig.Timestamp:
		return "is signed without a secure timestamp"
	}
	return ""
}

func developerID(sig Signature) bool {
	for _, a := range sig.Authority {
		if strings.HasPrefix(a, "Developer ID Application") {
			return true
		}
	}
	return false
}

func checkCredentials(req NotarizeRequest) error {
	c := req.Credentials
	ways := 0
	if c.KeychainProfile != "" {
		ways++
	}
	if c.Key != "" || c.KeyID != "" {
		ways++
	}
	if c.AppleID != "" || c.TeamID != "" {
		ways++
	}
	switch {
	case ways == 0:
		return problem.Misuse(msg("WAP-NOTARIZE-CREDENTIALS", "No way to reach the notary service was given."))
	case ways > 1:
		return problem.Misuse(msg("WAP-NOTARIZE-CREDENTIALS", "More than one way to reach the notary service was given."))
	case (c.Key != "") != (c.KeyID != ""):
		return problem.Misuse(msgHow("WAP-NOTARIZE-CREDENTIALS", "An API key needs both --key and --key-id.", "Give the .p8 file with --key and its ID with --key-id; add --issuer for a team key."))
	case c.AppleID != "" && (c.TeamID == "" || req.PasswordEnv == ""):
		return problem.Misuse(msgHow("WAP-NOTARIZE-CREDENTIALS", "An Apple ID needs --team-id and --password-env as well.", "Give --apple-id, --team-id and --password-env NAME, with the app specific password in the environment variable NAME."))
	case c.AppleID != "" && c.Password == "":
		return problem.Misuse(msg("WAP-NOTARIZE-PASSWORD", fmt.Sprintf("The environment variable %s is not set.", req.PasswordEnv)))
	}
	return nil
}

// rejected fetches Apple's log and turns its issues into the message.
func rejected(notary Notary, req NotarizeRequest, sub Submission) error {
	what := fmt.Sprintf("Apple did not accept %s (submission %s, status %s).", req.Path, sub.ID, sub.Status)
	data, err := notary.Log(sub.ID, req.Credentials)
	if err != nil {
		return problem.Fail(msgHow("WAP-NOTARIZE-REJECTED", what+" The log could not be fetched: "+err.Error(),
			"Fetch it yourself with: xcrun notarytool log "+sub.ID+" <credentials>"))
	}
	var log struct {
		StatusSummary string `json:"statusSummary"`
		Issues        []struct {
			Severity     string `json:"severity"`
			Path         string `json:"path"`
			Message      string `json:"message"`
			Architecture string `json:"architecture"`
		} `json:"issues"`
	}
	_ = json.Unmarshal(data, &log)
	var b strings.Builder
	b.WriteString(what)
	if log.StatusSummary != "" {
		b.WriteString(" " + strings.TrimSuffix(log.StatusSummary, ".") + ".")
	}
	for _, i := range log.Issues {
		fmt.Fprintf(&b, "\n  - %s: %s", i.Path, i.Message)
		if i.Architecture != "" {
			fmt.Fprintf(&b, " (%s)", i.Architecture)
		}
	}
	if len(log.Issues) == 0 && len(data) > 0 && log.StatusSummary == "" {
		fmt.Fprintf(&b, "\n%s", strings.TrimSpace(string(data)))
	}
	return problem.Fail(msg("WAP-NOTARIZE-REJECTED", b.String()))
}
