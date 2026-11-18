package ziptree

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/fun/dist"
)

func tree(t *testing.T) string {
	dir := t.TempDir()
	app := filepath.Join(dir, "My.app", "Contents")
	must(t, os.MkdirAll(filepath.Join(app, "MacOS"), 0o755))
	must(t, os.MkdirAll(filepath.Join(app, "Frameworks", "F.framework", "Versions", "A"), 0o755))
	must(t, os.WriteFile(filepath.Join(app, "MacOS", "my"), []byte("program"), 0o755))
	must(t, os.WriteFile(filepath.Join(app, "Info.plist"), []byte("<plist/>"), 0o644))
	must(t, os.Symlink("A", filepath.Join(app, "Frameworks", "F.framework", "Versions", "Current")))
	return dir
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLinksAndPermissionsSurvive(t *testing.T) {
	dir := tree(t)
	when := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var buf bytes.Buffer
	must(t, Write(&buf, os.DirFS(dir), "My.app", when))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	must(t, err)
	got := map[string]*zip.File{}
	for _, f := range zr.File {
		got[f.Name] = f
	}
	if _, ok := got["My.app/"]; !ok {
		t.Fatal("the bundle is not the top level entry")
	}
	exe := got["My.app/Contents/MacOS/my"]
	if exe == nil || exe.Mode().Perm()&0o111 == 0 {
		t.Fatal("the program lost its executable bit")
	}
	link := got["My.app/Contents/Frameworks/F.framework/Versions/Current"]
	if link == nil || link.Mode()&fs.ModeSymlink == 0 {
		t.Fatal("the link was not kept as a link")
	}
	rc, _ := link.Open()
	target, _ := io.ReadAll(rc)
	if string(target) != "A" {
		t.Fatalf("link points to %q", target)
	}
	if !exe.Modified.Equal(when) {
		t.Fatalf("time %v", exe.Modified)
	}

	var again bytes.Buffer
	must(t, Write(&again, os.DirFS(dir), "My.app", when))
	if !bytes.Equal(buf.Bytes(), again.Bytes()) {
		t.Fatal("two archives of the same tree differ")
	}
	spec.Verified(t, dist.RDistZip)
}
