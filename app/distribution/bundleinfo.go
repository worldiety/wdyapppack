package distribution

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/plist"
)

// bundleInfo is what this context needs to know about a bundle.
type bundleInfo struct {
	path, name, version, identifier, executable string
}

// readBundle checks that a path is an application bundle and reads its
// metadata.
func readBundle(ws Workspace, app string) (bundleInfo, error) {
	st, err := ws.Stat(app)
	if err != nil || !st.Exists {
		return bundleInfo{}, problem.Fail(msg("WAP-PATH-NOT-FOUND", fmt.Sprintf("%s does not exist.", app)))
	}
	if !st.Dir || !strings.HasSuffix(strings.TrimRight(app, "/"+string(filepath.Separator)), ".app") {
		return bundleInfo{}, problem.Fail(msg("WAP-NOT-A-BUNDLE", fmt.Sprintf("%s is not a directory ending in .app.", app)))
	}
	data, err := ws.Read(filepath.Join(app, "Contents", "Info.plist"))
	if err != nil {
		return bundleInfo{}, problem.Fail(msg("WAP-NOT-A-BUNDLE", fmt.Sprintf("%s has no readable Contents/Info.plist.", app)))
	}
	v, err := plist.Decode(data)
	p, ok := v.(map[string]any)
	if err != nil || !ok {
		reason := "it is not a property list"
		if errors.Is(err, plist.ErrBinary) {
			reason = "it is a binary property list"
		}
		return bundleInfo{}, problem.Fail(msg("WAP-NOT-A-BUNDLE", fmt.Sprintf("The Info.plist of %s cannot be read: %s.", app, reason)))
	}
	get := func(k string) string { s, _ := p[k].(string); return s }
	b := bundleInfo{
		path:       app,
		name:       get("CFBundleName"),
		version:    get("CFBundleShortVersionString"),
		identifier: get("CFBundleIdentifier"),
		executable: get("CFBundleExecutable"),
	}
	if b.name == "" {
		b.name = strings.TrimSuffix(filepath.Base(app), ".app")
	}
	return b, nil
}

// unavailable turns an adapter's report that it cannot work here into a
// message.
func unavailable(err error) error {
	var u *Unavailable
	if !errors.As(err, &u) {
		return err
	}
	if u.OS != "" {
		return problem.Missing(msg("WAP-PLATFORM-MACOS", fmt.Sprintf("This step needs %s, which exists only on macOS; this is %s.", u.Tool, u.OS)))
	}
	m := msg("WAP-TOOL-MISSING", fmt.Sprintf("%s was not found.", u.Tool))
	if u.Install != "" {
		m.How = u.Install
	}
	return problem.Missing(m)
}

// ready checks every port a use case needs before it does anything.
func ready(ports ...interface{ Ready() error }) error {
	for _, p := range ports {
		if err := p.Ready(); err != nil {
			return unavailable(err)
		}
	}
	return nil
}
