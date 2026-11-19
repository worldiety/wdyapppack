# wdyapppack

<!-- speclink:informative -->

This document describes what wdyapppack must do. It is the material the
requirements cite. Sections that carry no obligation are marked informative;
every other section is turned into at least one requirement.

## 1. Purpose

<!-- speclink:informative -->

wdyapppack packages programs that are already compiled — typically written in
Go, Rust, Zig or C — as macOS application bundles (`.app`). A bundle is what
macOS expects of an application: it can be dragged into the Applications
folder, shows a name and an icon in the Finder, the Dock and Launchpad, and
carries its metadata in an `Info.plist`.

Beyond creating the bundle, wdyapppack signs it, has it notarized by Apple and
wraps it into a disk image or a zip archive, so that it can be handed to other
people.

## 2. Principles

<!-- speclink:informative -->

These principles hold for every command.

### 2.1 Self-explaining

A person or an AI agent must be able to use wdyapppack successfully without
reading any documentation outside the program. Everything needed to reach the
goal — what the tool does, which commands exist, what every option means, what
went wrong and how to fix it — is available from the program itself.

### 2.2 English

The program, its output, its help and its source code are written in English.

### 2.3 Never interactive

wdyapppack never asks a question and never waits for input. When information is
missing, it stops and says what is missing and which alternatives exist. It is
often run by scripts, build servers and AI agents, for which a prompt is a
dead end.

### 2.4 Standard library only

wdyapppack is written in Go and uses nothing but the Go standard library, with
the single exception of the speclink specification package. Every additional
dependency is code nobody here reviewed, and a packaging tool sits in the
supply chain of every program it packages.

### 2.5 Works on every operating system

Creating a bundle, converting an icon, merging programs into a universal
program and creating a zip archive work on every operating system Go supports,
including Linux build servers. None of these tasks needs anything from Apple.

### 2.6 Apple's tools where only Apple's tools will do

Code signing, notarization and disk images are done by calling the tools Apple
ships with macOS (`codesign`, `xcrun notarytool`, `xcrun stapler`, `hdiutil`).
These commands are therefore only available on macOS. Reimplementing them
would mean reimplementing Apple's cryptographic formats and services.

## 3. Guidance

<!-- speclink:informative -->

How wdyapppack talks to the person or program using it.

### 3.1 Help without arguments

Running wdyapppack without any arguments, or with the command `help`, prints a
complete overview: what the tool does, a quick start, what a bundle looks like,
every command with a one-line summary, every help topic, typical recipes,
global options, exit codes and a hint for AI agents. This is a successful run,
not an error.

### 3.2 Help topics

Background knowledge that does not belong to one command is available as help
topics, for example on the bundle layout, the configuration file, the
`Info.plist`, icons, programs, universal programs, signing, notarization,
distribution, application categories, recipes, build servers, AI agents, error
codes and exit codes.

### 3.3 Help for every command

Every command explains itself: its purpose, every option with its type, default
and meaning, examples, the platforms it runs on, and related topics. The help is
generated from the same declaration that defines the options, so the help can
never describe an option the command does not have, or miss one it does.

### 3.4 Messages that say what to do

Every error and every note states what is wrong, why it matters and how to fix
it, carries a stable code and names the help topic to read next. When several
things are wrong at once, all of them are reported together rather than one per
run.

### 3.5 Suggestions for mistakes

A mistyped command, option, help topic, configuration field or category is
answered with the closest valid name.

### 3.6 Machine readable output

Every command can report its result, its errors and its notes as JSON, and the
complete help — commands, options, topics, error codes, exit codes and
configuration fields — is available as JSON as well.

### 3.7 Exit codes

The exit code tells a script what happened: success, a failed operation, a
usage error, or a missing tool or platform. The meaning of each code is fixed
and documented in the help.

### 3.8 Next steps

After a successful command, wdyapppack names the commands that usually follow,
so that a person who has never packaged an application is led from the program
to a file they can hand out.

### 3.9 Why it behaves as it does

The requirements in this document are compiled into the program and can be
listed from the help, so that a user can find out why the tool behaves as it
does.

### 3.10 Version

wdyapppack reports its own version.

### 3.11 Trying without writing

Every command that writes something can be run as a dry run, which shows what
would be written or which Apple tool would be called, and writes nothing.

### 3.12 Nothing is overwritten unasked

wdyapppack never replaces an existing file or directory unless explicitly told
to.

## 4. Configuration

<!-- speclink:informative -->

How a bundle is described.

### 4.1 Configuration file

Everything a build needs can be written into a configuration file in JSON, so
that a build can be repeated, reviewed and run on a build server. Relative
paths in the file are relative to the file.

### 4.2 Options override the file

Every value in the configuration file can also be given as a command line
option. An option given on the command line wins over the file.

### 4.3 Unknown fields are refused

A configuration file containing a field wdyapppack does not know is refused,
because a misspelled field silently ignored is a setting the user believes in
and does not have.

### 4.4 Environment variables

Text values in the configuration file may refer to environment variables, so
that values such as a version number can come from a build server. A reference
to a variable that is not set is an error.

### 4.5 Starter configuration

wdyapppack can write a complete starter configuration. Whatever can be learned
from the program itself is filled in; whatever cannot is filled with a
placeholder that a build refuses until it is replaced.

### 4.6 Schema

wdyapppack can print a JSON Schema of the configuration file, including a
description of every field, so that editors can complete and check it.

### 4.7 The bundle from the configuration

The commands that work on a finished bundle — signing, notarizing and
wrapping it — use the bundle the configuration file describes when no bundle is
named, so that the steps of a build server need no paths.

## 5. Building a bundle

<!-- speclink:informative -->

The central task.

### 5.1 Bundle layout

A bundle has the layout macOS expects: `Contents/Info.plist`,
`Contents/PkgInfo`, the program in `Contents/MacOS`, the icon and other
resources in `Contents/Resources`, libraries in `Contents/Frameworks` and
further programs in `Contents/Helpers`. The program is executable.

### 5.2 Required metadata

A bundle needs a name, an identifier and a version. The identifier is in
reverse domain notation and contains only letters, digits, dots and hyphens.
The version consists of one to three numbers separated by dots. A build without
these, or with malformed ones, is refused.

### 5.3 Default metadata

wdyapppack writes every `Info.plist` key an application needs, with sensible
defaults: package type, executable, display name, build number, icon file,
supported platform, development region, minimum macOS version and support for
high resolution displays. The user can set the copyright, whether the
application shows in the Dock, whether it runs only in the background,
environment variables, URL schemes and the texts macOS shows when the
application asks for access to the camera, the microphone and similar.

### 5.4 Application category

The application category is checked against the categories Apple defines. It
may be given in its full form or as the short name.

### 5.5 Additional metadata

Any further `Info.plist` key can be set, with its type, either as an option or
from a partial property list file that is merged in.

### 5.6 Protected keys

Keys that wdyapppack derives from the bundle itself, such as the name of the
executable or the package type, cannot be overwritten by additional metadata,
because a bundle whose metadata contradicts its contents does not start.

### 5.7 Resources, libraries and helpers

Files and directories can be added as resources, libraries as frameworks and
further programs as helpers. Helpers are checked like the main program.

### 5.8 Safe writing

A bundle is assembled next to its destination and moved into place only when it
is complete, so that an interrupted build never leaves a half written bundle
behind.

### 5.9 Reproducible output

The same input produces the same bundle, byte for byte: keys are written in a
fixed order and file times can be fixed.

### 5.10 Self-check

After writing a bundle, wdyapppack reads it back and checks it: the
`Info.plist` is readable and complete, the program exists and is executable,
and the icon it names exists.

### 5.11 Notes and strict mode

Problems that do not prevent a working bundle are reported as notes and do not
fail the build. A strict mode turns every note into an error, for build servers
that must not let anything slip.

### 5.12 Start behaviour

The program inside the bundle is started directly by macOS. wdyapppack does not
wrap programs without a window of their own in a terminal; the help explains
that such a program runs invisibly when the bundle is opened.

## 6. Programs

<!-- speclink:informative -->

What wdyapppack checks about the program it packages.

### 6.1 Only macOS programs

Only macOS programs (Mach-O) can be packaged. A program built for Linux or
Windows is refused with an explanation of how to build it for macOS.

### 6.2 Architectures

wdyapppack recognises which processor architectures a program is built for and
refuses a program whose stated architecture does not match its contents or
which is given twice.

### 6.3 Minimum macOS version

If no minimum macOS version is given, it is taken from the program, and if the
program does not state one, macOS 11 is assumed. A minimum version lower than
the one the program was built for is refused.

### 6.4 Signature on Apple silicon

macOS refuses to start a program for Apple silicon that carries no signature at
all. wdyapppack notes such a program and explains how to sign it.

### 6.5 Libraries outside the system

A program that loads libraries from outside the system, for example from a
package manager, does not run on a Mac without those libraries. wdyapppack
notes every such library and explains how to bundle it. It does not rewrite the
paths a program loads its libraries from.

### 6.6 Universal programs

Programs for Apple silicon and for Intel can be merged into one universal
program, either on their own or while building a bundle. Signatures of the
individual programs are kept.

## 7. Icons

<!-- speclink:informative -->

How the application gets its icon.

### 7.1 Accepted images

An icon can be given as a PNG or JPEG image, as an existing `.icns` file or as
an `.iconset` directory.

### 7.2 Sizes

From an image, wdyapppack produces every size macOS uses, from 16 to 1024
pixels, in an `.icns` file, without any tool from Apple.

### 7.3 Square images

An icon image must be square. An image smaller than 512 pixels is refused; one
smaller than 1024 pixels is noted.

### 7.4 Shape for macOS 26

Since macOS 26, icons that do not follow Apple's rounded square shape are shown
inside a grey frame. On request, wdyapppack cuts the image into that shape with
the margin Apple's grid prescribes. By default the image is used unchanged, and
a note points to the option when the corners of the image are not transparent.

### 7.5 Icon sets

An image can also be converted into an `.iconset` directory for use with
Apple's tools.

## 8. Distribution

<!-- speclink:informative -->

How a bundle is handed to other people.

### 8.1 Ad hoc signing

A bundle can be signed without a certificate (ad hoc). Such a bundle starts on
the Mac it was built on and on Macs that do not check the origin of
applications.

### 8.2 Developer ID signing

A bundle can be signed with a Developer ID certificate, given by name or by its
fingerprint, with the hardened runtime, a secure timestamp and optional
entitlements. If the certificate is not available, the available ones are
listed.

### 8.3 Signing order

Code inside a bundle is signed from the inside out: libraries, then helpers,
then the program, then the bundle.

### 8.4 Signature verification

After signing, the signature is verified, and a bundle whose signature does not
verify is reported as a failure.

### 8.5 Notarization

A bundle, a disk image or a zip archive signed with a Developer ID can be
submitted to Apple for notarization, and wdyapppack waits for the result. A
bundle that is not signed for notarization is refused before anything is sent.

### 8.6 Secrets

Certificates and keys are given as files or through environment variables,
never as the content of an option. Passwords are read from the keychain or from
environment variables, because option values end up in the shell history and
the process list. The one exception is the password of a signing certificate:
it may also be typed as an option at a terminal, and the help says that it then
ends up in the shell history and names the environment variable that keeps it
out.

### 8.7 Rejections

When Apple rejects a submission, wdyapppack fetches Apple's log and shows the
reasons.

### 8.8 Stapling

After a successful notarization, the ticket is attached to the bundle or the
disk image, so that it can be checked without a network connection.

### 8.9 Disk images

A bundle can be wrapped into a disk image, compressed as far as the format
allows, that contains a link to the Applications folder, so that installing
means dragging one icon onto another.

### 8.10 Zip archives

A bundle can be wrapped into a zip archive that keeps symbolic links and
executable permissions, on every operating system.

### 8.11 Platform

A command that needs Apple's tools explains, on any other operating system or
when a tool is missing, what is missing and where to run the command instead.

### 8.12 Signing on a build server

A build server has no keychain with the signing certificate in it, and
preparing one is a sequence of steps every project would otherwise script on
its own. When the environment variable WDYAPPPACK_CERTIFICATE holds a
Developer ID certificate as a base64 encoded PKCS#12 file, and
WDYAPPPACK_CERTIFICATE_PASSWORD its password, every command that signs creates
a temporary keychain for itself, imports the certificate, signs with it and
deletes the keychain again — also when the command fails or is interrupted. At
a terminal, the same works with options: the certificate as a .p12 file and its
password given on the command line; an option wins over the environment.
Nothing has to be prepared before and nothing cleaned up after. The identity to
sign with is still named explicitly, and it must be one the certificate
contains.

### 8.13 Notarization key from the environment

When no other way to reach the notary service is given, an App Store Connect
API key is taken from the environment: the key itself from WDYAPPPACK_API_KEY,
base64 encoded or as the text of the .p8 file, its ID from
WDYAPPPACK_API_KEY_ID and, for a team key, its issuer from
WDYAPPPACK_API_ISSUER. The key is written to a private temporary file only for
as long as the command runs.

### 8.14 Checksums

Next to every disk image and zip archive, a SHA-256 checksum file is written in
the format of shasum -a 256, so that whoever downloads it can check it.

## 9. Operation

<!-- speclink:informative -->

How wdyapppack runs.

### 9.1 Who runs it

wdyapppack acts with the rights of the operating system user who runs it. It
has no accounts and no permissions of its own.

## 10. Out of scope

<!-- speclink:informative -->

Not part of wdyapppack: compiling source code, installing bundles into the
Applications folder, inspecting or checking bundles made elsewhere, rewriting
library paths, Mac App Store packages, installer packages, sandbox profiles,
asset catalogues, decorated disk images, and bundles for iOS, iPadOS or tvOS.
