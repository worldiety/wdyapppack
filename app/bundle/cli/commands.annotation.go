package clibundle

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
)

// An option given on the command line is laid over the configuration file
// here, and only when it was actually given.
var _ = spec.ForDecl(applyFlags,
	spec.Satisfies(config.RConfigOverride),
)
