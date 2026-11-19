package distribution

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// SignRequest asks for a bundle to be signed.
type SignRequest struct {
	App string
	// AdHoc signs without a certificate.
	AdHoc bool
	// Identity is a certificate name, a unique part of it, a fingerprint, or
	// "-" for ad hoc.
	Identity     string
	Entitlements string
	// HardenedRuntime overrides the default: on with a Developer ID, off ad hoc.
	HardenedRuntime *bool
	// Certificate, when set, is imported into a temporary keychain that
	// lives as long as this request; the identity must be in it.
	Certificate *Certificate
	DryRun      bool
}

// Signed describes what was signed.
type Signed struct {
	Path            string `json:"path"`
	Identity        string `json:"identity"`
	AdHoc           bool   `json:"adHoc"`
	HardenedRuntime bool   `json:"hardenedRuntime"`
	Timestamp       bool   `json:"timestamp"`
	// Order lists what was signed, in the order it was signed, relative to
	// the bundle.
	Order    []string `json:"order"`
	Verified bool     `json:"verified"`
	// TemporaryKeychain reports that the certificate came from the
	// environment and was signed with from a keychain of its own.
	TemporaryKeychain bool `json:"temporaryKeychain"`
	DryRun            bool `json:"dryRun"`
	Report
}

var fingerprint = regexp.MustCompile(`^[0-9A-Fa-f]{40}$`)

func signBundle(ws Workspace, signer Signer, req SignRequest) (res Signed, err error) {
	adhoc := req.AdHoc || req.Identity == "-"
	if adhoc == (req.Identity != "" && req.Identity != "-") || (!adhoc && req.Identity == "") {
		return Signed{}, problem.Misuse(msg("WAP-SIGN-MODE", "Give either --adhoc or --identity."))
	}
	b, err := readBundle(ws, req.App)
	if err != nil {
		return Signed{}, err
	}
	if req.Entitlements != "" {
		if st, err := ws.Stat(req.Entitlements); err != nil || !st.Exists {
			return Signed{}, problem.Fail(msg("WAP-ENTITLEMENTS-MISSING", fmt.Sprintf("%s does not exist.", req.Entitlements)))
		}
	}
	if err := ready(signer); err != nil {
		return Signed{}, err
	}

	var notes problem.Collector
	opts := SignOptions{Identity: "-", Entitlements: req.Entitlements}
	res = Signed{Path: req.App, Identity: "-", AdHoc: adhoc, DryRun: req.DryRun}
	if !adhoc {
		kc, closeKC, err := openKeychain(ws, signer, req.Certificate, &notes)
		if err != nil {
			return Signed{}, err
		}
		defer func() { closeKC(); res.Notes = notes.Notes() }()
		id, err := resolveIdentity(signer, kc, req.Identity, req.Certificate)
		if err != nil {
			return Signed{}, err
		}
		opts.Identity, opts.HardenedRuntime, opts.Timestamp, opts.Keychain = id.Hash, true, true, kc.Path
		res.Identity, res.TemporaryKeychain = id.Name, kc.Path != ""
	}
	if req.HardenedRuntime != nil {
		opts.HardenedRuntime = *req.HardenedRuntime
	}
	res.HardenedRuntime, res.Timestamp = opts.HardenedRuntime, opts.Timestamp

	order, err := signingOrder(ws, b)
	if err != nil {
		return Signed{}, err
	}
	res.Order = order
	if req.DryRun {
		return res, nil
	}

	for _, rel := range order {
		o := opts
		target := req.App
		if rel != "." {
			target = filepath.Join(req.App, filepath.FromSlash(rel))
		}
		// Entitlements belong to the program and the bundle, not to the
		// libraries and helpers inside it.
		if rel != "." && rel != "Contents/MacOS/"+b.executable {
			o.Entitlements = ""
		}
		if err := signer.Sign(target, o); err != nil {
			return Signed{}, problem.Fail(msg("WAP-SIGN-FAILED", fmt.Sprintf("Signing %s failed: %v", target, strings.TrimSpace(err.Error()))))
		}
	}
	if err := signer.Verify(req.App); err != nil {
		return Signed{}, problem.Fail(msg("WAP-SIGN-VERIFY", fmt.Sprintf("The signature of %s does not verify: %v", req.App, strings.TrimSpace(err.Error()))))
	}
	res.Verified = true
	return res, nil
}

// resolveIdentity finds the identity meant in a keychain: by fingerprint, by
// exact name, or by a part of a name that only one identity contains.

func resolveIdentity(signer Signer, kc Keychain, want string, cert *Certificate) (Identity, error) {
	ids, err := signer.Identities(kc)
	if err != nil {
		return Identity{}, problem.Fail(msg("WAP-SIGN-FAILED", fmt.Sprintf("The signing identities cannot be listed: %v", err)))
	}
	var matches []Identity
	for _, id := range ids {
		switch {
		case fingerprint.MatchString(want) && strings.EqualFold(id.Hash, want):
			return id, nil
		case id.Name == want:
			return id, nil
		case strings.Contains(strings.ToLower(id.Name), strings.ToLower(want)):
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	list := func(ids []Identity) string {
		if len(ids) == 0 {
			return "There is no signing identity in the keychain."
		}
		var b strings.Builder
		b.WriteString("Available:")
		for _, id := range ids {
			fmt.Fprintf(&b, " %s (%s);", id.Name, id.Hash)
		}
		return strings.TrimSuffix(b.String(), ";") + "."
	}
	if len(matches) > 1 {
		return Identity{}, problem.Misuse(msg("WAP-IDENTITY-AMBIGUOUS", fmt.Sprintf("%q matches %d identities. %s", want, len(matches), list(matches))))
	}
	if kc.Path != "" {
		found := list(ids)
		if len(ids) == 0 {
			found = "It contains no valid signing identity: it has to be a Developer ID Application certificate exported together with its private key."
		}
		source := EnvCertificate
		if cert != nil && cert.File != "" {
			source = cert.File
		}
		return Identity{}, problem.Fail(msg("WAP-IDENTITY-NOT-FOUND", fmt.Sprintf("No signing identity in the certificate in %s matches %q. %s", source, want, found)))
	}
	return Identity{}, problem.Fail(msg("WAP-IDENTITY-NOT-FOUND", fmt.Sprintf("No signing identity in the keychain matches %q. %s", want, list(ids))))
}

// signingOrder lists the code in a bundle from the inside out: libraries,
// then helpers, then further programs, then the program, then the bundle.
// Code must be signed before what contains it, because signing the container
// seals the signatures inside.
func signingOrder(ws Workspace, b bundleInfo) ([]string, error) {
	list, err := ws.List(filepath.Join(b.path, "Contents"))
	if err != nil {
		return nil, problem.Fail(msg("WAP-NOT-A-BUNDLE", fmt.Sprintf("%s cannot be listed: %v.", b.path, err)))
	}
	var frameworks, helpers, programs []string
	for _, f := range list {
		dir, name := path.Split(f.Path)
		switch dir {
		case "Frameworks/":
			if f.Dir && !strings.HasSuffix(name, ".framework") && !strings.HasSuffix(name, ".bundle") {
				continue
			}
			frameworks = append(frameworks, "Contents/"+f.Path)
		case "Helpers/":
			if f.Dir && !strings.HasSuffix(name, ".app") {
				continue
			}
			helpers = append(helpers, "Contents/"+f.Path)
		case "MacOS/":
			if !f.Dir && name != b.executable {
				programs = append(programs, "Contents/"+f.Path)
			}
		}
	}
	slices.Sort(frameworks)
	slices.Sort(helpers)
	slices.Sort(programs)
	order := append(append(append(frameworks, helpers...), programs...), "Contents/MacOS/"+b.executable, ".")
	return order, nil
}
