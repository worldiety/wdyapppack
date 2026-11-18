package cst

import "github.com/worldiety/speclink/spec"

// RCstAppleToolsForSigning is the requirement "Apple's tools for signing, notarization and disk images".
var RCstAppleToolsForSigning = spec.Declare(spec.Requirement{
	ID:         "R-CST-APPLE-TOOLS-FOR-SIGNING",
	Kind:       spec.Constraint,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Apple's tools for signing, notarization and disk images",
	Text:       "Code signing, notarization and disk images MUST be done by calling the tools Apple ships with macOS.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "26-apples-tools-where-only-apples-tools-will-do"},
	},
})
