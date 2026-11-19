// Command wdyapppack packages prebuilt programs as macOS .app bundles.
//
// This is the only place that names an adapter and the only place that knows
// both contexts. Everything above it takes ports, which is what lets a test put
// a different implementation behind them.
package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/worldiety/wdyapppack/app/bundle"
	bundlefs "github.com/worldiety/wdyapppack/app/bundle/adapter/fs"
	clibundle "github.com/worldiety/wdyapppack/app/bundle/cli"
	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/app/distribution/adapter/codesign"
	distfs "github.com/worldiety/wdyapppack/app/distribution/adapter/fs"
	"github.com/worldiety/wdyapppack/app/distribution/adapter/hdiutil"
	"github.com/worldiety/wdyapppack/app/distribution/adapter/notarytool"
	clidistribution "github.com/worldiety/wdyapppack/app/distribution/cli"
	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/cli"
	"github.com/worldiety/wdyapppack/foundation/run"
	_ "github.com/worldiety/wdyapppack/requirements/catalogue"
)

func main() {
	// A signal ends the process without running deferred functions, so what
	// an adapter set up for this command — a temporary keychain holding a
	// signing certificate — is removed here.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		run.Interrupted()
		os.Exit(130)
	}()
	os.Exit(newApp(os.LookupEnv, os.Stderr).Run(os.Args[1:], os.Stdout, os.Stderr))
}

// newApp assembles the program over the real file system and Apple's tools.
func newApp(env func(string) (string, bool), log io.Writer) *cli.App {
	verbose := false
	runner := &run.Runner{Log: log, Verbose: func() bool { return verbose }}

	b := bundle.NewUseCases(bundlefs.New())
	d := distribution.NewUseCases(
		distfs.New(),
		codesign.New(runner, codesign.DefaultTools()),
		notarytool.New(runner, notarytool.DefaultTools()),
		hdiutil.New(runner, hdiutil.DefaultTools()),
	)

	app := describe(b, d, env)
	app.Before = func(inv *cli.Invocation) { verbose = inv.Verbose }
	return app
}

// describe builds the program's commands and topics from the use cases.
func describe(b bundle.UseCases, d distribution.UseCases, env func(string) (string, bool)) *cli.App {
	// The distribution commands work on the bundle the configuration
	// describes when none is named. Only the bundle context knows how a
	// configuration names a bundle, and this is the one place that knows both.
	bundles := func(configPath string, optional bool) (string, error) {
		loaded, err := b.LoadConfig(auth.System(), bundle.LoadConfigRequest{Path: configPath, Optional: optional, Lookup: env})
		if err != nil || !loaded.Found || loaded.Config.Name == "" {
			return "", err
		}
		return loaded.Config.BundlePath(), nil
	}

	app := &cli.App{
		Name:      "wdyapppack",
		Tagline:   "package prebuilt programs as macOS .app bundles",
		Intro:     intro,
		Outro:     outro,
		AgentHint: agentHint,
		Groups:    []string{"Create", "Distribute", "Reference"},
		Subject:   auth.System(),
		Commands: []cli.Command{
			clibundle.Init(b.WriteConfig, distribution.StarterSections()),
			clibundle.Build(b.LoadConfig, b.BuildBundle, env),
			clibundle.Icon(b.ConvertIcon, env),
			clibundle.Universal(b.MergeUniversal),
			clidistribution.Sign(d.LoadSettings, d.SignBundle, bundles, env),
			clidistribution.Notarize(d.LoadSettings, d.NotarizeArtifact, bundles, env),
			clidistribution.DMG(d.LoadSettings, d.CreateDiskImage, bundles, env),
			clidistribution.Zip(d.CreateArchive, bundles, env),
			schemaCommand(),
			versionCommand(),
		},
		Errors: append(append(cli.Codes(), bundle.Codes()...), distribution.Codes()...),
		Extra:  map[string]any{"configFields": configFields()},
	}
	app.Topics = append(app.Topics, clibundle.Topics()...)
	app.Topics = append(app.Topics, clidistribution.Topics()...)
	app.Topics = append(app.Topics, generalTopics(app)...)
	return app.WithHelp()
}
