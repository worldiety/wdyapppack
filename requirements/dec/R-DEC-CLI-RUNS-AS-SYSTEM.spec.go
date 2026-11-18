package dec

import "github.com/worldiety/speclink/spec"

// RDecCliRunsAsSystem is the requirement "Use cases run as the system subject".
var RDecCliRunsAsSystem = spec.Declare(spec.Requirement{
	ID:           "R-DEC-CLI-RUNS-AS-SYSTEM",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Use cases run as the system subject",
	Text:         "Every use case MUST run as the system subject, so that what wdyapppack may do is decided by the operating system user who runs it.",
	Rationale:    "The architecture demands a permission per use case, which keeps the use cases ready for a caller other than the local user. A command line tool has no identity beyond the one the operating system already checked, and inventing accounts would add nothing but a way to lock oneself out.",
	Consequences: "The permissions are declared and audited but grant nothing in practice. A future server or remote mode must introduce real subjects before it exposes any use case, or it exposes all of them.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "91-who-runs-it"},
	},
})
