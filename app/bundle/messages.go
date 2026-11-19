package bundle

import (
	"cmp"
	"slices"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// explanation is the part of a message that does not depend on the case: why
// it matters, how it is usually fixed, and where to read more.
type explanation struct {
	summary string
	why     string
	how     string
	see     string
}

// explanations lists every message this context can produce. The help topic
// on error codes is generated from it, and a test checks that every code used
// in the code is listed here.
func explanations() map[string]explanation {
	return map[string]explanation{
		"WAP-CONFIG-MISSING": {"the configuration file named does not exist",
			"A file named with --config must exist; only the default wdyapppack.json may be absent.",
			"Check the path, or write a starter configuration with: wdyapppack init", "config"},
		"WAP-CONFIG-SYNTAX": {"the configuration file is not valid JSON",
			"The file cannot be read at all until the syntax is fixed. JSON allows no comments and no trailing commas, and every key needs double quotes.",
			"Fix the file at the line named.", "config"},
		"WAP-CONFIG-FIELD-UNKNOWN": {"the configuration file contains a field wdyapppack does not know",
			"A misspelled field would be silently ignored, and the bundle would lack a setting you believe it has.",
			"Correct or remove the field. Every field is listed by: wdyapppack help config", "config"},
		"WAP-CONFIG-TYPE": {"a field of the configuration file has the wrong type",
			"The value cannot be used as it is written, for example a number where text is expected.",
			"Write the value with the type listed by: wdyapppack help config", "config"},
		"WAP-CONFIG-ENV-UNSET": {"the configuration refers to an environment variable that is not set",
			"Leaving ${NAME} in place or replacing it with nothing would produce a bundle with a wrong value, which is worse than none.",
			"Set the variable before running wdyapppack, or write the value into the file.", "config"},
		"WAP-CONFIG-PLACEHOLDER": {"a value still holds the placeholder written by init",
			"The placeholder marks a value that cannot be guessed and has to be chosen by you.",
			"Replace it in the configuration file, or give the value as an option.", "config"},
		"WAP-NAME-MISSING": {"the bundle has no name",
			"The name is shown in the Finder and the menu bar, and the bundle directory is named after it.",
			"Give it with --name, or as \"name\" in the configuration file.", "bundle"},
		"WAP-NAME-INVALID": {"the name cannot be used as the name of a bundle",
			"The bundle directory is named after it, and macOS does not allow / or : in file names; a leading dot would hide the bundle.",
			"Choose a name without / and :, not starting with a dot.", "bundle"},
		"WAP-NAME-LONG": {"the name is longer than 15 characters",
			"macOS shows CFBundleName in the menu bar and may shorten a longer one.",
			"Use a shorter name, and put the long one into --display-name.", "info-plist"},
		"WAP-IDENTIFIER-MISSING": {"the bundle has no identifier",
			"macOS tells applications apart by their identifier: preferences, permissions and signatures are attached to it.",
			"Give one in reverse domain notation with --identifier, e.g. com.example.mytool.", "info-plist"},
		"WAP-IDENTIFIER-INVALID": {"the identifier is not in reverse domain notation",
			"An identifier may contain only letters, digits, dots and hyphens, and needs at least one dot; codesign and the notary service refuse anything else.",
			"Use the form com.example.mytool.", "info-plist"},
		"WAP-VERSION-MISSING": {"the bundle has no version",
			"The version is shown in the Finder and in About, and updaters compare it.",
			"Give it with --version, e.g. 1.2.0.", "info-plist"},
		"WAP-VERSION-INVALID": {"a version is not one to three numbers separated by dots",
			"macOS compares versions number by number; text such as v1.2 or 1.2-beta cannot be compared.",
			"Write it as 1, 1.2 or 1.2.3.", "info-plist"},
		"WAP-BINARY-MISSING": {"no program was given",
			"A bundle exists to start a program; without one there is nothing to package.",
			"Give the program with --binary path/to/program, or under \"binaries\" in the configuration file.", "binaries"},
		"WAP-FILE-NOT-FOUND": {"a file named as input does not exist",
			"Everything a bundle is made of has to be there when it is built.",
			"Check the path. Paths in the configuration file are relative to the file, paths given as options to the current directory.", "config"},
		"WAP-FILE-UNREADABLE": {"a file could not be read",
			"The operating system refused to hand over its content.",
			"Check that the file is readable by you, and that it is a file and not a directory.", "config"},
		"WAP-BINARY-NOT-MACHO": {"a program is not a macOS program",
			"macOS runs only Mach-O programs. This one was built for another operating system, or it is not a program at all.",
			"Build it for macOS. For Go: GOOS=darwin GOARCH=arm64 go build (and GOARCH=amd64 for Intel Macs).", "binaries"},
		"WAP-BINARY-NOT-EXECUTABLE": {"a program is a library or another kind of Mach-O file, not an executable",
			"Only an executable can be started; a library has to be loaded by one.",
			"Give the executable as --binary, and put libraries in with --framework.", "binaries"},
		"WAP-ARCH-UNKNOWN": {"an architecture name is not known",
			"wdyapppack packages programs for Apple silicon (arm64) and Intel (x86_64, also written amd64).",
			"Write arm64 or amd64, or leave the architecture out: it is read from the program.", "architectures"},
		"WAP-ARCH-MISMATCH": {"the architecture stated for a program is not the one it is built for",
			"The two programs were probably swapped, and the bundle would not run where you expect it to.",
			"Correct the arch= prefix or the arch field, or leave it out: it is read from the program.", "architectures"},
		"WAP-ARCH-DUPLICATE": {"two programs are built for the same architecture",
			"A universal program holds one program per architecture; macOS could not choose between two.",
			"Give one program per architecture.", "architectures"},
		"WAP-HELPER-ARCH-MISSING": {"a helper is not built for every architecture of the program",
			"The program starts its helper on every Mac it runs on; on an architecture the helper lacks, starting it fails, although the bundle opens.",
			"Give the helper once per architecture of the program under \"binaries\", or with --helper arch=path; they are merged into one universal helper.", "architectures"},
		"WAP-HELPER-DUPLICATE": {"two helpers have the same name",
			"Both would be written to the same file in Contents/Helpers, and one would silently replace the other.",
			"Give one of them another name with the name field, or list the programs of one helper together under its \"binaries\".", "binaries"},
		"WAP-ARM64-UNSIGNED": {"a program for Apple silicon carries no signature",
			"macOS refuses to start an arm64 program that has no signature at all. Go adds an ad hoc signature itself; other toolchains may not.",
			"After building, run: wdyapppack sign <bundle> --adhoc (or with --identity for distribution).", "signing"},
		"WAP-DYLIB-EXTERNAL": {"a program loads a library from outside macOS",
			"The library exists on this machine, for example from Homebrew, but not on other Macs, where the program will not start.",
			"Link the library statically, or copy it in with --framework and make the program load it from @rpath or @executable_path/../Frameworks when you build it (install_name_tool). wdyapppack does not rewrite programs.", "binaries"},
		"WAP-MINOS-INVALID": {"the minimum macOS version is not a version",
			"macOS compares it number by number.",
			"Write it like 11.0 or 13.5.", "info-plist"},
		"WAP-MINOS-TOO-LOW": {"the minimum macOS version is lower than the program needs",
			"The program was built for a newer macOS and would refuse to start on the older ones the bundle claims to support.",
			"Leave the minimum version out, so that the program's own is used, or raise it.", "binaries"},
		"WAP-CATEGORY-UNKNOWN": {"the application category is not one Apple defines",
			"macOS and the App Store only know Apple's categories; any other value is ignored.",
			"Use one of the values listed by: wdyapppack help categories", "categories"},
		"WAP-USAGE-UNKNOWN": {"a usage description names an unknown kind of access",
			"Each kind maps to one Info.plist key that macOS reads before it asks the user for permission.",
			"Use one of the kinds listed by: wdyapppack help info-plist, or set the key itself with --plist.", "info-plist"},
		"WAP-URL-SCHEME-INVALID": {"a URL scheme is not valid",
			"A scheme starts with a letter and contains only letters, digits, +, - and dots.",
			"Write the scheme without :// , e.g. mytool.", "info-plist"},
		"WAP-PLIST-PROTECTED": {"additional metadata tries to set a key wdyapppack manages",
			"These keys follow from the bundle itself or have options of their own; a different value would make the metadata contradict the contents, and the bundle would not start.",
			"Use the option named, or remove the key.", "info-plist"},
		"WAP-PLIST-FILE": {"the partial property list cannot be used",
			"It is merged into Info.plist and has to be an XML property list whose top level is a dictionary.",
			"Fix the file; a binary one can be converted with: plutil -convert xml1 <file>", "info-plist"},
		"WAP-PLIST-VALUE": {"an additional Info.plist value cannot be used",
			"Property lists hold text, numbers, true/false, dates, data, arrays and dictionaries; null and other values have no equivalent.",
			"Write --plist Key=text, Key:bool=true, Key:int=3, Key:real=1.5 or Key:array=a,b.", "info-plist"},
		"WAP-ICON-FORMAT": {"the icon is not in a format wdyapppack reads",
			"An icon is made from a PNG or JPEG image, or taken from an .icns file or an .iconset directory.",
			"Convert the image to PNG, ideally 1024×1024 pixels with a transparent background.", "icons"},
		"WAP-ICON-NOT-SQUARE": {"the icon image is not square",
			"macOS icons are square; stretching the image would distort it, cropping would cut it.",
			"Make the image square, ideally 1024×1024 pixels.", "icons"},
		"WAP-ICON-TOO-SMALL": {"the icon image is smaller than 512 pixels",
			"The icon is shown at up to 1024 pixels on Retina displays; an image this small would look blurred everywhere.",
			"Use an image of 1024×1024 pixels.", "icons"},
		"WAP-ICON-SMALL": {"the icon image is smaller than 1024 pixels",
			"The largest size macOS shows is 1024 pixels; it is scaled up from the image and will look soft.",
			"Use an image of 1024×1024 pixels if you have one.", "icons"},
		"WAP-ICON-SHAPE": {"the icon's corners are not transparent",
			"Since macOS 26, an icon that does not follow Apple's rounded square is shown inside a grey frame.",
			"Give the icon the shape yourself, or let wdyapppack cut it with --icon-shape squircle.", "icons"},
		"WAP-ICON-SHAPE-ICNS": {"a shape was requested for an icon that is not an image",
			"The shape is cut out of an image; an .icns file or an icon set is used as it is.",
			"Give a PNG or JPEG image, or leave --icon-shape at none.", "icons"},
		"WAP-ICON-NONE": {"the bundle has no icon",
			"macOS shows a generic application icon instead.",
			"Give an image with --icon, ideally a 1024×1024 PNG.", "icons"},
		"WAP-ICONSET-INVALID": {"the icon set or .icns file cannot be used",
			"An icon set contains PNG files with the names and sizes iconutil defines; an .icns file starts with \"icns\".",
			"Fix the files named, or give a single 1024×1024 PNG instead.", "icons"},
		"WAP-OUTPUT-EXISTS": {"the output exists already",
			"Nothing is replaced unless you ask for it.",
			"Add --force to replace it, or choose another output.", "bundle"},
		"WAP-WRITE-FAILED": {"the output could not be written",
			"The operating system refused to create or move a file.",
			"Check that the output directory is writable and the disk is not full.", "bundle"},
		"WAP-SELF-CHECK": {"the bundle written does not pass its own check",
			"What was read back is not what was meant to be written, so the bundle would not work.",
			"Run the build again with --verbose. If it persists, report it with the command you ran.", "bundle"},
		"WAP-UNIVERSAL-INPUTS": {"a universal program needs at least two programs",
			"It combines programs for different architectures; with one there is nothing to combine.",
			"Give one program per architecture, e.g. --in arm64=prog-arm64 --in amd64=prog-amd64.", "architectures"},
	}
}

// msg builds a message from its code, with a case specific description.
func msg(code, what string) problem.Message {
	e, ok := explanations()[code]
	if !ok {
		panic("bundle: message code " + code + " is not explained")
	}
	return problem.Message{Code: code, Severity: problem.Error, What: what, Why: e.why, How: e.how, See: e.see}
}

// msgHow builds a message with a case specific remedy.
func msgHow(code, what, how string) problem.Message {
	m := msg(code, what)
	m.How = how
	return m
}

// Codes documents every message of this context.
func Codes() []problem.Doc {
	var out []problem.Doc
	for code, e := range explanations() {
		out = append(out, problem.Doc{Code: code, Summary: e.summary, See: e.see})
	}
	slices.SortFunc(out, func(a, b problem.Doc) int { return cmp.Compare(a.Code, b.Code) })
	return out
}
