# wdyapppack

Derived from the source by `speclink generate`. Do not edit: every sentence here is written somewhere else, and the point of this file is that there is only one such place.

## How to read this

This document is derived from the source. Nothing in it was written by hand, and nothing in it can be edited into agreement with something that is not true of the code — which is the property that makes it worth reading, and the reason it is regenerated rather than maintained.

It is written for four readers at once. Each chapter below names the ones it is for.

| If you are | Start at | Because it answers |
|---|---|---|
| **running the project** | Where it stands, Gaps, The register | how much is agreed, built, tested and signed, and what is left |
| **auditing it** | The register, Standards, Source documents, Requirements | what is claimed, what evidence stands behind each claim, and what was never measured |
| **the one who asked for it** | Courses of business, The material, Requirements | whether the sentences you wrote survived into the thing that was built |
| **building on it** | The boundary, What answers from outside, How it is put together | what the system exposes, what it talks to, and the rules the code is held to |

### What a blank means

A chapter with nothing in it says which of two things happened. _Not declared_ means the project states nothing of that kind, and the emptiness is a fact about the project. _Not measured_ means this run could not look, and the emptiness is a fact about the run. They are never printed as the same blank, because a reader has no other way to tell them apart, and treating an unasked question as a clean answer is the failure this whole tool exists to prevent.

## Where it stands

|  | measured | complete |
|---|---:|---:|
| Source segments accounted for | 63 | 100% |
| Normative requirements covered | 68 | 100% |
| … claimed by a test | 68 | 100% |
| … demonstrated by a run | 68 | 100% |
| … read by a person | 68 | 90% |

## Gaps

### Requirements nobody has read

- R-CONFIG-DEFAULT-BUNDLE
- R-DEC-TEMPORARY-KEYCHAIN
- R-DIST-API-KEY-ENV
- R-DIST-CHECKSUM
- R-DIST-DMG
- R-DIST-NOTARIZE-SECRETS
- R-DIST-TEMPORARY-KEYCHAIN

## What has actually been run

A test that claims a requirement is a claim. Evidence that the test ran is something else, and this chapter keeps them apart.

|  | count | of normative |
|---|---:|---:|
| Normative requirements | 68 |  |
| … a test claims | 68 | 100% |
| … a run demonstrated | 68 | 100% |

### How much of the code a run went through

_no coverage profile has been handed to speclink evidence, so nothing is known about which code ran_

## The material

| Document | Kind | Segments | Cited | Read | Drifted |
|---|---|---:|---:|---:|---:|
| `requirements/_sources/wdyapppack/product.md` | markdown | 63 | 63 | 57 | 0 |

## Themes

_No theme is declared, so the requirements are not grouped._

## Standards

_No standard is declared, so no external clause is answered here._

## What gets built

This module builds 1 program.

### wdyapppack

Command wdyapppack packages prebuilt programs as macOS .app bundles.

Built from `cmd/wdyapppack`.

**Assembles** `bundle`, `distribution`

**Chooses** the following implementations. This is the only place in the module where that choice is made, which is what makes it reviewable here.

- `app/bundle/adapter/fs`
- `app/distribution/adapter/codesign`
- `app/distribution/adapter/fs`
- `app/distribution/adapter/hdiutil`
- `app/distribution/adapter/notarytool`

_How this program is invoked could not be read from the source. That is a limit of the reading, not a statement that it takes no arguments._

## How it is put together

Every rule below is enforced by `speclink verify`. None of it is advice: a violation is a finding with the same identifier printed here, so a rule a reader finds in this chapter is one the build already refuses to ignore.

Convention: `ddd1`.

### Where code lives

| Layer | Path | What belongs there |
|---|---|---|
| **Bounded contexts** | `app/<context>` | One context per part of the business. What a context knows is its own; nothing reaches across. |
| **Domain** | `app/<context>` | The rules of the business, in terms of the business. It names no protocol and no store. |
| **Presentation** | `app/<context>/{rest,cli}` | How the outside reaches a use case. One package per way in. |
| **Adapter** | `app/<context>/adapter/<technology>` | The other side of a port: the code that speaks to a database, a file system or another service. |
| **Entry points** | `cmd/<binary>` | Where the program is assembled. The only place allowed to choose which adapter is used. |
| **Infrastructure** | `pkg/, foundation/` | Technical helpers that know nothing about the business, and may not learn. Shared by every context, which is why they must stay ignorant of all of them. |

### What the code is held to

**The module has a main package.** (`K8-MAIN-EXISTS`)

A module with no entry point is a library, and every statement in this document about what the system does would be about something that never runs.

**Every main package lives under cmd/.** (`K8-MAIN-LOCATION`)

The place a program is assembled is the place its dependencies are chosen. Scattering that makes the wiring impossible to find and impossible to review.

**pkg/ and foundation/ hold no business knowledge and declare no use case.** (`K7-INFRA-DOMAIN-FREE`)

Infrastructure is shared by every context. The moment it knows about one, every other context inherits that knowledge and the contexts stop being separate.

**A context does not import its own user interface.** (`K6-CTX-NO-UI-IMPORT`)

The direction of that dependency is the whole point of the separation: the interface is built on the rules, and rules that reach back into a screen cannot be reused behind a second one.

**A use case is declared in a file of its own, named after it.** (`K5-UC-FILE`)

It is the unit this document is organised around and the unit a reviewer is asked to read. A file holding three of them cannot be reviewed as any of them.

**A use case checks a permission before it does anything.** (`K5-UC-AUTHZ`)

Authorisation placed anywhere else is authorisation that a second caller can skip.

**Domain code does not import rest/ or cli/.** (`K6-CTX-PRESENTATION-NO-IMPORT`)

A rule that knows how it is being called is a rule that can only be called that way.

**Nothing but cmd/ imports an adapter.** (`K6-ADAPTER-WIRED-IN-CMD`)

A port exists so the choice of technology is made once, where the program is assembled. An import anywhere else makes that choice permanent and invisible.

**A handler takes the use cases it calls, not the whole bundle.** (`K6-PRESENTATION-NO-BUNDLE`)

A handler holding everything has no readable dependencies, so nothing can be said about what it may reach.

## How the code is composed

36 packages in 2 bounded contexts, and 99 dependencies between them. Only this module's own packages: a dependency on the standard library or on a third party is not a fact about the shape of this system.

11 packages declare this specification rather than the system — the requirements, the courses of business, the boundary. They are left out of the drawing below: in a project that uses this tool properly they are most of the nodes and most of the arrows, and the architecture disappears underneath its own documentation.

_No diagram is included in this document. Pass -figures to speclink generate, after rendering the sources written by speclink diagrams._

### Where one context reaches into another

No package of one bounded context imports a package of another. The contexts are separate in the only sense a compiler can hold them to.

## What the code declares

20 constructs, each recognised by what it is rather than by an annotation saying so. Everything elsewhere in this document that names one of them points here.

### app/bundle

<a id="req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle"></a>
#### BuildBundle

_use case_ — `app/bundle/uc_build_bundle.go:6–18`

**Answers to** [R-BINARY-ARCH-DETECT](#req-R-BINARY-ARCH-DETECT), [R-BINARY-ARM64-SIGNED](#req-R-BINARY-ARM64-SIGNED), [R-BINARY-DYLIB-EXTERNAL](#req-R-BINARY-DYLIB-EXTERNAL), [R-BINARY-MACHO-ONLY](#req-R-BINARY-MACHO-ONLY), [R-BINARY-MINOS](#req-R-BINARY-MINOS), [R-BINARY-UNIVERSAL](#req-R-BINARY-UNIVERSAL), [R-BUNDLE-CATEGORY](#req-R-BUNDLE-CATEGORY), [R-BUNDLE-LAYOUT](#req-R-BUNDLE-LAYOUT), [R-BUNDLE-METADATA-REQUIRED](#req-R-BUNDLE-METADATA-REQUIRED), [R-BUNDLE-PAYLOAD](#req-R-BUNDLE-PAYLOAD), [R-BUNDLE-PLIST-CUSTOM](#req-R-BUNDLE-PLIST-CUSTOM), [R-BUNDLE-PLIST-DEFAULTS](#req-R-BUNDLE-PLIST-DEFAULTS), [R-BUNDLE-PLIST-PROTECTED](#req-R-BUNDLE-PLIST-PROTECTED), [R-BUNDLE-REPRODUCIBLE](#req-R-BUNDLE-REPRODUCIBLE), [R-BUNDLE-SAFE-WRITE](#req-R-BUNDLE-SAFE-WRITE), [R-BUNDLE-SELF-CHECK](#req-R-BUNDLE-SELF-CHECK), [R-BUNDLE-STRICT](#req-R-BUNDLE-STRICT), [R-DEC-DIRECT-LAUNCH-ONLY](#req-R-DEC-DIRECT-LAUNCH-ONLY), [R-DEC-NO-LOADPATH-REWRITE](#req-R-DEC-NO-LOADPATH-REWRITE), [R-DEC-PURE-GO-ICNS](#req-R-DEC-PURE-GO-ICNS), [R-GUIDE-DRY-RUN](#req-R-GUIDE-DRY-RUN), [R-GUIDE-NO-OVERWRITE](#req-R-GUIDE-NO-OVERWRITE), [R-ICON-FORMATS](#req-R-ICON-FORMATS), [R-ICON-SHAPE](#req-R-ICON-SHAPE), [R-ICON-SIZES](#req-R-ICON-SIZES), [R-ICON-SQUARE](#req-R-ICON-SQUARE), [R-NFR-CROSS-PLATFORM-CORE](#req-R-NFR-CROSS-PLATFORM-CORE)

<a id="req-code-github-com-worldiety-wdyapppack-app-bundle-converticon"></a>
#### ConvertIcon

_use case_ — `app/bundle/uc_convert_icon.go:6–18`

**Answers to** [R-DEC-PURE-GO-ICNS](#req-R-DEC-PURE-GO-ICNS), [R-ICON-FORMATS](#req-R-ICON-FORMATS), [R-ICON-ICONSET](#req-R-ICON-ICONSET), [R-ICON-SHAPE](#req-R-ICON-SHAPE), [R-ICON-SIZES](#req-R-ICON-SIZES), [R-ICON-SQUARE](#req-R-ICON-SQUARE)

<a id="req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig"></a>
#### LoadConfig

_use case_ — `app/bundle/uc_load_config.go:6–18`

**Answers to** [R-CONFIG-ENV](#req-R-CONFIG-ENV), [R-CONFIG-FILE](#req-R-CONFIG-FILE), [R-CONFIG-STRICT](#req-R-CONFIG-STRICT), [R-DEC-JSON-CONFIG](#req-R-DEC-JSON-CONFIG)

<a id="req-code-github-com-worldiety-wdyapppack-app-bundle-mergeuniversal"></a>
#### MergeUniversal

_use case_ — `app/bundle/uc_merge_universal.go:6–18`

**Answers to** [R-BINARY-ARCH-DETECT](#req-R-BINARY-ARCH-DETECT), [R-BINARY-MACHO-ONLY](#req-R-BINARY-MACHO-ONLY), [R-BINARY-UNIVERSAL](#req-R-BINARY-UNIVERSAL)

<a id="req-code-github-com-worldiety-wdyapppack-app-bundle-writeconfig"></a>
#### WriteConfig

_use case_ — `app/bundle/uc_write_config.go:6–18`

**Answers to** [R-CONFIG-INIT](#req-R-CONFIG-INIT)

<a id="req-code-bundle-build"></a>
#### bundle.build

_permission_ — `app/bundle/perm.go:11`

<a id="req-code-bundle-config-load"></a>
#### bundle.config.load

_permission_ — `app/bundle/perm.go:10`

<a id="req-code-bundle-config-write"></a>
#### bundle.config.write

_permission_ — `app/bundle/perm.go:14`

<a id="req-code-bundle-icon-convert"></a>
#### bundle.icon.convert

_permission_ — `app/bundle/perm.go:12`

<a id="req-code-bundle-universal-merge"></a>
#### bundle.universal.merge

_permission_ — `app/bundle/perm.go:13`

### app/distribution

<a id="req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive"></a>
#### CreateArchive

_use case_ — `app/distribution/uc_create_archive.go:6–16`

**Answers to** [R-DIST-CHECKSUM](#req-R-DIST-CHECKSUM), [R-DIST-ZIP](#req-R-DIST-ZIP), [R-GUIDE-NO-OVERWRITE](#req-R-GUIDE-NO-OVERWRITE), [R-NFR-CROSS-PLATFORM-CORE](#req-R-NFR-CROSS-PLATFORM-CORE)

<a id="req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage"></a>
#### CreateDiskImage

_use case_ — `app/distribution/uc_create_disk_image.go:6–16`

**Answers to** [R-CST-APPLE-TOOLS-FOR-SIGNING](#req-R-CST-APPLE-TOOLS-FOR-SIGNING), [R-DIST-CHECKSUM](#req-R-DIST-CHECKSUM), [R-DIST-DMG](#req-R-DIST-DMG), [R-DIST-PLATFORM](#req-R-DIST-PLATFORM), [R-DIST-TEMPORARY-KEYCHAIN](#req-R-DIST-TEMPORARY-KEYCHAIN), [R-GUIDE-NO-OVERWRITE](#req-R-GUIDE-NO-OVERWRITE)

<a id="req-code-github-com-worldiety-wdyapppack-app-distribution-loadsettings"></a>
#### LoadSettings

_use case_ — `app/distribution/uc_load_settings.go:6–16`

**Answers to** [R-CONFIG-ENV](#req-R-CONFIG-ENV), [R-CONFIG-FILE](#req-R-CONFIG-FILE), [R-CONFIG-STRICT](#req-R-CONFIG-STRICT)

<a id="req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact"></a>
#### NotarizeArtifact

_use case_ — `app/distribution/uc_notarize_artifact.go:6–16`

**Answers to** [R-CST-APPLE-TOOLS-FOR-SIGNING](#req-R-CST-APPLE-TOOLS-FOR-SIGNING), [R-DIST-API-KEY-ENV](#req-R-DIST-API-KEY-ENV), [R-DIST-NOTARIZE](#req-R-DIST-NOTARIZE), [R-DIST-NOTARIZE-LOG](#req-R-DIST-NOTARIZE-LOG), [R-DIST-NOTARIZE-SECRETS](#req-R-DIST-NOTARIZE-SECRETS), [R-DIST-PLATFORM](#req-R-DIST-PLATFORM), [R-DIST-STAPLE](#req-R-DIST-STAPLE)

<a id="req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle"></a>
#### SignBundle

_use case_ — `app/distribution/uc_sign_bundle.go:6–16`

**Answers to** [R-CST-APPLE-TOOLS-FOR-SIGNING](#req-R-CST-APPLE-TOOLS-FOR-SIGNING), [R-DEC-TEMPORARY-KEYCHAIN](#req-R-DEC-TEMPORARY-KEYCHAIN), [R-DIST-PLATFORM](#req-R-DIST-PLATFORM), [R-DIST-SIGN-ADHOC](#req-R-DIST-SIGN-ADHOC), [R-DIST-SIGN-IDENTITY](#req-R-DIST-SIGN-IDENTITY), [R-DIST-SIGN-ORDER](#req-R-DIST-SIGN-ORDER), [R-DIST-SIGN-VERIFY](#req-R-DIST-SIGN-VERIFY), [R-DIST-TEMPORARY-KEYCHAIN](#req-R-DIST-TEMPORARY-KEYCHAIN), [R-GUIDE-DRY-RUN](#req-R-GUIDE-DRY-RUN)

<a id="req-code-distribution-dmg-create"></a>
#### distribution.dmg.create

_permission_ — `app/distribution/perm.go:12`

<a id="req-code-distribution-notarize"></a>
#### distribution.notarize

_permission_ — `app/distribution/perm.go:11`

<a id="req-code-distribution-settings-load"></a>
#### distribution.settings.load

_permission_ — `app/distribution/perm.go:9`

<a id="req-code-distribution-sign"></a>
#### distribution.sign

_permission_ — `app/distribution/perm.go:10`

<a id="req-code-distribution-zip-create"></a>
#### distribution.zip.create

_permission_ — `app/distribution/perm.go:13`

## The boundary

What this system talks to, and every way across. A channel that does not name its protocol, its data, who may use it and what protects it in transit is refused at the source, so every row below is complete by construction.

_No diagram is included in this document. Pass -figures to speclink generate, after rendering the sources written by speclink diagrams._

| Outside | Kind | Role |
|---|---|---|
| Operator | actor | A person, a build script or an AI agent who packages a program, signs it and hands it out. Acts with the rights of their operating system user. |
| File system | foreign system | The local file system. What is written there outlives the process; access rights and backups belong to whoever runs wdyapppack. |
| Apple toolchain | foreign system | codesign, security and hdiutil on macOS: they sign code with certificates from the user's keychain and create disk images. Present only on macOS. |
| Apple notary service | foreign system | Apple's online service that scans signed software and issues the ticket Gatekeeper checks. Reached through xcrun notarytool; the ticket is attached with xcrun stapler. |

### What this system relies on receiving

Every shape below is recorded in `speclink.lock`, so a change to one is a finding rather than a surprise.

A change to a recorded structure is reported the next run, which is the only warning available: the far end of a channel has not heard of this repository and cannot be told.

#### init, build, icon, universal (operator → app/bundle/cli)

**Shape** `Config`

| Field | Wire | Shape | Omitted when empty |
|---|---|---|---:|
| `Schema` | `$schema` | `string` | yes |
| `Name` | `name` | `string` | yes |
| `DisplayName` | `displayName` | `string` | yes |
| `Identifier` | `identifier` | `string` | yes |
| `Version` | `version` | `string` | yes |
| `BuildNumber` | `buildNumber` | `string` | yes |
| `Binaries` | `binaries` | [\[\]Binary](#req-shape-github-com-worldiety-wdyapppack-app-bundle-binary) | yes |
| `ExecutableName` | `executableName` | `string` | yes |
| `Icon` | `icon` | [Icon](#req-shape-github-com-worldiety-wdyapppack-app-bundle-icon) | yes |
| `MinimumMacOS` | `minimumMacOS` | `string` | yes |
| `Category` | `category` | `string` | yes |
| `Copyright` | `copyright` | `string` | yes |
| `Agent` | `agent` | `bool` | yes |
| `BackgroundOnly` | `backgroundOnly` | `bool` | yes |
| `Resources` | `resources` | [\[\]Resource](#req-shape-github-com-worldiety-wdyapppack-app-bundle-resource) | yes |
| `Frameworks` | `frameworks` | `[]string` | yes |
| `Helpers` | `helpers` | `[]string` | yes |
| `URLSchemes` | `urlSchemes` | `[]string` | yes |
| `UsageDescriptions` | `usageDescriptions` | `map[string]string` | yes |
| `Environment` | `environment` | `map[string]string` | yes |
| `Plist` | `plist` | `map[string]any` | yes |
| `PlistFile` | `plistFile` | `string` | yes |
| `Out` | `out` | `string` | yes |
| `Sign` | `sign` | `string` | yes |
| `Notarize` | `notarize` | `string` | yes |
| `DMG` | `dmg` | `string` | yes |

### Every way across

| Channel | Protocol | Data | Authentication | In transit |
|---|---|---|---|---|
| **init, build, icon, universal**<br>operator → app/bundle/cli | Command line, arguments and wdyapppack.json; text or JSON on standard output | Program paths, metadata, icon images, the configuration file | none of its own; the rights of the operating system user who runs it | none, local |
| **sign, notarize, dmg, zip**<br>operator → app/distribution/cli | Command line, arguments and the sign, notarize and dmg sections of wdyapppack.json | Bundle paths, signing identities, notarization credentials; on a build server a .p12 certificate and an App Store Connect API key from fixed environment variables (never options) | none of its own; the rights of the operating system user, and the keychain they can open or the certificate the environment hands over | none, local |
| **Programs, images and bundles**<br>app/bundle/adapter/fs → filesystem | File system | Programs, images and property lists read; bundles, icons, universal programs and configuration files written, each moved into place whole | Rights of the process user | none, local |
| **Bundles and archives**<br>app/distribution/adapter/fs → filesystem | File system | Bundles read; zip archives written | Rights of the process user | none, local |
| **codesign and security**<br>app/distribution/adapter/codesign → apple-toolchain | Process execution | Paths of code to sign and verify, identity fingerprints, entitlements, a .p12 certificate for a temporary keychain; signatures and identity lists back | The user's keychain holds the private keys; on a build server a temporary keychain with a random password, created and deleted by the command | Signatures by codesign; secure timestamps fetched by codesign from Apple over TLS |
| **hdiutil**<br>app/distribution/adapter/hdiutil → apple-toolchain | Process execution | A staging directory with the bundle and a link to /Applications; a compressed disk image back | Rights of the process user | none, local |
| **notarytool and stapler**<br>app/distribution/adapter/notarytool → apple-notary | Process execution of xcrun notarytool, which speaks HTTPS to Apple | A zip archive or disk image of signed software; a verdict, a log of issues and a ticket back | A keychain profile, an App Store Connect API key (from a file or the environment, written to a private temporary file), or an Apple ID with an app specific password | TLS between notarytool and Apple |

## What answers from outside

_No route was recognised: this system answers on no address of its own._

## Structures in full

Every structure the tables above pointed at, set out one field to a line. This is the form somebody writing the far end needs, and it is derived from the same types the shapes are: nothing here is maintained separately, so nothing here can drift.

<a id="req-shape-github-com-worldiety-wdyapppack-app-bundle-binary"></a>
### Binary

```
{
    arch: string
    path: string
}
```

<a id="req-shape-github-com-worldiety-wdyapppack-app-bundle-icon"></a>
### Icon

```
{
    path: string
    shape: string
}
```

<a id="req-shape-github-com-worldiety-wdyapppack-app-bundle-resource"></a>
### Resource

```
{
    from: string
    to: string
}
```

## Courses of business

### Preparing a build

The inputs of a build are put into shape: a starter configuration, an icon, a universal program.

`P-PREPARE`

Drawing source: `process-P-PREPARE.puml`

Answers to: R-CONFIG-INIT, R-ICON-ICONSET, R-BINARY-UNIVERSAL

| From | To | When |
|---|---|---|
| A compiled program and an image exist _(start)_ | need _(choice)_ | — |
| need _(choice)_ | [WriteConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-writeconfig) _(activity)_ | no configuration file yet |
| need _(choice)_ | [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) _(activity)_ | the icon is to be checked or used elsewhere |
| need _(choice)_ | [MergeUniversal](#req-code-github-com-worldiety-wdyapppack-app-bundle-mergeuniversal) _(activity)_ | a universal program is needed outside a bundle |
| [WriteConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-writeconfig) _(activity)_ | prepared _(merge)_ | — |
| [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) _(activity)_ | prepared _(merge)_ | — |
| [MergeUniversal](#req-code-github-com-worldiety-wdyapppack-app-bundle-mergeuniversal) _(activity)_ | prepared _(merge)_ | — |
| prepared _(merge)_ | ready to build _(end)_ | — |

### From a program to a download

A compiled program becomes a bundle, is signed, notarized when it is to be handed out, and wrapped for download.

`P-DISTRIBUTE`

Drawing source: `process-P-DISTRIBUTE.puml`

Answers to: R-BUNDLE-LAYOUT, R-DIST-SIGN-ORDER, R-DIST-NOTARIZE, R-DIST-DMG, R-DIST-ZIP

| From | To | When |
|---|---|---|
| A program is compiled for macOS _(start)_ | [LoadConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig) _(activity)_ | — |
| [LoadConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig) _(activity)_ | [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) _(activity)_ | — |
| [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) _(activity)_ | [LoadSettings](#req-code-github-com-worldiety-wdyapppack-app-distribution-loadsettings) _(activity)_ | — |
| [LoadSettings](#req-code-github-com-worldiety-wdyapppack-app-distribution-loadsettings) _(activity)_ | [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) _(activity)_ | — |
| [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) _(activity)_ | audience _(choice)_ | — |
| audience _(choice)_ | runs on this Mac _(end)_ | only for this Mac |
| audience _(choice)_ | [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) _(activity)_ | for other people |
| [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) _(activity)_ | format _(choice)_ | — |
| format _(choice)_ | [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) _(activity)_ | as a disk image |
| format _(choice)_ | [CreateArchive](#req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive) _(activity)_ | as a zip archive |
| [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) _(activity)_ | disk image ready to download _(end)_ | — |
| [CreateArchive](#req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive) _(activity)_ | zip archive ready to download _(end)_ | — |

## The register

Every requirement that was read, and how far each one has got. A mark states what was measured; where nothing looked, it says so rather than reporting a zero.

| Requirement | Kind | Field | Status | Built | Tested | Run | Read |
|---|---|---|---|---:|---:|---:|---:|
| [R-BINARY-ARCH-DETECT](#req-R-BINARY-ARCH-DETECT) | functional | business | normative | yes | yes | yes | yes |
| [R-BINARY-ARM64-SIGNED](#req-R-BINARY-ARM64-SIGNED) | functional | business | normative | yes | yes | yes | yes |
| [R-BINARY-DYLIB-EXTERNAL](#req-R-BINARY-DYLIB-EXTERNAL) | functional | business | normative | yes | yes | yes | yes |
| [R-BINARY-MACHO-ONLY](#req-R-BINARY-MACHO-ONLY) | functional | business | normative | yes | yes | yes | yes |
| [R-BINARY-MINOS](#req-R-BINARY-MINOS) | functional | business | normative | yes | yes | yes | yes |
| [R-BINARY-UNIVERSAL](#req-R-BINARY-UNIVERSAL) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-CATEGORY](#req-R-BUNDLE-CATEGORY) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-LAYOUT](#req-R-BUNDLE-LAYOUT) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-METADATA-REQUIRED](#req-R-BUNDLE-METADATA-REQUIRED) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-PAYLOAD](#req-R-BUNDLE-PAYLOAD) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-PLIST-CUSTOM](#req-R-BUNDLE-PLIST-CUSTOM) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-PLIST-DEFAULTS](#req-R-BUNDLE-PLIST-DEFAULTS) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-PLIST-PROTECTED](#req-R-BUNDLE-PLIST-PROTECTED) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-REPRODUCIBLE](#req-R-BUNDLE-REPRODUCIBLE) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-SAFE-WRITE](#req-R-BUNDLE-SAFE-WRITE) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-SELF-CHECK](#req-R-BUNDLE-SELF-CHECK) | functional | business | normative | yes | yes | yes | yes |
| [R-BUNDLE-STRICT](#req-R-BUNDLE-STRICT) | functional | business | normative | yes | yes | yes | yes |
| [R-CONFIG-DEFAULT-BUNDLE](#req-R-CONFIG-DEFAULT-BUNDLE) | functional | business | normative | yes | yes | yes | no |
| [R-CONFIG-ENV](#req-R-CONFIG-ENV) | functional | business | normative | yes | yes | yes | yes |
| [R-CONFIG-FILE](#req-R-CONFIG-FILE) | functional | business | normative | yes | yes | yes | yes |
| [R-CONFIG-INIT](#req-R-CONFIG-INIT) | functional | business | normative | yes | yes | yes | yes |
| [R-CONFIG-OVERRIDE](#req-R-CONFIG-OVERRIDE) | functional | business | normative | yes | yes | yes | yes |
| [R-CONFIG-SCHEMA](#req-R-CONFIG-SCHEMA) | functional | business | normative | yes | yes | yes | yes |
| [R-CONFIG-STRICT](#req-R-CONFIG-STRICT) | functional | business | normative | yes | yes | yes | yes |
| [R-CST-APPLE-TOOLS-FOR-SIGNING](#req-R-CST-APPLE-TOOLS-FOR-SIGNING) | constraint | technical | normative | yes | yes | yes | yes |
| [R-CST-STDLIB-ONLY](#req-R-CST-STDLIB-ONLY) | constraint | technical | normative | yes | yes | yes | yes |
| [R-DEC-CLI-RUNS-AS-SYSTEM](#req-R-DEC-CLI-RUNS-AS-SYSTEM) | decision | technical | normative | yes | yes | yes | yes |
| [R-DEC-DIRECT-LAUNCH-ONLY](#req-R-DEC-DIRECT-LAUNCH-ONLY) | decision | technical | normative | yes | yes | yes | yes |
| [R-DEC-HELP-SINGLE-SOURCE](#req-R-DEC-HELP-SINGLE-SOURCE) | decision | technical | normative | yes | yes | yes | yes |
| [R-DEC-JSON-CONFIG](#req-R-DEC-JSON-CONFIG) | decision | technical | normative | yes | yes | yes | yes |
| [R-DEC-NO-LOADPATH-REWRITE](#req-R-DEC-NO-LOADPATH-REWRITE) | decision | technical | normative | yes | yes | yes | yes |
| [R-DEC-PURE-GO-ICNS](#req-R-DEC-PURE-GO-ICNS) | decision | technical | normative | yes | yes | yes | yes |
| [R-DEC-TEMPORARY-KEYCHAIN](#req-R-DEC-TEMPORARY-KEYCHAIN) | decision | technical | normative | yes | yes | yes | no |
| [R-DIST-API-KEY-ENV](#req-R-DIST-API-KEY-ENV) | functional | business | normative | yes | yes | yes | no |
| [R-DIST-CHECKSUM](#req-R-DIST-CHECKSUM) | functional | business | normative | yes | yes | yes | no |
| [R-DIST-DMG](#req-R-DIST-DMG) | functional | business | normative | yes | yes | yes | no |
| [R-DIST-NOTARIZE](#req-R-DIST-NOTARIZE) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-NOTARIZE-LOG](#req-R-DIST-NOTARIZE-LOG) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-NOTARIZE-SECRETS](#req-R-DIST-NOTARIZE-SECRETS) | functional | business | normative | yes | yes | yes | no |
| [R-DIST-PLATFORM](#req-R-DIST-PLATFORM) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-SIGN-ADHOC](#req-R-DIST-SIGN-ADHOC) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-SIGN-IDENTITY](#req-R-DIST-SIGN-IDENTITY) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-SIGN-ORDER](#req-R-DIST-SIGN-ORDER) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-SIGN-VERIFY](#req-R-DIST-SIGN-VERIFY) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-STAPLE](#req-R-DIST-STAPLE) | functional | business | normative | yes | yes | yes | yes |
| [R-DIST-TEMPORARY-KEYCHAIN](#req-R-DIST-TEMPORARY-KEYCHAIN) | functional | business | normative | yes | yes | yes | no |
| [R-DIST-ZIP](#req-R-DIST-ZIP) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-COMMAND-HELP](#req-R-GUIDE-COMMAND-HELP) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-DRY-RUN](#req-R-GUIDE-DRY-RUN) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-EXIT-CODES](#req-R-GUIDE-EXIT-CODES) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-HELP-TOPICS](#req-R-GUIDE-HELP-TOPICS) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-JSON](#req-R-GUIDE-JSON) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-MESSAGES](#req-R-GUIDE-MESSAGES) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-NEXT-STEPS](#req-R-GUIDE-NEXT-STEPS) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-NO-ARGS](#req-R-GUIDE-NO-ARGS) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-NO-OVERWRITE](#req-R-GUIDE-NO-OVERWRITE) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-REQUIREMENTS](#req-R-GUIDE-REQUIREMENTS) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-SUGGEST](#req-R-GUIDE-SUGGEST) | functional | business | normative | yes | yes | yes | yes |
| [R-GUIDE-VERSION](#req-R-GUIDE-VERSION) | functional | business | normative | yes | yes | yes | yes |
| [R-ICON-FORMATS](#req-R-ICON-FORMATS) | functional | business | normative | yes | yes | yes | yes |
| [R-ICON-ICONSET](#req-R-ICON-ICONSET) | functional | business | normative | yes | yes | yes | yes |
| [R-ICON-SHAPE](#req-R-ICON-SHAPE) | functional | business | normative | yes | yes | yes | yes |
| [R-ICON-SIZES](#req-R-ICON-SIZES) | functional | business | normative | yes | yes | yes | yes |
| [R-ICON-SQUARE](#req-R-ICON-SQUARE) | functional | business | normative | yes | yes | yes | yes |
| [R-NFR-CROSS-PLATFORM-CORE](#req-R-NFR-CROSS-PLATFORM-CORE) | non-functional | technical | normative | yes | yes | yes | yes |
| [R-NFR-ENGLISH](#req-R-NFR-ENGLISH) | non-functional | technical | normative | yes | yes | yes | yes |
| [R-NFR-NON-INTERACTIVE](#req-R-NFR-NON-INTERACTIVE) | non-functional | technical | normative | yes | yes | yes | yes |
| [R-NFR-SELF-EXPLAINING](#req-R-NFR-SELF-EXPLAINING) | non-functional | mixed | normative | yes | yes | yes | yes |

### Reading the marks

- yes  yes, and it was measured
- part  in part
- no  no, and it was measured
- ?  not measured; nothing is claimed either way
- n/a  does not apply to this entry

- **Built** — something in the source declares that it satisfies this.
- **Tested** — a test claims it.
- **Run** — that test was seen to pass against this exact wording.
- **Read** — a named person recorded that they read this exact wording.

## Requirements

<a id="req-R-BINARY-ARCH-DETECT"></a>
### R-BINARY-ARCH-DETECT — Architectures

The architectures of a program MUST be recognised, and a stated architecture that does not match the program or is given twice MUST be refused.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#62-architectures
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [MergeUniversal](#req-code-github-com-worldiety-wdyapppack-app-bundle-mergeuniversal) — `app/bundle/uc_merge_universal.go:6–18`
- **Demonstrated by** TestArchitecturesAreRecognisedAndChecked
- **Read by** Timo

<a id="req-R-BINARY-ARM64-SIGNED"></a>
### R-BINARY-ARM64-SIGNED — Signature on Apple silicon

A program for Apple silicon without a signature MUST be noted with an explanation of how to sign it.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#64-signature-on-apple-silicon
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestUnsignedAppleSiliconProgramIsNoted
- **Read by** Timo

<a id="req-R-BINARY-DYLIB-EXTERNAL"></a>
### R-BINARY-DYLIB-EXTERNAL — Libraries outside the system

Every library a program loads from outside the system MUST be noted with an explanation of how to bundle it.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#65-libraries-outside-the-system
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestTheProgramIsPackagedAsItIs
- **Read by** Timo

<a id="req-R-BINARY-MACHO-ONLY"></a>
### R-BINARY-MACHO-ONLY — Only macOS programs

A program that is not a Mach-O executable MUST be refused with an explanation of how to build it for macOS.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#61-only-macos-programs
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [MergeUniversal](#req-code-github-com-worldiety-wdyapppack-app-bundle-mergeuniversal) — `app/bundle/uc_merge_universal.go:6–18`
- **Demonstrated by** TestOnlyMacOSProgramsArePackaged
- **Read by** Timo

<a id="req-R-BINARY-MINOS"></a>
### R-BINARY-MINOS — Minimum macOS version

The minimum macOS version MUST default to the one the program states, else to 11.0, and a version lower than the program's MUST be refused.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#63-minimum-macos-version
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestMinimumMacOS
- **Read by** Timo

<a id="req-R-BINARY-UNIVERSAL"></a>
### R-BINARY-UNIVERSAL — Universal programs

Programs for different architectures MUST be mergeable into one universal program that keeps their signatures, on their own and during a build.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#66-universal-programs
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [MergeUniversal](#req-code-github-com-worldiety-wdyapppack-app-bundle-mergeuniversal) — `app/bundle/uc_merge_universal.go:6–18`
- **Demonstrated by** TestMergeKeepsEverySliceByteForByte, TestUniversalProgramFromTwo
- **Read by** Timo

<a id="req-R-BUNDLE-CATEGORY"></a>
### R-BUNDLE-CATEGORY — Application category

The application category MUST be one that Apple defines, given in full or by its short name.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#54-application-category, Apple, Information Property List Key Reference
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestCategoryShortAndFull
- **Read by** Timo

<a id="req-R-BUNDLE-LAYOUT"></a>
### R-BUNDLE-LAYOUT — Bundle layout

A bundle MUST have the layout macOS expects, with an executable program in Contents/MacOS.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#51-bundle-layout, Apple, Bundle Programming Guide: Anatomy of a macOS Application Bundle
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestBuildWritesTheLayoutMacOSExpects
- **Read by** Timo

<a id="req-R-BUNDLE-METADATA-REQUIRED"></a>
### R-BUNDLE-METADATA-REQUIRED — Required metadata

A build MUST be refused without a name, an identifier in reverse domain notation of letters, digits, dots and hyphens, and a version of one to three dot-separated numbers.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#52-required-metadata, Apple, Information Property List Key Reference
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestRequiredMetadataIsReportedTogether
- **Read by** Timo

<a id="req-R-BUNDLE-PAYLOAD"></a>
### R-BUNDLE-PAYLOAD — Resources, libraries and helpers

Resources, frameworks and helpers MUST be copied into their directories in the bundle, and helpers MUST be checked like the main program.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#57-resources-libraries-and-helpers
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestResourcesFrameworksAndHelpers
- **Read by** Timo

<a id="req-R-BUNDLE-PLIST-CUSTOM"></a>
### R-BUNDLE-PLIST-CUSTOM — Additional metadata

Any further Info.plist key MUST be settable with its type, as an option or from a partial property list file.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#55-additional-metadata
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestAdditionalMetadataWithTypes
- **Read by** Timo

<a id="req-R-BUNDLE-PLIST-DEFAULTS"></a>
### R-BUNDLE-PLIST-DEFAULTS — Default metadata

The Info.plist MUST contain every key an application needs, with defaults, together with the optional keys the user sets.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#53-default-metadata, Apple, Information Property List Key Reference
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestInfoPlistHasEveryKeyAnApplicationNeeds
- **Read by** Timo

<a id="req-R-BUNDLE-PLIST-PROTECTED"></a>
### R-BUNDLE-PLIST-PROTECTED — Protected keys

Keys derived from the bundle itself MUST NOT be overwritten by additional metadata.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#56-protected-keys
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestProtectedKeysAreRefused
- **Read by** Timo

<a id="req-R-BUNDLE-REPRODUCIBLE"></a>
### R-BUNDLE-REPRODUCIBLE — Reproducible output

The same input MUST produce a byte-identical bundle when the file time is fixed.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#59-reproducible-output
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestSameInputSameBundle
- **Read by** Timo

<a id="req-R-BUNDLE-SAFE-WRITE"></a>
### R-BUNDLE-SAFE-WRITE — Safe writing

A bundle MUST be assembled next to its destination and moved into place only when it is complete.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#58-safe-writing
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestTreesAreMovedIntoPlaceWhole
- **Read by** Timo

<a id="req-R-BUNDLE-SELF-CHECK"></a>
### R-BUNDLE-SELF-CHECK — Self-check

After writing, the bundle MUST be read back and checked for a complete Info.plist, an executable program and the icon it names.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#510-self-check
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestTheBundleIsReadBack
- **Read by** Timo

<a id="req-R-BUNDLE-STRICT"></a>
### R-BUNDLE-STRICT — Notes and strict mode

Notes MUST NOT fail a build unless strict mode is requested, which MUST turn every note into an error.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#511-notes-and-strict-mode
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestNotesFailOnlyInStrictMode
- **Read by** Timo

<a id="req-R-CONFIG-DEFAULT-BUNDLE"></a>
### R-CONFIG-DEFAULT-BUNDLE — The bundle from the configuration

Signing, notarizing and wrapping MUST use the bundle the configuration file describes when no bundle is named.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#47-the-bundle-from-the-configuration
- **Implemented by**
  - `github.com/worldiety/wdyapppack/app/distribution/cli.target`
- **Demonstrated by** TestTheBundleComesFromTheConfiguration

<a id="req-R-CONFIG-ENV"></a>
### R-CONFIG-ENV — Environment variables

Text values in the configuration MAY refer to environment variables as ${NAME}, and a reference to an unset variable MUST be an error.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#44-environment-variables
- **Implemented by**
  - [LoadConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig) — `app/bundle/uc_load_config.go:6–18`
  - [LoadSettings](#req-code-github-com-worldiety-wdyapppack-app-distribution-loadsettings) — `app/distribution/uc_load_settings.go:6–16`
- **Demonstrated by** TestEnvironmentVariables, TestSettingsFromTheConfigurationFile
- **Read by** Timo

<a id="req-R-CONFIG-FILE"></a>
### R-CONFIG-FILE — Configuration file

Everything a build needs MUST be expressible in a JSON configuration file whose relative paths are resolved against the file's directory.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#41-configuration-file
- **Implemented by**
  - [LoadConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig) — `app/bundle/uc_load_config.go:6–18`
  - [LoadSettings](#req-code-github-com-worldiety-wdyapppack-app-distribution-loadsettings) — `app/distribution/uc_load_settings.go:6–16`
- **Demonstrated by** TestLoadConfigResolvesPathsAgainstTheFile, TestSettingsFromTheConfigurationFile
- **Read by** Timo

<a id="req-R-CONFIG-INIT"></a>
### R-CONFIG-INIT — Starter configuration

wdyapppack MUST write a complete starter configuration, filled from the program where possible and with placeholders that a build refuses everywhere else.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#45-starter-configuration
- **Implemented by**
  - [WriteConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-writeconfig) — `app/bundle/uc_write_config.go:6–18`
- **Demonstrated by** TestStarterConfiguration
- **Read by** Timo

<a id="req-R-CONFIG-OVERRIDE"></a>
### R-CONFIG-OVERRIDE — Options override the file

Every configuration value MUST also be settable as an option, and an option MUST win over the file.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#42-options-override-the-file
- **Implemented by**
  - `github.com/worldiety/wdyapppack/app/bundle/cli.applyFlags`
- **Demonstrated by** TestOptionsOverrideTheFileOnlyWhenGiven
- **Read by** Timo

<a id="req-R-CONFIG-SCHEMA"></a>
### R-CONFIG-SCHEMA — Schema of the configuration

wdyapppack MUST print a JSON Schema of the configuration file with a description of every field.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#46-schema
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.schema`
- **Demonstrated by** TestSchemaDescribesTheStarterConfiguration
- **Read by** Timo

<a id="req-R-CONFIG-STRICT"></a>
### R-CONFIG-STRICT — Unknown fields are refused

A configuration file with an unknown field MUST be refused.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#43-unknown-fields-are-refused
- **Implemented by**
  - [LoadConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig) — `app/bundle/uc_load_config.go:6–18`
  - [LoadSettings](#req-code-github-com-worldiety-wdyapppack-app-distribution-loadsettings) — `app/distribution/uc_load_settings.go:6–16`
- **Demonstrated by** TestSettingsFromTheConfigurationFile, TestUnknownFieldsAreRefused
- **Read by** Timo

<a id="req-R-CST-APPLE-TOOLS-FOR-SIGNING"></a>
### R-CST-APPLE-TOOLS-FOR-SIGNING — Apple's tools for signing, notarization and disk images

Code signing, notarization and disk images MUST be done by calling the tools Apple ships with macOS.

_constraint, technical, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#26-apples-tools-where-only-apples-tools-will-do
- **Implemented by**
  - [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) — `app/distribution/uc_create_disk_image.go:6–16`
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestCodesignIsAskedToSign
- **Read by** Timo

<a id="req-R-CST-STDLIB-ONLY"></a>
### R-CST-STDLIB-ONLY — Standard library only

wdyapppack MUST depend on nothing but the Go standard library and the speclink specification package.

_constraint, technical, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#24-standard-library-only
- **Derived requirements** [R-DEC-JSON-CONFIG](#req-R-DEC-JSON-CONFIG), [R-DEC-PURE-GO-ICNS](#req-R-DEC-PURE-GO-ICNS)
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.newApp`
- **Demonstrated by** TestStandardLibraryOnly
- **Read by** Timo

<a id="req-R-DEC-CLI-RUNS-AS-SYSTEM"></a>
### R-DEC-CLI-RUNS-AS-SYSTEM — Use cases run as the system subject

Every use case MUST run as the system subject, so that what wdyapppack may do is decided by the operating system user who runs it.

_decision, technical, normative._

**Why.** The architecture demands a permission per use case, which keeps the use cases ready for a caller other than the local user. A command line tool has no identity beyond the one the operating system already checked, and inventing accounts would add nothing but a way to lock oneself out.

**What it costs.** The permissions are declared and audited but grant nothing in practice. A future server or remote mode must introduce real subjects before it exposes any use case, or it exposes all of them.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#91-who-runs-it
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.describe`
- **Demonstrated by** TestUseCasesRunAsTheSystem
- **Read by** Timo

<a id="req-R-DEC-DIRECT-LAUNCH-ONLY"></a>
### R-DEC-DIRECT-LAUNCH-ONLY — Programs are started directly

The program in a bundle MUST be started directly by macOS, without a launcher.

_decision, technical, normative._

**Why.** A launcher script would be a second program in the bundle with failure modes of its own, would need a signature of its own and would change the process macOS sees. Most programs worth a bundle open a window or a browser themselves.

**What it costs.** A command line program packaged as a bundle runs invisibly when opened from the Finder and its output is lost. The help has to say so, because nothing else will.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#512-start-behaviour
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestTheProgramIsPackagedAsItIs
- **Read by** Timo

<a id="req-R-DEC-HELP-SINGLE-SOURCE"></a>
### R-DEC-HELP-SINGLE-SOURCE — Help and options come from one declaration

A command's options, its help text and its JSON help MUST be generated from one declaration.

_decision, technical, normative._

**Why.** Help written beside the code drifts away from it, and a self-explaining tool whose explanation is wrong is worse than one with none, because it is believed.

**What it costs.** Every command has to be declared through the foundation's command layer rather than with the flag package directly, and an option's help is limited to what fits its declaration; longer explanations go into help topics.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#33-help-for-every-command
- **Derived from** [R-NFR-SELF-EXPLAINING](#req-R-NFR-SELF-EXPLAINING)
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestCommandHelpComesFromTheDeclaration
- **Read by** Timo

<a id="req-R-DEC-JSON-CONFIG"></a>
### R-DEC-JSON-CONFIG — The configuration file is JSON

The configuration file MUST be written in JSON.

_decision, technical, normative._

**Why.** JSON is read and written by the standard library, while YAML or TOML would need a dependency that the standard library constraint rules out; JSON is also what editors complete and check from a JSON Schema.

**What it costs.** JSON has no comments, so a configuration cannot explain itself in place: the explanation has to live in the help and the schema. Trailing commas and missing quotes are syntax errors users will meet, so the error has to point at the line.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#41-configuration-file
- **Derived from** [R-CST-STDLIB-ONLY](#req-R-CST-STDLIB-ONLY)
- **Implemented by**
  - [LoadConfig](#req-code-github-com-worldiety-wdyapppack-app-bundle-loadconfig) — `app/bundle/uc_load_config.go:6–18`
- **Demonstrated by** TestLoadConfigResolvesPathsAgainstTheFile
- **Read by** Timo

<a id="req-R-DEC-NO-LOADPATH-REWRITE"></a>
### R-DEC-NO-LOADPATH-REWRITE — Library load paths are not rewritten

wdyapppack MUST NOT rewrite the library load paths of a program; it MUST only report libraries loaded from outside the system.

_decision, technical, normative._

**Why.** Rewriting load commands changes the program, invalidates its signature and is only safe with install\_name\_tool, which exists only on macOS. A packaging tool that edits the program it packages becomes responsible for that program's correctness.

**What it costs.** A program linked against libraries from a package manager has to be fixed by whoever builds it before its bundle runs on other Macs; wdyapppack can only point at the problem.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#65-libraries-outside-the-system
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
- **Demonstrated by** TestTheProgramIsPackagedAsItIs
- **Read by** Timo

<a id="req-R-DEC-PURE-GO-ICNS"></a>
### R-DEC-PURE-GO-ICNS — Icons are produced in Go

Icons MUST be resized and encoded into .icns by wdyapppack itself, without sips or iconutil.

_decision, technical, normative._

**Why.** Apple's image tools exist only on macOS, and icons must be produced on Linux build servers too. The .icns container is a simple list of tagged entries, and current macOS accepts PNG data in every entry it reads.

**What it costs.** Resampling quality is ours to get right and to test. Icon formats Apple adds later need code here, and asset catalogues (Assets.car) and Icon Composer files cannot be produced at all.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#72-sizes
- **Derived from** [R-NFR-CROSS-PLATFORM-CORE](#req-R-NFR-CROSS-PLATFORM-CORE), [R-CST-STDLIB-ONLY](#req-R-CST-STDLIB-ONLY)
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) — `app/bundle/uc_convert_icon.go:6–18`
- **Demonstrated by** TestEveryIconSizeInGo
- **Read by** Timo

<a id="req-R-DEC-TEMPORARY-KEYCHAIN"></a>
### R-DEC-TEMPORARY-KEYCHAIN — A temporary keychain per command

A command that signs with a certificate from the environment MUST create its own temporary keychain and delete it before it ends, rather than keep one between commands.

_decision, technical, normative._

**Why.** A keychain kept between commands is state a build server has to set up before the first step and tear down after the last, also when a step fails in between; that is exactly the scripting this was meant to remove. A keychain that lives as long as one command leaves nothing behind and needs no step of its own.

**What it costs.** Every command that signs imports the certificate again, which costs about a second, so sign followed by a signed disk image does it twice. While a command runs, its temporary keychain is in the user's keychain search list. A process killed outright (SIGKILL) cannot clean up and leaves a keychain file in the temporary directory.

- **Asked for in** requirements/\_sources/wdyapppack/product.md#812-signing-on-a-build-server
- **Implemented by**
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestSigningFromACertificateInTheEnvironment

<a id="req-R-DIST-API-KEY-ENV"></a>
### R-DIST-API-KEY-ENV — Notarization key from the environment

When no other way to reach the notary service is given, the App Store Connect API key MUST be taken from WDYAPPPACK\_API\_KEY, base64 encoded or as text, with WDYAPPPACK\_API\_KEY\_ID and optionally WDYAPPPACK\_API\_ISSUER, and written to a private temporary file only while the command runs.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#813-notarization-key-from-the-environment
- **Implemented by**
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
- **Demonstrated by** TestAPIKeyFromTheEnvironment

<a id="req-R-DIST-CHECKSUM"></a>
### R-DIST-CHECKSUM — Checksums

Next to every disk image and zip archive a SHA-256 checksum file MUST be written in the format of shasum -a 256.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#814-checksums
- **Implemented by**
  - [CreateArchive](#req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive) — `app/distribution/uc_create_archive.go:6–16`
  - [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) — `app/distribution/uc_create_disk_image.go:6–16`
- **Demonstrated by** TestChecksumsNextToImagesAndArchives

<a id="req-R-DIST-DMG"></a>
### R-DIST-DMG — Disk images

A bundle MUST be wrappable into a disk image with the highest compression the format offers that contains a link to the Applications folder.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#89-disk-images
- **Implemented by**
  - [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) — `app/distribution/uc_create_disk_image.go:6–16`
- **Demonstrated by** TestDiskImage, TestHdiutilIsAskedForACompressedImage

<a id="req-R-DIST-NOTARIZE"></a>
### R-DIST-NOTARIZE — Notarization

A bundle, disk image or zip archive signed with a Developer ID MUST be submittable to Apple for notarization, waiting for the result, and an unsuitable signature MUST be refused before anything is sent.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#85-notarization, Apple, Notarizing macOS software before distribution
- **Implemented by**
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
- **Demonstrated by** TestNotarization
- **Read by** Timo

<a id="req-R-DIST-NOTARIZE-LOG"></a>
### R-DIST-NOTARIZE-LOG — Rejections

When Apple rejects a submission, its log MUST be fetched and its reasons shown.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#87-rejections
- **Implemented by**
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
- **Demonstrated by** TestRejectionsShowApplesReasons
- **Read by** Timo

<a id="req-R-DIST-NOTARIZE-SECRETS"></a>
### R-DIST-NOTARIZE-SECRETS — Secrets

Certificates and keys MUST be given as files or environment variables, never as option content, and passwords MUST come from the keychain or environment variables, except that the password of a signing certificate MAY be given as an option, with the help warning that it then ends up in the shell history.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#86-secrets
- **Implemented by**
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
- **Demonstrated by** TestAPIKeyFromTheEnvironment, TestPasswordsComeFromTheEnvironment, TestSecretsComeFromTheEnvironment

<a id="req-R-DIST-PLATFORM"></a>
### R-DIST-PLATFORM — Commands that need macOS

A command that needs Apple's tools MUST explain, wherever they are missing, what is missing and where to run the command instead.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#811-platform
- **Implemented by**
  - [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) — `app/distribution/uc_create_disk_image.go:6–16`
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestWhereAppleToolsAreMissingItSaysSo
- **Read by** Timo

<a id="req-R-DIST-SIGN-ADHOC"></a>
### R-DIST-SIGN-ADHOC — Ad hoc signing

A bundle MUST be signable ad hoc, without a certificate.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#81-ad-hoc-signing
- **Implemented by**
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestAdHocSigning
- **Read by** Timo

<a id="req-R-DIST-SIGN-IDENTITY"></a>
### R-DIST-SIGN-IDENTITY — Developer ID signing

A bundle MUST be signable with a Developer ID given by name or fingerprint, with hardened runtime, secure timestamp and optional entitlements, and a missing identity MUST be answered with the available ones.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#82-developer-id-signing
- **Implemented by**
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestDeveloperIDSigning
- **Read by** Timo

<a id="req-R-DIST-SIGN-ORDER"></a>
### R-DIST-SIGN-ORDER — Signing order

Code in a bundle MUST be signed from the inside out: frameworks, then helpers, then the program, then the bundle.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#83-signing-order
- **Implemented by**
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestSigningOrderIsInsideOut
- **Read by** Timo

<a id="req-R-DIST-SIGN-VERIFY"></a>
### R-DIST-SIGN-VERIFY — Signature verification

A signature MUST be verified after signing, and a signature that does not verify MUST fail the command.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#84-signature-verification
- **Implemented by**
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestSignatureIsVerified
- **Read by** Timo

<a id="req-R-DIST-STAPLE"></a>
### R-DIST-STAPLE — Stapling

After a successful notarization the ticket MUST be stapled to the bundle or disk image.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#88-stapling
- **Implemented by**
  - [NotarizeArtifact](#req-code-github-com-worldiety-wdyapppack-app-distribution-notarizeartifact) — `app/distribution/uc_notarize_artifact.go:6–16`
- **Demonstrated by** TestStapling
- **Read by** Timo

<a id="req-R-DIST-TEMPORARY-KEYCHAIN"></a>
### R-DIST-TEMPORARY-KEYCHAIN — Signing on a build server

When a PKCS#12 certificate is given, base64 encoded in WDYAPPPACK\_CERTIFICATE or as a file by option, with its password from WDYAPPPACK\_CERTIFICATE\_PASSWORD or by option, every command that signs MUST sign from a temporary keychain holding it, which the command creates and deletes itself, also when it fails or is interrupted, and the named identity MUST be one the certificate contains.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#812-signing-on-a-build-server
- **Implemented by**
  - [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) — `app/distribution/uc_create_disk_image.go:6–16`
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestSecretsComeFromTheEnvironment, TestSigningFromACertificateInTheEnvironment, TestTemporaryKeychainSteps

<a id="req-R-DIST-ZIP"></a>
### R-DIST-ZIP — Zip archives

A bundle MUST be wrappable, on every operating system, into a zip archive that keeps symbolic links and executable permissions.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#810-zip-archives
- **Implemented by**
  - [CreateArchive](#req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive) — `app/distribution/uc_create_archive.go:6–16`
- **Demonstrated by** TestLinksAndPermissionsSurvive, TestZipArchive
- **Read by** Timo

<a id="req-R-GUIDE-COMMAND-HELP"></a>
### R-GUIDE-COMMAND-HELP — Help for every command

Every command MUST explain its purpose, each option with its type, default and meaning, examples, the platforms it runs on and related topics.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#33-help-for-every-command
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestCommandHelpComesFromTheDeclaration
- **Read by** Timo

<a id="req-R-GUIDE-DRY-RUN"></a>
### R-GUIDE-DRY-RUN — Dry run

Every command that writes MUST offer a dry run that shows what would be written or called and writes nothing.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#311-trying-without-writing
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [SignBundle](#req-code-github-com-worldiety-wdyapppack-app-distribution-signbundle) — `app/distribution/uc_sign_bundle.go:6–16`
- **Demonstrated by** TestDryRunWritesNothing, TestNotarization
- **Read by** Timo

<a id="req-R-GUIDE-EXIT-CODES"></a>
### R-GUIDE-EXIT-CODES — Exit codes

wdyapppack MUST exit with 0 on success, 1 on a failed operation, 2 on a usage error and 3 when a required tool or platform is missing.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#37-exit-codes
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestExitCodes
- **Read by** Timo

<a id="req-R-GUIDE-HELP-TOPICS"></a>
### R-GUIDE-HELP-TOPICS — Help topics

wdyapppack MUST offer help topics on the bundle layout, the configuration, the Info.plist, icons, programs, universal programs, signing, notarization, distribution, categories, recipes, build servers, AI agents, error codes and exit codes.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#32-help-topics
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.generalTopics`
- **Demonstrated by** TestHelpTopicsCoverTheBackground, TestTopicsAreHelp
- **Read by** Timo

<a id="req-R-GUIDE-JSON"></a>
### R-GUIDE-JSON — Machine readable output

Every command MUST be able to report its result, errors and notes as JSON, and the complete help MUST be available as JSON.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#36-machine-readable-output
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestJSONForResultsErrorsAndHelp
- **Read by** Timo

<a id="req-R-GUIDE-MESSAGES"></a>
### R-GUIDE-MESSAGES — Messages that say what to do

Every error and note MUST state what is wrong, why it matters and how to fix it, carry a stable code and name a help topic, and all problems found MUST be reported together.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#34-messages-that-say-what-to-do
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestMessagesSayWhatWhyAndHow
- **Read by** Timo

<a id="req-R-GUIDE-NEXT-STEPS"></a>
### R-GUIDE-NEXT-STEPS — Next steps

After a successful command wdyapppack MUST name the commands that usually follow.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#38-next-steps
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestNextStepsFollowASuccess
- **Read by** Timo

<a id="req-R-GUIDE-NO-ARGS"></a>
### R-GUIDE-NO-ARGS — Help without arguments

Running wdyapppack without arguments or with the command help MUST print the complete overview and exit successfully.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#31-help-without-arguments
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestNoArgumentsPrintsTheOverview
- **Read by** Timo

<a id="req-R-GUIDE-NO-OVERWRITE"></a>
### R-GUIDE-NO-OVERWRITE — Nothing is overwritten unasked

wdyapppack MUST NOT replace an existing file or directory unless told to with --force.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#312-nothing-is-overwritten-unasked
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [CreateArchive](#req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive) — `app/distribution/uc_create_archive.go:6–16`
  - [CreateDiskImage](#req-code-github-com-worldiety-wdyapppack-app-distribution-creatediskimage) — `app/distribution/uc_create_disk_image.go:6–16`
- **Demonstrated by** TestNothingIsOverwrittenUnasked
- **Read by** Timo

<a id="req-R-GUIDE-REQUIREMENTS"></a>
### R-GUIDE-REQUIREMENTS — Requirements in the help

The requirements MUST be compiled into the program and listed by a help topic.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#39-why-it-behaves-as-it-does
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.requirementsTopic`
- **Demonstrated by** TestRequirementsAreListed
- **Read by** Timo

<a id="req-R-GUIDE-SUGGEST"></a>
### R-GUIDE-SUGGEST — Suggestions for mistakes

A mistyped command, option, help topic, configuration field or category MUST be answered with the closest valid name.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#35-suggestions-for-mistakes
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestMistakesGetSuggestions
- **Read by** Timo

<a id="req-R-GUIDE-VERSION"></a>
### R-GUIDE-VERSION — Version

wdyapppack MUST report its own version.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#310-version
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.versionCommand`
- **Demonstrated by** TestVersion
- **Read by** Timo

<a id="req-R-ICON-FORMATS"></a>
### R-ICON-FORMATS — Accepted images

An icon MUST be accepted as a PNG or JPEG image, an .icns file or an .iconset directory.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#71-accepted-images
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) — `app/bundle/uc_convert_icon.go:6–18`
- **Demonstrated by** TestIconFormats
- **Read by** Timo

<a id="req-R-ICON-ICONSET"></a>
### R-ICON-ICONSET — Icon sets

An image MUST be convertible into an .iconset directory.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#75-icon-sets
- **Implemented by**
  - [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) — `app/bundle/uc_convert_icon.go:6–18`
- **Demonstrated by** TestConvertIconToIcnsAndIconset
- **Read by** Timo

<a id="req-R-ICON-SHAPE"></a>
### R-ICON-SHAPE — Shape for macOS 26

On request an icon image MUST be cut into Apple's rounded square with the margin of Apple's grid; by default it MUST be used unchanged, with a note when its corners are not transparent.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#74-shape-for-macos-26
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) — `app/bundle/uc_convert_icon.go:6–18`
- **Demonstrated by** TestIconShape
- **Read by** Timo

<a id="req-R-ICON-SIZES"></a>
### R-ICON-SIZES — Icon sizes

An image MUST be turned into an .icns file containing every size from 16 to 1024 pixels.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#72-sizes
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) — `app/bundle/uc_convert_icon.go:6–18`
- **Demonstrated by** TestEveryIconSizeInGo, TestIconHasEverySize
- **Read by** Timo

<a id="req-R-ICON-SQUARE"></a>
### R-ICON-SQUARE — Square images

An icon image MUST be square and at least 512 pixels wide, and an image below 1024 pixels MUST be noted.

_functional, business, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#73-square-images
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [ConvertIcon](#req-code-github-com-worldiety-wdyapppack-app-bundle-converticon) — `app/bundle/uc_convert_icon.go:6–18`
- **Demonstrated by** TestIconMustBeSquareAndLarge
- **Read by** Timo

<a id="req-R-NFR-CROSS-PLATFORM-CORE"></a>
### R-NFR-CROSS-PLATFORM-CORE — Core works on every operating system

Building a bundle, converting an icon, merging a universal program and creating a zip archive MUST work on every operating system Go supports.

_non-functional, technical, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#25-works-on-every-operating-system
- **Derived requirements** [R-DEC-PURE-GO-ICNS](#req-R-DEC-PURE-GO-ICNS)
- **Implemented by**
  - [BuildBundle](#req-code-github-com-worldiety-wdyapppack-app-bundle-buildbundle) — `app/bundle/uc_build_bundle.go:6–18`
  - [CreateArchive](#req-code-github-com-worldiety-wdyapppack-app-distribution-createarchive) — `app/distribution/uc_create_archive.go:6–16`
- **Demonstrated by** TestFromAGoProgramToABundleOnAnySystem
- **Read by** Timo

<a id="req-R-NFR-ENGLISH"></a>
### R-NFR-ENGLISH — English

The program, its output, its help and its source code MUST be written in English.

_non-functional, technical, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#22-english
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.describe`
- **Demonstrated by** TestEnglishOnly
- **Read by** Timo

<a id="req-R-NFR-NON-INTERACTIVE"></a>
### R-NFR-NON-INTERACTIVE — Never interactive

wdyapppack MUST NOT ask questions or wait for input; when information is missing it MUST stop and name what is missing and the alternatives.

_non-functional, technical, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#23-never-interactive
- **Implemented by**
  - `github.com/worldiety/wdyapppack/foundation/cli.App`
- **Demonstrated by** TestNeverAsks
- **Read by** Timo

<a id="req-R-NFR-SELF-EXPLAINING"></a>
### R-NFR-SELF-EXPLAINING — Self-explaining

wdyapppack MUST give a person or an AI agent everything needed to use it from the program itself, without documentation outside the program.

_non-functional, mixed, normative._

- **Asked for in** requirements/\_sources/wdyapppack/product.md#21-self-explaining
- **Derived requirements** [R-DEC-HELP-SINGLE-SOURCE](#req-R-DEC-HELP-SINGLE-SOURCE)
- **Implemented by**
  - `github.com/worldiety/wdyapppack/cmd/wdyapppack.describe`
- **Demonstrated by** TestTheProgramExplainsItself
- **Read by** Timo

## Source documents

What people wrote, and what became of each part of it.

### requirements/\_sources/wdyapppack/product.md

| section | became |
|---|---|
| wdyapppack | _nothing, and says so_ |
| 1. Purpose | _nothing, and says so_ |
| 2. Principles | _nothing, and says so_ |
| 2.1 Self-explaining | R-NFR-SELF-EXPLAINING |
| 2.2 English | R-NFR-ENGLISH |
| 2.3 Never interactive | R-NFR-NON-INTERACTIVE |
| 2.4 Standard library only | R-CST-STDLIB-ONLY |
| 2.5 Works on every operating system | R-NFR-CROSS-PLATFORM-CORE |
| 2.6 Apple's tools where only Apple's tools will do | R-CST-APPLE-TOOLS-FOR-SIGNING |
| 3. Guidance | _nothing, and says so_ |
| 3.1 Help without arguments | R-GUIDE-NO-ARGS |
| 3.2 Help topics | R-GUIDE-HELP-TOPICS |
| 3.3 Help for every command | R-DEC-HELP-SINGLE-SOURCE, R-GUIDE-COMMAND-HELP |
| 3.4 Messages that say what to do | R-GUIDE-MESSAGES |
| 3.5 Suggestions for mistakes | R-GUIDE-SUGGEST |
| 3.6 Machine readable output | R-GUIDE-JSON |
| 3.7 Exit codes | R-GUIDE-EXIT-CODES |
| 3.8 Next steps | R-GUIDE-NEXT-STEPS |
| 3.9 Why it behaves as it does | R-GUIDE-REQUIREMENTS |
| 3.10 Version | R-GUIDE-VERSION |
| 3.11 Trying without writing | R-GUIDE-DRY-RUN |
| 3.12 Nothing is overwritten unasked | R-GUIDE-NO-OVERWRITE |
| 4. Configuration | _nothing, and says so_ |
| 4.1 Configuration file | R-CONFIG-FILE, R-DEC-JSON-CONFIG |
| 4.2 Options override the file | R-CONFIG-OVERRIDE |
| 4.3 Unknown fields are refused | R-CONFIG-STRICT |
| 4.4 Environment variables | R-CONFIG-ENV |
| 4.5 Starter configuration | R-CONFIG-INIT |
| 4.6 Schema | R-CONFIG-SCHEMA |
| 4.7 The bundle from the configuration | R-CONFIG-DEFAULT-BUNDLE |
| 5. Building a bundle | _nothing, and says so_ |
| 5.1 Bundle layout | R-BUNDLE-LAYOUT |
| 5.2 Required metadata | R-BUNDLE-METADATA-REQUIRED |
| 5.3 Default metadata | R-BUNDLE-PLIST-DEFAULTS |
| 5.4 Application category | R-BUNDLE-CATEGORY |
| 5.5 Additional metadata | R-BUNDLE-PLIST-CUSTOM |
| 5.6 Protected keys | R-BUNDLE-PLIST-PROTECTED |
| 5.7 Resources, libraries and helpers | R-BUNDLE-PAYLOAD |
| 5.8 Safe writing | R-BUNDLE-SAFE-WRITE |
| 5.9 Reproducible output | R-BUNDLE-REPRODUCIBLE |
| 5.10 Self-check | R-BUNDLE-SELF-CHECK |
| 5.11 Notes and strict mode | R-BUNDLE-STRICT |
| 5.12 Start behaviour | R-DEC-DIRECT-LAUNCH-ONLY |
| 6. Programs | _nothing, and says so_ |
| 6.1 Only macOS programs | R-BINARY-MACHO-ONLY |
| 6.2 Architectures | R-BINARY-ARCH-DETECT |
| 6.3 Minimum macOS version | R-BINARY-MINOS |
| 6.4 Signature on Apple silicon | R-BINARY-ARM64-SIGNED |
| 6.5 Libraries outside the system | R-BINARY-DYLIB-EXTERNAL, R-DEC-NO-LOADPATH-REWRITE |
| 6.6 Universal programs | R-BINARY-UNIVERSAL |
| 7. Icons | _nothing, and says so_ |
| 7.1 Accepted images | R-ICON-FORMATS |
| 7.2 Sizes | R-DEC-PURE-GO-ICNS, R-ICON-SIZES |
| 7.3 Square images | R-ICON-SQUARE |
| 7.4 Shape for macOS 26 | R-ICON-SHAPE |
| 7.5 Icon sets | R-ICON-ICONSET |
| 8. Distribution | _nothing, and says so_ |
| 8.1 Ad hoc signing | R-DIST-SIGN-ADHOC |
| 8.2 Developer ID signing | R-DIST-SIGN-IDENTITY |
| 8.3 Signing order | R-DIST-SIGN-ORDER |
| 8.4 Signature verification | R-DIST-SIGN-VERIFY |
| 8.5 Notarization | R-DIST-NOTARIZE |
| 8.6 Secrets | R-DIST-NOTARIZE-SECRETS |
| 8.7 Rejections | R-DIST-NOTARIZE-LOG |
| 8.8 Stapling | R-DIST-STAPLE |
| 8.9 Disk images | R-DIST-DMG |
| 8.10 Zip archives | R-DIST-ZIP |
| 8.11 Platform | R-DIST-PLATFORM |
| 8.12 Signing on a build server | R-DEC-TEMPORARY-KEYCHAIN, R-DIST-TEMPORARY-KEYCHAIN |
| 8.13 Notarization key from the environment | R-DIST-API-KEY-ENV |
| 8.14 Checksums | R-DIST-CHECKSUM |
| 9. Operation | _nothing, and says so_ |
| 9.1 Who runs it | R-DEC-CLI-RUNS-AS-SYSTEM |
| 10. Out of scope | _nothing, and says so_ |

