package dec

import "github.com/worldiety/speclink/spec"

// RDecDirectLaunchOnly is the requirement "Programs are started directly".
var RDecDirectLaunchOnly = spec.Declare(spec.Requirement{
	ID:           "R-DEC-DIRECT-LAUNCH-ONLY",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Programs are started directly",
	Text:         "The program in a bundle MUST be started directly by macOS, without a launcher.",
	Rationale:    "A launcher script would be a second program in the bundle with failure modes of its own, would need a signature of its own and would change the process macOS sees. Most programs worth a bundle open a window or a browser themselves.",
	Consequences: "A command line program packaged as a bundle runs invisibly when opened from the Finder and its output is lost. The help has to say so, because nothing else will.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "512-start-behaviour"},
	},
})
