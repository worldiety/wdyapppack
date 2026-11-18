package clibundle

import (
	"embed"
	"fmt"
	"slices"
	"strings"

	"github.com/worldiety/wdyapppack/app/bundle"
	"github.com/worldiety/wdyapppack/foundation/cli"
)

//go:embed help/*.txt
var help embed.FS

func text(name string) string {
	data, err := help.ReadFile("help/" + name + ".txt")
	if err != nil {
		panic("clibundle: help text " + name + " is missing")
	}
	return string(data)
}

// Topics are the help topics of this context.
func Topics() []cli.Topic {
	wrapped := func(name string, fill func(string) string) func() string {
		return func() string {
			t := text(name)
			if fill != nil {
				t = fill(t)
			}
			return cli.Wrap(t, 0, cli.Width)
		}
	}
	return []cli.Topic{
		{Name: "bundle", Summary: "what a .app bundle is and how wdyapppack builds one", Body: wrapped("bundle", nil)},
		{Name: "info-plist", Summary: "every Info.plist key, usage descriptions, more keys", Body: wrapped("info-plist", usageKinds)},
		{Name: "icons", Summary: "image formats, sizes, and the icon shape of macOS 26", Body: wrapped("icons", nil)},
		{Name: "binaries", Summary: "programs: architectures, signatures, libraries", Body: wrapped("binaries", nil)},
		{Name: "architectures", Summary: "Apple silicon and Intel: one universal program for both", Body: wrapped("universal", nil)},
		{Name: "categories", Summary: "every application category Apple defines", Body: categories},
	}
}

func usageKinds(t string) string {
	kinds := bundle.UsageKinds()
	names := make([]string, 0, len(kinds))
	for k := range kinds {
		names = append(names, k)
	}
	slices.Sort(names)
	var b strings.Builder
	for _, k := range names {
		fmt.Fprintf(&b, "  %-22s %s\n", k, kinds[k])
	}
	return strings.Replace(t, "%USAGE%", strings.TrimRight(b.String(), "\n"), 1)
}

func categories() string {
	var b strings.Builder
	b.WriteString(cli.Wrap("The application category (LSApplicationCategoryType) tells macOS and the App Store what kind of application this is. Give it with --category or \"category\" in the configuration file, either by the short name or in full: developer-tools and public.app-category.developer-tools are the same.", 0, cli.Width))
	b.WriteString("\n")
	for _, c := range bundle.Categories() {
		fmt.Fprintf(&b, "  %-22s public.app-category.%s\n", c, c)
	}
	return b.String()
}
