package bundle

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
)

var _ = spec.For[WriteConfig](
	spec.Satisfies(config.RConfigInit),
	spec.Help(`Writes a complete configuration file with every field, so that you can see what can be set. Whatever can be read from the program — its name, architecture and minimum macOS version — is filled in. What cannot be guessed, such as the identifier, holds the placeholder CHANGE-ME, which build refuses until you replace it.`),
)
