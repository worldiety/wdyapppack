// Package fs implements the distribution context's Workspace port over the
// local file system.
package fs

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/pkg/ziptree"
)

// Workspace is the file system.
type Workspace struct{}

// New returns the workspace.
func New() Workspace { return Workspace{} }

// Read returns the content of a file.
func (Workspace) Read(path string) ([]byte, error) { return os.ReadFile(path) }

// Stat describes a path, following links.
func (Workspace) Stat(path string) (distribution.FileInfo, error) {
	st, err := os.Stat(path)
	if os.IsNotExist(err) {
		return distribution.FileInfo{Path: path}, nil
	}
	if err != nil {
		return distribution.FileInfo{Path: path}, err
	}
	return distribution.FileInfo{Path: path, Exists: true, Dir: st.IsDir(), Mode: st.Mode()}, nil
}

// List returns everything below a directory, without following links.
func (Workspace) List(dir string) ([]distribution.FileInfo, error) {
	var out []distribution.FileInfo
	err := filepath.WalkDir(dir, func(p string, d iofs.DirEntry, err error) error {
		if err != nil || p == dir {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		out = append(out, distribution.FileInfo{Path: filepath.ToSlash(rel), Exists: true, Dir: d.IsDir(), Mode: info.Mode()})
		return nil
	})
	return out, err
}

// Zip archives a directory next to its destination and renames it into place.
func (Workspace) Zip(dir, out string, o distribution.ZipOptions) error {
	if !o.Force {
		if _, err := os.Lstat(out); err == nil {
			return fmt.Errorf("%s: %w", out, iofs.ErrExist)
		}
	}
	dir = strings.TrimRight(dir, string(filepath.Separator))
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(out), "."+filepath.Base(out)+".wdyapppack-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := ziptree.Write(tmp, os.DirFS(filepath.Dir(dir)), filepath.Base(dir), o.ModTime); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), out)
}

// TempDir returns a scratch directory.
func (Workspace) TempDir() (string, func(), error) {
	d, err := os.MkdirTemp("", "wdyapppack-*")
	if err != nil {
		return "", func() {}, err
	}
	return d, func() { os.RemoveAll(d) }, nil
}

var _ distribution.Workspace = Workspace{}
