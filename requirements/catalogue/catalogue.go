// Package catalogue links every requirement into the program that imports it.
//
// Go initialises only the packages a program imports, so a requirement that no
// code binds to — a constraint, a quality goal — would be missing from
// spec.Requirements() at run time. The program lists its requirements in its
// help, and this import is what makes that list the whole tree.
package catalogue

import (
	_ "github.com/worldiety/wdyapppack/requirements/cst"
	_ "github.com/worldiety/wdyapppack/requirements/dec"
	_ "github.com/worldiety/wdyapppack/requirements/fun/binary"
	_ "github.com/worldiety/wdyapppack/requirements/fun/bundle"
	_ "github.com/worldiety/wdyapppack/requirements/fun/config"
	_ "github.com/worldiety/wdyapppack/requirements/fun/dist"
	_ "github.com/worldiety/wdyapppack/requirements/fun/guide"
	_ "github.com/worldiety/wdyapppack/requirements/fun/icon"
	_ "github.com/worldiety/wdyapppack/requirements/nfr"
)
