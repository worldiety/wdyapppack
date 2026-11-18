package bundle

import "github.com/worldiety/wdyapppack/foundation/auth"

// WriteConfig writes a starter configuration file.
type WriteConfig func(subject auth.Subject, req WriteConfigRequest) (WrittenConfig, error)

// NewWriteConfig builds the use case over its port.
func NewWriteConfig(files Files) WriteConfig {
	return func(subject auth.Subject, req WriteConfigRequest) (WrittenConfig, error) {
		// The permission is audited before anything is read or written, so
		// that a refusal leaves no trace of the work it refused.
		if err := subject.Audit(PermWriteConfig); err != nil {
			return WrittenConfig{}, err
		}
		return writeConfig(files, req)
	}
}
