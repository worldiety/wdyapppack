package bundle

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/binary"
)

var _ = spec.For[MergeUniversal](
	spec.Satisfies(binary.RBinaryUniversal, binary.RBinaryArchDetect, binary.RBinaryMachoOnly),
	spec.Help(`Merges programs built for Apple silicon (arm64) and for Intel (x86_64) into one universal program that runs natively on both, like Apple's lipo. The programs are copied byte for byte, so their signatures stay valid. build does the same on its own when given several programs; this command is for when the universal program is needed outside a bundle.`),
)
