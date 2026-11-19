package distribution

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
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
	// Certificate, when set, is imported into a temporary keychain to sign
	// the image from.
	Certificate *Certificate
	Force       bool
	DryRun      bool
}

// DiskImage describes the image.
type DiskImage struct {
	Path   string `json:"path"`
	Volume string `json:"volume"`
	Signed bool   `json:"signed"`
	// Checksum is the file holding the SHA-256 of the image.
	Checksum string `json:"checksum"`
	SHA256   string `json:"sha256,omitempty"`
	DryRun   bool   `json:"dryRun"`
	Report
}

func createDiskImage(ws Workspace, imager DiskImager, signer Signer, req DiskImageRequest) (res DiskImage, err error) {
	b, err := readBundle(ws, req.App)
	if err != nil {
		return DiskImage{}, err
	}
	out := cmp.Or(req.Out, artifactName(b, ".dmg"))
	if err := refuseExisting(ws, req.Force, out, out+".sha256"); err != nil {
		return DiskImage{}, err
	}
	ports := []interface{ Ready() error }{imager}
	if req.Identity != "" {
		ports = append(ports, signer)
	}
	if err := ready(ports...); err != nil {
		return DiskImage{}, err
	}

	var notes problem.Collector
	res = DiskImage{Path: out, Volume: cmp.Or(req.Volume, b.name), Signed: req.Identity != "", Checksum: out + ".sha256", DryRun: req.DryRun}
	o := SignOptions{Identity: "-"}
	if req.Identity != "" && req.Identity != "-" {
		kc, closeKC, err := openKeychain(ws, signer, req.Certificate, &notes)
		if err != nil {
			return DiskImage{}, err
		}
		defer func() { closeKC(); res.Notes = notes.Notes() }()
		id, err := resolveIdentity(signer, kc, req.Identity, req.Certificate)
		if err != nil {
			return DiskImage{}, err
		}
		o = SignOptions{Identity: id.Hash, Timestamp: true, Keychain: kc.Path}
	}
	if req.DryRun {
		return res, nil
	}
	if err := imager.Create(DiskImageSpec{App: req.App, Out: out, Volume: res.Volume, Force: req.Force}); err != nil {
		return DiskImage{}, problem.Fail(msg("WAP-DMG-FAILED", fmt.Sprintf("%s could not be created: %v", out, strings.TrimSpace(err.Error()))))
	}
	if req.Identity != "" {
		if err := signer.Sign(out, o); err != nil {
			return DiskImage{}, problem.Fail(msg("WAP-SIGN-FAILED", fmt.Sprintf("Signing %s failed: %v", out, strings.TrimSpace(err.Error()))))
		}
	}
	// The checksum is taken last, of the image as it is handed out.
	if res.SHA256, err = writeChecksum(ws, out, req.Force); err != nil {
		return DiskImage{}, err
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
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
	SHA256   string `json:"sha256,omitempty"`
	DryRun   bool   `json:"dryRun"`
	Report
}

func createArchive(ws Workspace, req ArchiveRequest) (Archive, error) {
	b, err := readBundle(ws, req.App)
	if err != nil {
		return Archive{}, err
	}
	out := cmp.Or(req.Out, artifactName(b, ".zip"))
	if err := refuseExisting(ws, req.Force, out, out+".sha256"); err != nil {
		return Archive{}, err
	}
	res := Archive{Path: out, Checksum: out + ".sha256", DryRun: req.DryRun}
	if req.DryRun {
		return res, nil
	}
	if err := ws.Zip(req.App, out, ZipOptions{Force: req.Force, ModTime: req.ModTime}); err != nil {
		return Archive{}, problem.Fail(msg("WAP-ZIP-FAILED", fmt.Sprintf("%s could not be written: %v.", out, err)))
	}
	if res.SHA256, err = writeChecksum(ws, out, req.Force); err != nil {
		return Archive{}, err
	}
	return res, nil
}

// writeChecksum writes <file>.sha256 in the format of shasum -a 256, so that
// "shasum -a 256 -c <file>.sha256" checks the download.
func writeChecksum(ws Workspace, file string, force bool) (string, error) {
	data, err := ws.Read(file)
	if err != nil {
		return "", problem.Fail(msg("WAP-CHECKSUM-FAILED", fmt.Sprintf("%s cannot be read back: %v.", file, err)))
	}
	sum := sha256.Sum256(data)
	hexSum := hex.EncodeToString(sum[:])
	line := hexSum + "  " + filepath.Base(file) + "\n"
	if err := ws.WriteFile(file+".sha256", []byte(line), 0o644, force); err != nil {
		return "", problem.Fail(msg("WAP-CHECKSUM-FAILED", fmt.Sprintf("%s.sha256 could not be written: %v.", file, err)))
	}
	return hexSum, nil
}

func artifactName(b bundleInfo, ext string) string {
	name := strings.ReplaceAll(b.name, " ", "")
	if b.version != "" {
		name += "-" + b.version
	}
	return filepath.Join(filepath.Dir(strings.TrimRight(b.path, "/")), name+ext)
}

func refuseExisting(ws Workspace, force bool, paths ...string) error {
	if force {
		return nil
	}
	for _, out := range paths {
		if st, err := ws.Stat(out); err == nil && st.Exists {
			return problem.Fail(msg("WAP-ARTIFACT-EXISTS", fmt.Sprintf("%s exists already.", out)))
		}
	}
	return nil
}
