// Package ziptree writes a directory tree into a zip archive the way macOS's
// ditto does: symbolic links stay links, executable files stay executable, and
// the directory itself is the top level entry.
//
// A zip written naively follows links and drops permissions. For an
// application bundle both break it: a framework is a set of links, and a
// program without its executable bit does not start.
package ziptree

import (
	"archive/zip"
	"io"
	"io/fs"
	"path"
	"time"
)

// Write archives the tree at root inside fsys. Entries are named relative to
// fsys, so with fsys = os.DirFS(parent) and root = "MyTool.app" the archive
// unpacks to MyTool.app/…. A non-zero mtime is written as the time of every
// entry, which makes the archive reproducible.
func Write(w io.Writer, fsys fs.FS, root string, mtime time.Time) error {
	zw := zip.NewWriter(w)
	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		h := &zip.FileHeader{Name: p}
		h.SetMode(info.Mode())
		h.Modified = info.ModTime()
		if !mtime.IsZero() {
			h.Modified = mtime
		}
		h.Modified = h.Modified.UTC()

		switch {
		case d.IsDir():
			h.Name = path.Clean(p) + "/"
			h.Method = zip.Store
			_, err := zw.CreateHeader(h)
			return err
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := fs.ReadLink(fsys, p)
			if err != nil {
				return err
			}
			h.Method = zip.Store
			fw, err := zw.CreateHeader(h)
			if err != nil {
				return err
			}
			_, err = io.WriteString(fw, target)
			return err
		case info.Mode().IsRegular():
			h.Method = zip.Deflate
			fw, err := zw.CreateHeader(h)
			if err != nil {
				return err
			}
			f, err := fsys.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(fw, f)
			return err
		}
		// Sockets, devices and pipes have no place in a bundle.
		return nil
	})
	if err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}
