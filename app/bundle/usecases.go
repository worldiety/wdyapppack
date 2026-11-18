package bundle

// UseCases are the capabilities of this context.
//
// The bundle takes the weight out of the place a context is assembled. It is
// deliberately not what a command takes: a command takes the use cases it
// calls, so that it can be tested with exactly those.
type UseCases struct {
	LoadConfig     LoadConfig
	BuildBundle    BuildBundle
	ConvertIcon    ConvertIcon
	MergeUniversal MergeUniversal
	WriteConfig    WriteConfig
}

// NewUseCases wires the context over its port.
func NewUseCases(files Files) UseCases {
	return UseCases{
		LoadConfig:     NewLoadConfig(files),
		BuildBundle:    NewBuildBundle(files),
		ConvertIcon:    NewConvertIcon(files),
		MergeUniversal: NewMergeUniversal(files),
		WriteConfig:    NewWriteConfig(files),
	}
}
