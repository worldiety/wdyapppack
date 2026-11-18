package distribution

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/confdoc"
	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/foundation/suggest"
)

// ConfigFields describes the sections of the configuration file this context
// reads.
func ConfigFields() []confdoc.Field {
	return []confdoc.Field{
		{Path: "sign", Kind: confdoc.Object, Description: "How sign signs the bundle."},
		{Path: "sign.identity", Kind: confdoc.String, Flag: "identity", Example: `"Developer ID Application: Example GmbH (TEAMID1234)"`,
			Description: "The signing identity: a Developer ID Application certificate's name (or a unique part of it) or its SHA-1 fingerprint, or \"-\" for an ad hoc signature. May come from the environment: \"${SIGN_IDENTITY}\"."},
		{Path: "sign.entitlements", Kind: confdoc.String, Flag: "entitlements",
			Description: "An entitlements property list applied to the program and the bundle, relative to the configuration file."},
		{Path: "sign.hardenedRuntime", Kind: confdoc.Boolean, Flag: "hardened-runtime", Default: "true with a Developer ID, false ad hoc",
			Description: "Sign with the hardened runtime, which notarization requires."},
		{Path: "notarize", Kind: confdoc.Object, Description: "How notarize reaches Apple's notary service. Give one way: keychainProfile, or key with keyId (and issuer), or appleId with teamId and passwordEnv."},
		{Path: "notarize.keychainProfile", Kind: confdoc.String, Flag: "keychain-profile", Example: `"wdyapppack"`,
			Description: "Credentials stored once with: xcrun notarytool store-credentials <profile>. The recommended way."},
		{Path: "notarize.key", Kind: confdoc.String, Flag: "key", Description: "An App Store Connect API key file (.p8), relative to the configuration file."},
		{Path: "notarize.keyId", Kind: confdoc.String, Flag: "key-id", Description: "The ID of the API key."},
		{Path: "notarize.issuer", Kind: confdoc.String, Flag: "issuer", Description: "The issuer ID of a team API key."},
		{Path: "notarize.appleId", Kind: confdoc.String, Flag: "apple-id", Description: "An Apple ID, used with teamId and passwordEnv."},
		{Path: "notarize.teamId", Kind: confdoc.String, Flag: "team-id", Description: "The developer team ID."},
		{Path: "notarize.passwordEnv", Kind: confdoc.String, Flag: "password-env", Example: `"AC_PASSWORD"`,
			Description: "The name of the environment variable holding the Apple ID's app specific password. The password itself never goes into the file."},
		{Path: "notarize.staple", Kind: confdoc.Boolean, Flag: "no-staple", Default: "true",
			Description: "Staple the ticket to the bundle or disk image after a successful notarization."},
		{Path: "dmg", Kind: confdoc.Object, Description: "How dmg makes the disk image."},
		{Path: "dmg.volumeName", Kind: confdoc.String, Flag: "volume-name", Default: "the bundle's name",
			Description: "The name shown in the Finder when the image is opened."},
		{Path: "dmg.sign", Kind: confdoc.Boolean, Flag: "sign-identity", Default: "false",
			Description: "Sign the disk image with the identity of the sign section."},
	}
}

// StarterSections are the sections init writes, so that a new configuration
// shows them.
func StarterSections() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		// hardenedRuntime is left out on purpose: its default follows the
		// identity, and a false written here would survive the switch to a
		// Developer ID and make notarization fail.
		"sign":     json.RawMessage(`{"identity": "-", "entitlements": ""}`),
		"notarize": json.RawMessage(`{"keychainProfile": ""}`),
		"dmg":      json.RawMessage(`{"volumeName": "", "sign": false}`),
	}
}

var envRef = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// LoadSettingsRequest names the configuration file.
type LoadSettingsRequest struct {
	Path     string
	Optional bool
	Lookup   func(string) (string, bool)
}

// LoadedSettings are the sections read.
type LoadedSettings struct {
	Settings Settings `json:"settings"`
	Path     string   `json:"path"`
	Found    bool     `json:"found"`
}

func loadSettings(ws Workspace, req LoadSettingsRequest) (LoadedSettings, error) {
	p := cmp.Or(req.Path, "wdyapppack.json")
	st, err := ws.Stat(p)
	if err == nil && !st.Exists {
		if req.Optional {
			return LoadedSettings{Path: p}, nil
		}
		return LoadedSettings{}, problem.Misuse(msg("WAP-SETTINGS-MISSING", fmt.Sprintf("The configuration file %s does not exist.", p)))
	}
	data, err := ws.Read(p)
	if err != nil {
		return LoadedSettings{}, problem.Misuse(msg("WAP-SETTINGS-INVALID", fmt.Sprintf("%s cannot be read: %v.", p, err)))
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return LoadedSettings{}, problem.Misuse(msg("WAP-SETTINGS-INVALID", fmt.Sprintf("%s is not valid JSON: %v.", p, err)))
	}

	var s Settings
	var c problem.Collector
	section := func(name string, into any) {
		raw, ok := top[name]
		if !ok || string(raw) == "null" {
			return
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(into); err != nil {
			how := "Fix the section; every field is listed by: wdyapppack help config"
			if f, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
				f = strings.Trim(f, `"`)
				var known []string
				for _, cf := range ConfigFields() {
					if strings.HasPrefix(cf.Path, name+".") {
						known = append(known, strings.TrimPrefix(cf.Path, name+"."))
					}
				}
				if sug, ok := suggest.Closest(f, known); ok {
					how = fmt.Sprintf("Did you mean %q? %s", sug, how)
				}
			}
			c.Error(problem.Usage, msgHow("WAP-SETTINGS-INVALID", fmt.Sprintf("The section %q of %s: %s.", name, p, strings.TrimPrefix(err.Error(), "json: ")), how))
		}
	}
	section("sign", &s.Sign)
	section("notarize", &s.Notarize)
	section("dmg", &s.DMG)
	if err := c.Err(); err != nil {
		return LoadedSettings{}, err
	}

	lookup := req.Lookup
	if lookup == nil {
		lookup = func(string) (string, bool) { return "", false }
	}
	for _, f := range []struct {
		name string
		p    *string
	}{
		{"sign.identity", &s.Sign.Identity}, {"sign.entitlements", &s.Sign.Entitlements},
		{"notarize.keychainProfile", &s.Notarize.KeychainProfile}, {"notarize.key", &s.Notarize.Key},
		{"notarize.keyId", &s.Notarize.KeyID}, {"notarize.issuer", &s.Notarize.Issuer},
		{"notarize.appleId", &s.Notarize.AppleID}, {"notarize.teamId", &s.Notarize.TeamID},
		{"notarize.passwordEnv", &s.Notarize.PasswordEnv}, {"dmg.volumeName", &s.DMG.VolumeName},
	} {
		*f.p = envRef.ReplaceAllStringFunc(*f.p, func(ref string) string {
			name := envRef.FindStringSubmatch(ref)[1]
			v, ok := lookup(name)
			if !ok {
				c.Error(problem.Usage, msg("WAP-SETTINGS-ENV-UNSET", fmt.Sprintf("The field %s refers to ${%s}, which is not set.", f.name, name)))
			}
			return v
		})
	}
	if err := c.Err(); err != nil {
		return LoadedSettings{}, err
	}

	dir := filepath.Dir(p)
	for _, f := range []*string{&s.Sign.Entitlements, &s.Notarize.Key} {
		if *f != "" && !filepath.IsAbs(*f) {
			*f = filepath.Join(dir, *f)
		}
	}
	return LoadedSettings{Settings: s, Path: p, Found: true}, nil
}
