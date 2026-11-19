package clidistribution

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
)

// Without a path, the commands work on the bundle the configuration
// describes; this is where that is decided.
var _ = spec.ForDecl(target,
	spec.Satisfies(config.RConfigDefaultBundle),
)
