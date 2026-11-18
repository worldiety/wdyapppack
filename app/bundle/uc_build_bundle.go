package bundle

import "github.com/worldiety/wdyapppack/foundation/auth"

// BuildBundle creates an application bundle from a prebuilt program.
type BuildBundle func(subject auth.Subject, req BuildRequest) (BuildResult, error)

// NewBuildBundle builds the use case over its port.
func NewBuildBundle(files Files) BuildBundle {
	return func(subject auth.Subject, req BuildRequest) (BuildResult, error) {
		// The permission is audited before anything is read or written, so
		// that a refusal leaves no trace of the work it refused.
		if err := subject.Audit(PermBuildBundle); err != nil {
			return BuildResult{}, err
		}
		return build(files, req)
	}
}
