package bundle

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/binary"
	rbundle "github.com/worldiety/wdyapppack/requirements/fun/bundle"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/fun/icon"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

var _ = spec.For[BuildBundle](
	spec.Satisfies(
		rbundle.RBundleLayout, rbundle.RBundleMetadataRequired, rbundle.RBundlePlistDefaults,
		rbundle.RBundleCategory, rbundle.RBundlePlistCustom, rbundle.RBundlePlistProtected,
		rbundle.RBundlePayload, rbundle.RBundleSafeWrite, rbundle.RBundleReproducible,
		rbundle.RBundleSelfCheck, rbundle.RBundleStrict,
		binary.RBinaryMachoOnly, binary.RBinaryArchDetect, binary.RBinaryMinos,
		binary.RBinaryArm64Signed, binary.RBinaryDylibExternal, binary.RBinaryUniversal,
		icon.RIconFormats, icon.RIconSizes, icon.RIconSquare, icon.RIconShape,
		guide.RGuideDryRun, guide.RGuideNoOverwrite,
		dec.RDecDirectLaunchOnly, dec.RDecNoLoadpathRewrite, dec.RDecPureGoIcns,
		nfr.RNfrCrossPlatformCore,
	),
	spec.Help(`Creates <name>.app from a prebuilt macOS program. The program is copied byte for byte into Contents/MacOS and started directly by macOS; programs for several architectures are merged into one universal program. The icon is made from an image in every size macOS uses, and Info.plist is written with every key an application needs.

Before anything is written, everything is checked: the program must be a macOS executable, the identifier and version well formed, the icon square. Every problem found is reported at once. Problems that do not prevent a working bundle are reported as notes. The bundle is assembled next to its destination and moved into place only when complete, then read back and checked.`),
)
