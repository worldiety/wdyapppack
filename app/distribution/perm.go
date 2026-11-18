package distribution

import "github.com/worldiety/wdyapppack/foundation/permission"

// Permissions are declared once per use case. On the command line every use
// case runs as the system subject: what the tool may do is decided by the
// operating system user who runs it, and by the keychain that user can open.
var (
	PermLoadSettings     = permission.Declare[LoadSettings]("distribution.settings.load", "Read distribution settings", "Holders may read the distribution sections of a configuration file.")
	PermSignBundle       = permission.Declare[SignBundle]("distribution.sign", "Sign bundle", "Holders may sign a bundle with an identity from the keychain.")
	PermNotarizeArtifact = permission.Declare[NotarizeArtifact]("distribution.notarize", "Notarize", "Holders may submit a bundle, disk image or archive to Apple's notary service.")
	PermCreateDiskImage  = permission.Declare[CreateDiskImage]("distribution.dmg.create", "Create disk image", "Holders may create a disk image of a bundle.")
	PermCreateArchive    = permission.Declare[CreateArchive]("distribution.zip.create", "Create archive", "Holders may create a zip archive of a bundle.")
)
