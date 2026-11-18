package bundle

import "github.com/worldiety/wdyapppack/foundation/auth"

// MergeUniversal merges programs for different processors into one universal program.
type MergeUniversal func(subject auth.Subject, req MergeRequest) (Merged, error)

// NewMergeUniversal builds the use case over its port.
func NewMergeUniversal(files Files) MergeUniversal {
	return func(subject auth.Subject, req MergeRequest) (Merged, error) {
		// The permission is audited before anything is read or written, so
		// that a refusal leaves no trace of the work it refused.
		if err := subject.Audit(PermMergeUniversal); err != nil {
			return Merged{}, err
		}
		return mergeUniversal(files, req)
	}
}
