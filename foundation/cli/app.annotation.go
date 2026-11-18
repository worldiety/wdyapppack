package cli

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
	"github.com/worldiety/wdyapppack/requirements/nfr"
)

// The command line layer is where the program explains itself: the overview,
// the help of every command, the messages, the suggestions, the JSON form and
// the exit codes all come from here.
var _ = spec.For[App](
	spec.Satisfies(
		guide.RGuideNoArgs, guide.RGuideCommandHelp, guide.RGuideMessages, guide.RGuideSuggest,
		guide.RGuideJson, guide.RGuideExitCodes, guide.RGuideNextSteps,
		dec.RDecHelpSingleSource, nfr.RNfrNonInteractive,
	),
)
