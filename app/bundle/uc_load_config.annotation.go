package bundle

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
)

var _ = spec.For[LoadConfig](
	spec.Satisfies(config.RConfigFile, config.RConfigStrict, config.RConfigEnv, dec.RDecJsonConfig),
	spec.Help(`Reads the configuration file. A syntax error is reported with its line, an unknown field is refused with the closest known one, ${NAME} is replaced by the environment variable NAME, and relative paths are taken relative to the file.`),
)
