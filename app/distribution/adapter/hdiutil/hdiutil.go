// Package hdiutil implements the distribution context's DiskImager port with
// Apple's hdiutil.
package hdiutil

import (
	"io"
	iofs "io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/run"
)

// Tools names the program; tests put a fake here.
type Tools struct {
	Hdiutil string
	OS      string
}

// DefaultTools are Apple's tools on this machine.
func DefaultTools() Tools { return Tools{Hdiutil: "hdiutil", OS: runtime.GOOS} }

// Imager is hdiutil.
type Imager struct {
	tools Tools
	run   *run.Runner
}

// New returns the imager.
func New(r *run.Runner, t Tools) Imager { return Imager{tools: t, run: r} }

// Ready reports whether hdiutil can run here.
func (i Imager) Ready() error {
	if i.tools.OS != "darwin" {
		return &distribution.Unavailable{Tool: "hdiutil", OS: i.tools.OS}
	}
	if _, err := exec.LookPath(i.tools.Hdiutil); err != nil {
		return &distribution.Unavailable{Tool: "hdiutil"}
	}
	return nil
}

// Create stages the bundle with a link to /Applications and makes a read only
// image of it, compressed as far as UDZO goes.
func (i Imager) Create(spec distribution.DiskImageSpec) error {
	stage, err := os.MkdirTemp("", "wdyapppack-dmg-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)

	app := strings.TrimRight(spec.App, string(filepath.Separator))
	if err := copyTree(app, filepath.Join(stage, filepath.Base(app))); err != nil {
		return err
	}
	if err := os.Symlink("/Applications", filepath.Join(stage, "Applications")); err != nil {
		return err
	}
	if spec.Force {
		if err := os.Remove(spec.Out); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(spec.Out), 0o755); err != nil {
		return err
	}
	_, err = i.run.Run(nil, i.tools.Hdiutil, "create",
		"-volname", spec.Volume, "-srcfolder", stage, "-fs", "HFS+", "-format", "UDZO", "-imagekey", "zlib-level=9", spec.Out)
	return err
}

// copyTree copies a directory, keeping links as links and permissions as
// they are.
func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(p string, d iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, p)
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			return os.MkdirAll(target, info.Mode().Perm()|0o700)
		case info.Mode()&iofs.ModeSymlink != 0:
			link, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_EXCL, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}

var _ distribution.DiskImager = Imager{}
