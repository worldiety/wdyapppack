package bundle

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/icon"
)

var _ = spec.For[ConvertIcon](
	spec.Satisfies(icon.RIconFormats, icon.RIconSizes, icon.RIconSquare, icon.RIconShape, icon.RIconIconset, dec.RDecPureGoIcns),
	spec.Help(`Makes a macOS icon from a square PNG or JPEG image: an .icns file with every size from 16 to 1024 pixels, or an .iconset directory for Apple's iconutil. Nothing from Apple is needed, so this works on every operating system. build does the same on its own; this command is for checking an icon first or for using it elsewhere.`),
)
