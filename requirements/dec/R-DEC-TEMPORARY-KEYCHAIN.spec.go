package dec

import "github.com/worldiety/speclink/spec"

// RDecTemporaryKeychain is the requirement "A temporary keychain per command".
var RDecTemporaryKeychain = spec.Declare(spec.Requirement{
	ID:           "R-DEC-TEMPORARY-KEYCHAIN",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "A temporary keychain per command",
	Text:         "A command that signs with a certificate from the environment MUST create its own temporary keychain and delete it before it ends, rather than keep one between commands.",
	Rationale:    "A keychain kept between commands is state a build server has to set up before the first step and tear down after the last, also when a step fails in between; that is exactly the scripting this was meant to remove. A keychain that lives as long as one command leaves nothing behind and needs no step of its own.",
	Consequences: "Every command that signs imports the certificate again, which costs about a second, so sign followed by a signed disk image does it twice. While a command runs, its temporary keychain is in the user's keychain search list. A process killed outright (SIGKILL) cannot clean up and leaves a keychain file in the temporary directory.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/wdyapppack/product.md", Anchor: "812-signing-on-a-build-server"},
	},
})
