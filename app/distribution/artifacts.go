package distribution

import (
	"cmp"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

// DiskImageRequest asks for a disk image.
type DiskImageRequest struct {
	App string
	// Out defaults to <name>-<version>.dmg next to the bundle.
	Out    string
	Volume string
	// Identity, when set, signs the image as well.
	Identity string
	Force    bool
	DryRun   bool
}

// DiskImage describes the image.
type DiskImage struct {
	Path   string `json:"path"`
	Volume string `json:"volume"`
	Signed bool   `json:"signed"`
	DryRun bool   `json:"dryRun"`
	Report
}

func createDiskImage(ws Workspace, imager DiskImager, signer Signer, req DiskImageRequest) (DiskImage, error) {
	b, err := readBundle(ws, req.App)
	if err != nil {
		return DiskImage{}, err
	}
	out := cmp.Or(req.Out, artifactName(b, ".dmg"))
	if err := refuseExisting(ws, out, req.Force); err != nil {
		return DiskImage{}, err
	}
	ports := []interface{ Ready() error }{imager}
	if req.Identity != "" {
		ports = append(ports, signer)
	}
	if err := ready(ports...); err != nil {
		return DiskImage{}, err
	}
	var id Identity
	if req.Identity != "" && req.Identity != "-" {
		if id, err = resolveIdentity(signer, req.Identity); err != nil {
			return DiskImage{}, err
		}
	}
	res := DiskImage{Path: out, Volume: cmp.Or(req.Volume, b.name), Signed: req.Identity != "", DryRun: req.DryRun}
	if req.DryRun {
		return res, nil
	}
	if err := imager.Create(DiskImageSpec{App: req.App, Out: out, Volume: res.Volume, Force: req.Force}); err != nil {
		return DiskImage{}, problem.Fail(msg("WAP-DMG-FAILED", fmt.Sprintf("%s could not be created: %v", out, strings.TrimSpace(err.Error()))))
	}
	if req.Identity != "" {
		o := SignOptions{Identity: "-"}
		if id.Hash != "" {
			o = SignOptions{Identity: id.Hash, Timestamp: true}
		}
		if err := signer.Sign(out, o); err != nil {
			return DiskImage{}, problem.Fail(msg("WAP-SIGN-FAILED", fmt.Sprintf("Signing %s failed: %v", out, strings.TrimSpace(err.Error()))))
		}
	}
	return res, nil
}

// ArchiveRequest asks for a zip archive.
type ArchiveRequest struct {
	App string
	// Out defaults to <name>-<version>.zip next to the bundle.
	Out     string
	ModTime time.Time
	Force   bool
	DryRun  bool
}

// Archive describes the archive.
type Archive struct {
	Path   string `json:"path"`
	DryRun bool   `json:"dryRun"`
	Report
}

func createArchive(ws Workspace, req ArchiveRequest) (Archive, error) {
	b, err := readBundle(ws, req.App)
	if err != nil {
		return Archive{}, err
	}
	out := cmp.Or(req.Out, artifactName(b, ".zip"))
	if err := refuseExisting(ws, out, req.Force); err != nil {
		return Archive{}, err
	}
	res := Archive{Path: out, DryRun: req.DryRun}
	if req.DryRun {
		return res, nil
	}
	if err := ws.Zip(req.App, out, ZipOptions{Force: req.Force, ModTime: req.ModTime}); err != nil {
		return Archive{}, problem.Fail(msg("WAP-ZIP-FAILED", fmt.Sprintf("%s could not be written: %v.", out, err)))
	}
	return res, nil
}

func artifactName(b bundleInfo, ext string) string {
	name := strings.ReplaceAll(b.name, " ", "")
	if b.version != "" {
		name += "-" + b.version
	}
	return filepath.Join(filepath.Dir(strings.TrimRight(b.path, "/")), name+ext)
}

func refuseExisting(ws Workspace, out string, force bool) error {
	if force {
		return nil
	}
	if st, err := ws.Stat(out); err == nil && st.Exists {
		return problem.Fail(msg("WAP-ARTIFACT-EXISTS", fmt.Sprintf("%s exists already.", out)))
	}
	return nil
}
