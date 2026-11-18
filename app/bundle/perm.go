package bundle

import "github.com/worldiety/wdyapppack/foundation/permission"

// Permissions are declared once per use case, at package level, so that the
// set of them is fixed when the binary is linked. On the command line every
// use case runs as the system subject, which holds all of them: what the tool
// may do is decided by the operating system user who runs it.
var (
	PermLoadConfig     = permission.Declare[LoadConfig]("bundle.config.load", "Read configuration", "Holders may read a configuration file.")
	PermBuildBundle    = permission.Declare[BuildBundle]("bundle.build", "Build bundle", "Holders may create an application bundle.")
	PermConvertIcon    = permission.Declare[ConvertIcon]("bundle.icon.convert", "Convert icon", "Holders may make an icon from an image.")
	PermMergeUniversal = permission.Declare[MergeUniversal]("bundle.universal.merge", "Merge programs", "Holders may merge programs into a universal program.")
	PermWriteConfig    = permission.Declare[WriteConfig]("bundle.config.write", "Write configuration", "Holders may write a starter configuration file.")
)
