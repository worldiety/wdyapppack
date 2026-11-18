// Package process describes the courses of business: in which order the use
// cases of both contexts run, where the way branches and how it ends.
//
// It lives above the contexts because a course of business is the one thing no
// single context owns: the way to a download starts in the bundle context and
// ends in the distribution context.
package process

import (
	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/app/bundle"
	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/requirements/fun/binary"
	rbundle "github.com/worldiety/wdyapppack/requirements/fun/bundle"
	"github.com/worldiety/wdyapppack/requirements/fun/config"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
	"github.com/worldiety/wdyapppack/requirements/fun/icon"
)

// PPrepare is what may happen before a build: a configuration is written, an
// icon checked, programs merged.
var PPrepare = spec.Process{
	ID:        "P-PREPARE",
	Title:     "Preparing a build",
	Purpose:   "The inputs of a build are put into shape: a starter configuration, an icon, a universal program.",
	Satisfies: []spec.Requirement{config.RConfigInit, icon.RIconIconset, binary.RBinaryUniversal},
	Nodes: []spec.Node{
		spec.Start("material", "A compiled program and an image exist"),
		spec.Choice("need"),
		spec.Do[bundle.WriteConfig]("init"),
		spec.Do[bundle.ConvertIcon]("icon"),
		spec.Do[bundle.MergeUniversal]("universal"),
		spec.Merge("prepared"),
		spec.End("ready", "ready to build"),
	},
	Edges: []spec.Edge{
		{From: "material", To: "need"},
		{From: "need", To: "init", When: "no configuration file yet"},
		{From: "need", To: "icon", When: "the icon is to be checked or used elsewhere"},
		{From: "need", To: "universal", When: "a universal program is needed outside a bundle"},
		{From: "init", To: "prepared"},
		{From: "icon", To: "prepared"},
		{From: "universal", To: "prepared"},
		{From: "prepared", To: "ready"},
	},
}

// PDistribute is the way from a compiled program to something other people
// can download.
var PDistribute = spec.Process{
	ID:      "P-DISTRIBUTE",
	Title:   "From a program to a download",
	Purpose: "A compiled program becomes a bundle, is signed, notarized when it is to be handed out, and wrapped for download.",
	Satisfies: []spec.Requirement{
		rbundle.RBundleLayout, dist.RDistSignOrder, dist.RDistNotarize, dist.RDistDmg, dist.RDistZip,
	},
	Nodes: []spec.Node{
		spec.Start("program", "A program is compiled for macOS"),
		spec.Do[bundle.LoadConfig]("load-config"),
		spec.Do[bundle.BuildBundle]("build"),
		spec.Do[distribution.LoadSettings]("load-settings"),
		spec.Do[distribution.SignBundle]("sign").Note("ad hoc for this Mac, with a Developer ID to hand it out"),
		spec.Choice("audience"),
		spec.End("own-mac", "runs on this Mac"),
		spec.Do[distribution.NotarizeArtifact]("notarize").Note("needs macOS and an Apple Developer account"),
		spec.Choice("format"),
		spec.Do[distribution.CreateDiskImage]("dmg"),
		spec.Do[distribution.CreateArchive]("zip"),
		spec.End("dmg-ready", "disk image ready to download"),
		spec.End("zip-ready", "zip archive ready to download"),
	},
	Edges: []spec.Edge{
		{From: "program", To: "load-config"},
		{From: "load-config", To: "build"},
		{From: "build", To: "load-settings"},
		{From: "load-settings", To: "sign"},
		{From: "sign", To: "audience"},
		{From: "audience", To: "own-mac", When: "only for this Mac"},
		{From: "audience", To: "notarize", When: "for other people"},
		{From: "notarize", To: "format"},
		{From: "format", To: "dmg", When: "as a disk image"},
		{From: "format", To: "zip", When: "as a zip archive"},
		{From: "dmg", To: "dmg-ready"},
		{From: "zip", To: "zip-ready"},
	},
}
