package bundle

// Files is how this context reaches the file system.
//
// It is declared here, where it is used. Which implementation is behind it is
// decided under cmd, so that a test can put one in memory there.
type Files interface {
	// Read returns the content of a file.
	Read(path string) ([]byte, error)
	// Stat describes a path. A path that does not exist is not an error:
	// Exists is false.
	Stat(path string) (FileInfo, error)
	// List returns everything below a directory, recursively and sorted, with
	// paths relative to it.
	List(dir string) ([]FileInfo, error)
	// WriteFile writes one file. It replaces an existing one only with Force,
	// and never leaves a half written file behind.
	WriteFile(path string, data []byte, mode uint32, o WriteOptions) error
	// WriteTree writes a directory tree. The tree is assembled next to its
	// destination and moved into place when complete; an existing destination
	// is replaced only with Force.
	WriteTree(dir string, nodes []Node, o WriteOptions) error
}
