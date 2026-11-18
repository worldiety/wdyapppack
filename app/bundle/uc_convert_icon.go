package bundle

import "github.com/worldiety/wdyapppack/foundation/auth"

// ConvertIcon makes a macOS icon from an image.
type ConvertIcon func(subject auth.Subject, req ConvertIconRequest) (ConvertedIcon, error)

// NewConvertIcon builds the use case over its port.
func NewConvertIcon(files Files) ConvertIcon {
	return func(subject auth.Subject, req ConvertIconRequest) (ConvertedIcon, error) {
		// The permission is audited before anything is read or written, so
		// that a refusal leaves no trace of the work it refused.
		if err := subject.Audit(PermConvertIcon); err != nil {
			return ConvertedIcon{}, err
		}
		return convertIcon(files, req)
	}
}
