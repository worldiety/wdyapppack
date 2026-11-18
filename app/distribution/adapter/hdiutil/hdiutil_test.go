package hdiutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/worldiety/wdyapppack/app/distribution"
	"github.com/worldiety/wdyapppack/foundation/run"
	"github.com/worldiety/wdyapppack/foundation/run/runtest"
)

func app(t *testing.T) string {
	dir := t.TempDir()
	a := filepath.Join(dir, "My.app")
	if err := os.MkdirAll(filepath.Join(a, "Contents", "MacOS"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "Contents", "MacOS", "my"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("MacOS/my", filepath.Join(a, "Contents", "link")); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestHdiutilIsAskedForACompressedImage(t *testing.T) {
	h := runtest.Fake(t, t.TempDir(), "hdiutil")
	i := New(&run.Runner{}, Tools{Hdiutil: h.Path, OS: "darwin"})
	out := filepath.Join(t.TempDir(), "My.dmg")
	if err := i.Create(distribution.DiskImageSpec{App: app(t), Out: out, Volume: "My"}); err != nil {
		t.Fatal(err)
	}
	c := h.Calls(t)[0]
	if c[0] != "create" || !slices.Contains(c, "UDZO") || c[len(c)-1] != out || c[slices.Index(c, "-volname")+1] != "My" {
		t.Fatalf("%q", c)
	}
}

// On a Mac, the real hdiutil makes an image that holds the bundle and a link
// to the Applications folder.
func TestRealDiskImage(t *testing.T) {
	if runtime.GOOS != "darwin" || testing.Short() {
		t.Skip("hdiutil exists only on macOS")
	}
	if _, err := exec.LookPath("hdiutil"); err != nil {
		t.Skip("hdiutil not installed")
	}
	i := New(&run.Runner{}, DefaultTools())
	out := filepath.Join(t.TempDir(), "My.dmg")
	if err := i.Create(distribution.DiskImageSpec{App: app(t), Out: out, Volume: "My Test"}); err != nil {
		t.Fatal(err)
	}
	mnt := t.TempDir()
	if b, err := exec.Command("hdiutil", "attach", "-nobrowse", "-readonly", "-mountpoint", mnt, out).CombinedOutput(); err != nil {
		t.Fatalf("%s", b)
	}
	defer exec.Command("hdiutil", "detach", mnt).Run()
	if target, err := os.Readlink(filepath.Join(mnt, "Applications")); err != nil || target != "/Applications" {
		t.Fatalf("%q %v", target, err)
	}
	if st, err := os.Stat(filepath.Join(mnt, "My.app", "Contents", "MacOS", "my")); err != nil || st.Mode()&0o111 == 0 {
		t.Fatalf("%v %v", st, err)
	}
}
