package distribution

// UseCases are the capabilities of this context.
type UseCases struct {
	LoadSettings     LoadSettings
	SignBundle       SignBundle
	NotarizeArtifact NotarizeArtifact
	CreateDiskImage  CreateDiskImage
	CreateArchive    CreateArchive
}

// NewUseCases wires the context over its ports.
func NewUseCases(ws Workspace, signer Signer, notary Notary, imager DiskImager) UseCases {
	return UseCases{
		LoadSettings:     NewLoadSettings(ws),
		SignBundle:       NewSignBundle(ws, signer),
		NotarizeArtifact: NewNotarizeArtifact(ws, signer, notary),
		CreateDiskImage:  NewCreateDiskImage(ws, imager, signer),
		CreateArchive:    NewCreateArchive(ws),
	}
}
