// Package fs implements the bundle context's Files port over the local file
// system.
//
// An adapter implements a port and is named after the technology behind it.
// Nothing outside cmd imports this: the context declares what it needs, this
// provides it, and the entry point is where the two meet.
package fs

import (
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/worldiety/wdyapppack/app/bundle"
)

// Files is the file system.
type Files struct{}

// New returns the file system adapter.
func New() Files { return Files{} }

// Read returns the content of a file.
func (Files) Read(path string) ([]byte, error) { return os.ReadFile(path) }

// Stat describes a path, following links.
func (Files) Stat(path string) (bundle.FileInfo, error) {
	st, err := os.Stat(path)
	if os.IsNotExist(err) {
		return bundle.FileInfo{Path: path}, nil
	}
	if err != nil {
		return bundle.FileInfo{Path: path}, err
	}
	return bundle.FileInfo{Path: path, Exists: true, Dir: st.IsDir(), Mode: st.Mode(), Size: st.Size()}, nil
}

// List returns everything below a directory, without following links.
func (Files) List(dir string) ([]bundle.FileInfo, error) {
	return list(dir)
}

func list(dir string) ([]bundle.FileInfo, error) {
	var out []bundle.FileInfo
	err := filepath.WalkDir(dir, func(p string, d iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dir {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		fi := bundle.FileInfo{Path: filepath.ToSlash(rel), Exists: true, Dir: d.IsDir(), Mode: info.Mode(), Size: info.Size()}
		if info.Mode()&iofs.ModeSymlink != 0 {
			if fi.Link, err = os.Readlink(p); err != nil {
				return err
			}
		}
		out = append(out, fi)
		return nil
	})
	return out, err
}

// WriteFile writes a file next to its destination and renames it into place,
// so that a reader never sees half of it.
func (Files) WriteFile(path string, data []byte, mode uint32, o bundle.WriteOptions) error {
	if err := refuseExisting(path, o.Force); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".wdyapppack-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), iofs.FileMode(mode)); err != nil {
		return err
	}
	if !o.ModTime.IsZero() {
		if err := os.Chtimes(tmp.Name(), o.ModTime, o.ModTime); err != nil {
			return err
		}
	}
	if o.Force {
		if st, err := os.Lstat(path); err == nil && st.IsDir() {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		}
	}
	return os.Rename(tmp.Name(), path)
}

// WriteTree assembles a tree in a hidden directory next to its destination and
// moves it into place only when it is complete. An existing destination is
// moved aside first and removed afterwards, so that an interruption at any
// point leaves either the old tree or the new one.
func (Files) WriteTree(dir string, nodes []bundle.Node, o bundle.WriteOptions) (err error) {
	if err := refuseExisting(dir, o.Force); err != nil {
		return err
	}
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(parent, "."+filepath.Base(dir)+".wdyapppack-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(tmp)
		}
	}()
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}

	var dirs []string
	for _, n := range nodes {
		target := filepath.Join(tmp, filepath.FromSlash(n.Path))
		if !strings.HasPrefix(target, tmp+string(filepath.Separator)) {
			return fmt.Errorf("%s leaves the tree", n.Path)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		switch {
		case n.Dir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			if err := os.Chmod(target, mode(n.Mode, 0o755)); err != nil {
				return err
			}
			dirs = append(dirs, target)
			continue
		case n.Link != "":
			if err := os.Symlink(n.Link, target); err != nil {
				return err
			}
			continue
		case n.From != "":
			if err := copyFile(n.From, target); err != nil {
				return err
			}
		default:
			if err := os.WriteFile(target, n.Data, 0o600); err != nil {
				return err
			}
		}
		if err := os.Chmod(target, mode(n.Mode, 0o644)); err != nil {
			return err
		}
		if !o.ModTime.IsZero() {
			if err := os.Chtimes(target, o.ModTime, o.ModTime); err != nil {
				return err
			}
		}
	}
	if !o.ModTime.IsZero() {
		// Deepest first: setting a child's time does not touch its parent's,
		// but creating it did.
		slices.SortFunc(dirs, func(a, b string) int { return len(b) - len(a) })
		for _, d := range append(dirs, tmp) {
			if err := os.Chtimes(d, o.ModTime, o.ModTime); err != nil {
				return err
			}
		}
	}

	if _, statErr := os.Lstat(dir); statErr == nil {
		old := tmp + ".old"
		if err := os.Rename(dir, old); err != nil {
			return err
		}
		if err := os.Rename(tmp, dir); err != nil {
			_ = os.Rename(old, dir)
			return err
		}
		return os.RemoveAll(old)
	}
	return os.Rename(tmp, dir)
}

func mode(m iofs.FileMode, def iofs.FileMode) iofs.FileMode {
	if m.Perm() == 0 {
		return def
	}
	return m.Perm()
}

func refuseExisting(path string, force bool) error {
	if force {
		return nil
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%s: %w", path, iofs.ErrExist)
	}
	return nil
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

var _ bundle.Files = Files{}
