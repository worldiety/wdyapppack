package bundle

import "github.com/worldiety/wdyapppack/foundation/auth"

// LoadConfig reads the configuration file strictly, expands environment variables and resolves its paths.
type LoadConfig func(subject auth.Subject, req LoadConfigRequest) (LoadedConfig, error)

// NewLoadConfig builds the use case over its port.
func NewLoadConfig(files Files) LoadConfig {
	return func(subject auth.Subject, req LoadConfigRequest) (LoadedConfig, error) {
		// The permission is audited before anything is read or written, so
		// that a refusal leaves no trace of the work it refused.
		if err := subject.Audit(PermLoadConfig); err != nil {
			return LoadedConfig{}, err
		}
		return loadConfig(files, req)
	}
}
