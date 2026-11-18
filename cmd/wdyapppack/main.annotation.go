package main

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/cst"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

// The program is assembled here: every command, every topic, in English, and
// every use case acting as the operating system user who started it.
var _ = spec.ForDecl(describe,
	spec.Satisfies(nfr.RNfrSelfExplaining, nfr.RNfrEnglish, dec.RDecCliRunsAsSystem),
)

// Everything the program is made of is wired here, over the standard library
// and nothing else.
var _ = spec.ForDecl(newApp,
	spec.Satisfies(cst.RCstStdlibOnly),
)
