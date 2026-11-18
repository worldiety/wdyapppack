package distribution

import "github.com/worldiety/wdyapppack/foundation/auth"

// SignBundle signs a bundle from the inside out and verifies the signature.
type SignBundle func(subject auth.Subject, req SignRequest) (Signed, error)

// NewSignBundle builds the use case over its ports.
func NewSignBundle(ws Workspace, signer Signer) SignBundle {
	return func(subject auth.Subject, req SignRequest) (Signed, error) {
		if err := subject.Audit(PermSignBundle); err != nil {
			return Signed{}, err
		}
		return signBundle(ws, signer, req)
	}
}
