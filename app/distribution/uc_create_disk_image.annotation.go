package distribution

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
)

var _ = spec.For[CreateDiskImage](
	spec.Satisfies(dist.RDistDmg, dist.RDistChecksum, dist.RDistTemporaryKeychain, dist.RDistPlatform, cst.RCstAppleToolsForSigning, guide.RGuideNoOverwrite),
	spec.Help(`Wraps a bundle into a compressed, read only disk image with Apple's hdiutil. The image holds the bundle and a link to the Applications folder, so that installing means dragging one icon onto the other. It is compressed as far as the format allows. Optionally the image is signed too; a signed image can be notarized itself. Next to it, a .sha256 file holds its checksum in the format of shasum -a 256.`),
)
