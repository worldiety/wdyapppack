package bundle

import "github.com/worldiety/speclink/spec"

// RBundleMetadataRequired is the requirement "Required metadata".
var RBundleMetadataRequired = spec.Declare(spec.Requirement{
	ID:         "R-BUNDLE-METADATA-REQUIRED",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Required metadata",
	Text:       "A build MUST be refused without a name, an identifier in reverse domain notation of letters, digits, dots and hyphens, and a version of one to three dot-separated numbers.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "52-required-metadata"},
		{Extern: "Apple, Information Property List Key Reference"},
	},
})
