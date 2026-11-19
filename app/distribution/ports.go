package distribution

// Signer signs code and describes signatures. It is Apple's codesign and
// security behind a port.
type Signer interface {
	// Ready reports *Unavailable when signing cannot happen here.
	Ready() error
	// OpenKeychain creates a temporary keychain holding a PKCS#12
	// certificate. The function returned deletes it again; until it is
	// called, the keychain is also deleted if the process is interrupted.
	OpenKeychain(p12 []byte, password string) (Keychain, func() error, error)
	// Identities lists the code signing identities in a keychain, or in the
	// user's keychains for the zero Keychain.
	Identities(k Keychain) ([]Identity, error)
	// Sign signs one piece of code, replacing an existing signature.
	Sign(path string, o SignOptions) error
	// Verify checks a signature, deeply and strictly.
	Verify(path string) error
	// Describe reads the signature of a piece of code.
	Describe(path string) (Signature, error)
}

// Notary is Apple's notary service, reached through xcrun notarytool, and the
// stapler.
type Notary interface {
	Ready() error
	// Submit uploads a file and waits for the verdict.
	Submit(path string, c Credentials) (Submission, error)
	// Log fetches the service's report on a submission, as JSON.
	Log(id string, c Credentials) ([]byte, error)
	// Staple attaches the ticket to a bundle or disk image.
	Staple(path string) error
}

// DiskImager creates disk images. It is Apple's hdiutil behind a port.
type DiskImager interface {
	Ready() error
	// Create makes a compressed image holding the bundle and a link to
	// /Applications.
	Create(spec DiskImageSpec) error
}

// Workspace is the file system as this context needs it.
type Workspace interface {
	Read(path string) ([]byte, error)
	Stat(path string) (FileInfo, error)
	// List returns everything below a directory, recursively and sorted, with
	// slash separated paths relative to it.
	List(dir string) ([]FileInfo, error)
	// Zip archives a directory, keeping links and permissions, with the
	// directory itself as the top level entry. The archive is written next
	// to its destination and moved into place when complete.
	Zip(dir, out string, o ZipOptions) error
	// TempDir returns a scratch directory, readable only by this user, and a
	// function removing it.
	TempDir() (string, func(), error)
	// WriteFile writes a file with the given permissions, next to its
	// destination first and then renamed into place.
	WriteFile(path string, data []byte, perm uint32, force bool) error
}
