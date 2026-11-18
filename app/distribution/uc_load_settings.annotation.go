package distribution

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
)

var _ = spec.For[LoadSettings](
	spec.Satisfies(config.RConfigFile, config.RConfigStrict, config.RConfigEnv),
	spec.Help(`Reads the sections sign, notarize and dmg of the configuration file. An unknown field in them is refused with the closest known one, ${NAME} is replaced by the environment variable NAME, and relative paths are taken relative to the file.`),
)
