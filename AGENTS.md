# Working in this repository

`wdyapppack` packages prebuilt programs as macOS `.app` bundles. The program
explains itself: run `wdyapppack` or `wdyapppack help` for the user facing
documentation. This file is for whoever changes the code.

The project is checked by [speclink](https://github.com/worldiety/speclink)
under the profile `go_bare_ddd1`, which links every requirement to the code
that satisfies it and the test that verifies it. The rules below are not style
preferences: speclink enforces them, and a change that breaks one fails the
check with a message naming the fix.

## Constraints

- **Standard library only.** The one module beyond it is
  `github.com/worldiety/speclink/spec`, which itself carries nothing but the
  standard library. Do not add a dependency; if one seems unavoidable, stop and
  ask.
- **English everywhere**: code, comments, output, requirements, documents.
- **Never interactive.** A missing answer is an error that names the
  alternatives, because the caller is often an agent.
- **Help and behaviour have one source.** A command's flags, help text and JSON
  help are generated from one declaration in `foundation/cli`. Never write a
  flag into a help text by hand.

## The loop

Run all of it before calling a change finished. Every step must end clean.

    go mod tidy
    go build ./...
    go vet ./...
    go test -json ./... | speclink evidence
    speclink verify
    speclink generate -out SPECIFICATION.md

`evidence` records which tests verified which requirements. `verify` reports
everything else. **Every finding is an error** — there are no warnings, and
nothing is advisory. Read the `How:` line of a finding; it is usually the fix.

When a promised shape changes on purpose (the configuration file, a channel
contract), record the new one and commit `speclink.lock`:

    speclink freeze

To regenerate the specification document:

    speclink generate -out SPECIFICATION.md

Never run `speclink freeze -reviewer` or `speclink attest -reviewer` on behalf
of a person. A review is a statement a human makes.

## Where things go

    cmd/wdyapppack/          the entry point; the only place naming an adapter
    foundation/              the little framework: auth, permission, cli, problem,
                             suggest, confdoc, run
    pkg/                     domain free libraries: plist, icns, imaging, macho, ziptree
    app/bundle/              context: creating a bundle (build, icon, universal, init, schema)
    app/distribution/        context: handing it out (sign, notarize, dmg, zip)
    app/<context>/cli/       package cli<context>: the commands of that context
    app/<context>/adapter/<tech>/   implementations of the context's ports
    process/                 courses of business across contexts (P-PREPARE, P-DISTRIBUTE)
    topology/                the world outside, and every way across the boundary;
                             every adapter needs a channel here
    requirements/            requirements as Go values
    requirements/_sources/   the documents requirements cite

A context must not import another context, a presentation, or an adapter. It
declares ports; `cmd` decides what is behind them. `pkg/` and `foundation/`
never import a context.

## Why some things are the way they are

The decisions that shape the product are requirements in `requirements/dec/`,
each with its rationale and its cost. These are the ones below that level:

- **Options are parsed by `foundation/cli`, not by the `flag` package.** `flag`
  stops at the first argument that is not an option and reports mistakes in
  its own words; a self-explaining tool needs paths before or after options
  and the closest option suggested for a typo.
- **`--dry-run` comes with every command that writes, `--force` only with
  those that create an output** (`Command.Writes`, `Command.Replaces`). `sign`
  and `notarize` create nothing that could already exist.
- **Shared parts live in neutral packages.** speclink forbids a context to
  import the presentation, so what both need sits apart: `problem` (messages),
  `suggest` (closest names), `confdoc` (field docs and schema), `run`
  (starting Apple's tools).
- **`requirements/catalogue` imports every requirement package.** Go
  initialises only imported packages; without it, requirements no code binds
  to would be missing from `wdyapppack help requirements`.
- **The help topic about universal programs is called `architectures`**,
  because a topic may not share its name with the command `universal`.

## Adding a use case

1. **The requirement**, in `requirements/fun/<subject>/R-<ID>.spec.go`, citing a
   section of `requirements/_sources/wdyapppack/product.md`.
2. **The use case**, in `app/<context>/uc_<name>.go`: a function type with a
   `New<Type>` constructor taking the ports it needs. The first thing it does
   is audit its permission.
3. **The permission** in `app/<context>/perm.go`, and the use case in the
   `UseCases` bundle.
4. **The binding** in `app/<context>/uc_<name>.annotation.go`, with
   `spec.Satisfies` and a `spec.Help` text — the help text is shown by
   `wdyapppack help <command>`.
5. **The process step** in `process/`, so the use case belongs to a course of
   business.
6. **A test** that ends with `spec.Verified(t, ...)`.

## Help texts

Command help is generated from the declaration in `app/<context>/cli/commands.go`
and the `spec.Help` text of the use case. Topics are embedded text files in
`app/<context>/cli/help/` and `cmd/wdyapppack/help/`: paragraphs are filled to
80 columns, lines starting with two spaces are kept as they are. Every message
code (`WAP-…`) must be explained in the context's `messages.go`; a test fails
for a code that is used but not listed by `wdyapppack help errors`.

The configuration file's shape is a contract recorded in `speclink.lock`
(channel "init, build, icon, universal"). Fields may be added; renaming or
removing one is reported by `speclink verify`.

## Requirements and sources

Do not edit `requirements/_sources/` without the owner's consent. If a source
changes, speclink reports which requirements depended on the changed section.
Every section must be cited by a requirement or marked
`<!-- speclink:informative -->` on the line after its heading.

## Commits

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):

    <type>(<scope>): <summary in the imperative, lower case, no full stop>

    <body: what changed and why, wrapped at 72 columns>

- **Types**: `feat` (a new capability for the user), `fix` (a bug), `docs`
  (help texts, AGENTS.md, README.md, SPECIFICATION.md), `refactor`, `test`,
  `build`, `ci`, `chore`. A change to `requirements/` is `feat` or `fix`
  depending on what it does for the user.
- **Scope**: optional; the part of the tree that changed, e.g. `bundle`,
  `distribution`, `cli`, `icns`, `requirements`, `topology`.
- **Breaking changes** — a field removed from or renamed in the configuration
  file, an option removed, an exit code or message code changed — carry a `!`
  after the type or scope and a `BREAKING CHANGE:` footer saying what users
  must change.

Examples:

    feat(bundle): add --document-type for CFBundleDocumentTypes
    fix(distribution): staple disk images after notarization
    docs: explain entitlements in help signing
    feat(cli)!: rename --min-macos to --minimum-macos

## When a rule is genuinely wrong

Waive it at the construct it applies to, with a reason a person can check:

    var _ = spec.For[Thing](
        spec.Waive("K14-REQ-UNVERIFIED", "why this one cannot be tested"),
    )
