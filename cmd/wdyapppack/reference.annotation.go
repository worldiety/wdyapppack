package main

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/guide"
)

var _ = spec.ForDecl(generalTopics, spec.Satisfies(guide.RGuideHelpTopics))
var _ = spec.ForDecl(requirementsTopic, spec.Satisfies(guide.RGuideRequirements))
var _ = spec.ForDecl(versionCommand, spec.Satisfies(guide.RGuideVersion))
var _ = spec.ForDecl(schema, spec.Satisfies(config.RConfigSchema))
