package distribution

import "github.com/worldiety/wdyapppack/foundation/auth"

// CreateArchive wraps a bundle into a zip archive that keeps links and permissions.
type CreateArchive func(subject auth.Subject, req ArchiveRequest) (Archive, error)

// NewCreateArchive builds the use case over its ports.
func NewCreateArchive(ws Workspace) CreateArchive {
	return func(subject auth.Subject, req ArchiveRequest) (Archive, error) {
		if err := subject.Audit(PermCreateArchive); err != nil {
			return Archive{}, err
		}
		return createArchive(ws, req)
	}
}
