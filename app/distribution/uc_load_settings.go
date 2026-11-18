package distribution

import "github.com/worldiety/wdyapppack/foundation/auth"

// LoadSettings reads the distribution sections of the configuration file.
type LoadSettings func(subject auth.Subject, req LoadSettingsRequest) (LoadedSettings, error)

// NewLoadSettings builds the use case over its ports.
func NewLoadSettings(ws Workspace) LoadSettings {
	return func(subject auth.Subject, req LoadSettingsRequest) (LoadedSettings, error) {
		if err := subject.Audit(PermLoadSettings); err != nil {
			return LoadedSettings{}, err
		}
		return loadSettings(ws, req)
	}
}
