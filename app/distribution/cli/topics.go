package clidistribution

import (
	"embed"

	"github.com/worldiety/wdyapppack/foundation/cli"
)

//go:embed help/*.txt
var help embed.FS

// Topics are the help topics of this context.
func Topics() []cli.Topic {
	text := func(name string) func() string {
		return func() string {
			data, err := help.ReadFile("help/" + name + ".txt")
			if err != nil {
				panic("clidistribution: help text " + name + " is missing")
			}
			return cli.Wrap(string(data), 0, cli.Width)
		}
	}
	return []cli.Topic{
		{Name: "signing", Summary: "ad hoc and Developer ID signatures, identities, entitlements", Body: text("signing")},
		{Name: "notarization", Summary: "notarizing: credentials, stapling, rejections", Body: text("notarization")},
		{Name: "distribution", Summary: "from a bundle to a download: disk image or zip", Body: text("distribution")},
	}
}
