package distribution

import "github.com/worldiety/wdyapppack/foundation/auth"

// CreateDiskImage wraps a bundle into a compressed disk image with a link to the Applications folder.
type CreateDiskImage func(subject auth.Subject, req DiskImageRequest) (DiskImage, error)

// NewCreateDiskImage builds the use case over its ports.
func NewCreateDiskImage(ws Workspace, imager DiskImager, signer Signer) CreateDiskImage {
	return func(subject auth.Subject, req DiskImageRequest) (DiskImage, error) {
		if err := subject.Audit(PermCreateDiskImage); err != nil {
			return DiskImage{}, err
		}
		return createDiskImage(ws, imager, signer, req)
	}
}
