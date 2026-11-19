package fs

import (
	"archive/zip"
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/worldiety/wdyapppack/app/distribution"
)

func TestZipIsMovedIntoPlaceAndNotReplacedUnasked(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "My.app")
	if err := os.MkdirAll(filepath.Join(app, "Contents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(app, "Contents", "Info.plist"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out", "My.zip")
	w := New()
	if err := w.Zip(app+"/", out, distribution.ZipOptions{}); err != nil {
		t.Fatal(err)
	}
	r, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	if r.File[0].Name != "My.app/" {
		t.Fatalf("%s", r.File[0].Name)
	}
	r.Close()
	if err := w.Zip(app, out, distribution.ZipOptions{}); !errors.Is(err, iofs.ErrExist) {
		t.Fatalf("got %v", err)
	}
	if err := w.Zip(app, out, distribution.ZipOptions{Force: true}); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Dir(out))
	if len(entries) != 1 {
		t.Fatalf("leftovers: %v", entries)
	}
}

func TestPrivateFilesArePrivate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "AuthKey_X.p8")
	w := New()
	if err := w.WriteFile(p, []byte("key"), 0o600, false); err != nil {
		t.Fatal(err)
	}
	st, _ := os.Stat(p)
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", st.Mode())
	}
	if err := w.WriteFile(p, []byte("key"), 0o600, false); !errors.Is(err, iofs.ErrExist) {
		t.Fatalf("got %v", err)
	}
	dir, cleanup, err := w.TempDir()
	if err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Stat(dir); st.Mode().Perm() != 0o700 {
		t.Fatalf("scratch directory mode %v", st.Mode())
	}
	cleanup()
}
