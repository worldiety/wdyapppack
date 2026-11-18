package distribution

import "github.com/worldiety/wdyapppack/foundation/auth"

// NotarizeArtifact has a bundle, disk image or archive notarized by Apple and staples the ticket.
type NotarizeArtifact func(subject auth.Subject, req NotarizeRequest) (Notarized, error)

// NewNotarizeArtifact builds the use case over its ports.
func NewNotarizeArtifact(ws Workspace, signer Signer, notary Notary) NotarizeArtifact {
	return func(subject auth.Subject, req NotarizeRequest) (Notarized, error) {
		if err := subject.Audit(PermNotarizeArtifact); err != nil {
			return Notarized{}, err
		}
		return notarize(ws, signer, notary, req)
	}
}
