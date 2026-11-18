package distribution

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

var _ = spec.For[CreateArchive](
	spec.Satisfies(dist.RDistZip, nfr.RNfrCrossPlatformCore, guide.RGuideNoOverwrite),
	spec.Help(`Wraps a bundle into a zip archive with the bundle as its top level entry, keeping symbolic links as links and executable files executable — as Apple's ditto -c -k --keepParent does, but on every operating system.`),
)
