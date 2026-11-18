package fs

import (
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/app/bundle"
	rbundle "github.com/worldiety/wdyapppack/requirements/fun/bundle"
)

func TestTreesAreMovedIntoPlaceWhole(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out", "My.app")
	f := New()
	when := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	first := []bundle.Node{
		{Path: "Contents", Dir: true, Mode: 0o755},
		{Path: "Contents/MacOS", Dir: true, Mode: 0o755},
		{Path: "Contents/MacOS/my", Mode: 0o755, Data: []byte("v1")},
		{Path: "Contents/Link", Link: "MacOS/my"},
	}
	if err := f.WriteTree(dest, first, bundle.WriteOptions{ModTime: when}); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(filepath.Join(dest, "Contents/MacOS/my"))
	if err != nil || st.Mode().Perm() != 0o755 || !st.ModTime().Equal(when) {
		t.Fatalf("%v %v", st, err)
	}
	if target, _ := os.Readlink(filepath.Join(dest, "Contents/Link")); target != "MacOS/my" {
		t.Fatalf("link %q", target)
	}

	// Not replaced without force.
	if err := f.WriteTree(dest, first, bundle.WriteOptions{}); !errors.Is(err, iofs.ErrExist) {
		t.Fatalf("got %v", err)
	}

	// A tree that fails half way leaves the old one in place and nothing else.
	broken := []bundle.Node{
		{Path: "Contents/MacOS/my", Mode: 0o755, Data: []byte("v2")},
		{Path: "Contents/Missing", From: filepath.Join(dir, "does-not-exist")},
	}
	if err := f.WriteTree(dest, broken, bundle.WriteOptions{Force: true}); err == nil {
		t.Fatal("a broken tree was written")
	}
	if data, _ := os.ReadFile(filepath.Join(dest, "Contents/MacOS/my")); string(data) != "v1" {
		t.Fatalf("the old tree was damaged: %q", data)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "out"))
	if len(entries) != 1 {
		t.Fatalf("leftovers next to the destination: %v", entries)
	}

	// With force, the new tree replaces the old one completely.
	second := []bundle.Node{{Path: "Contents/MacOS/my", Mode: 0o755, Data: []byte("v2")}}
	if err := f.WriteTree(dest, second, bundle.WriteOptions{Force: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(dest, "Contents/Link")); !os.IsNotExist(err) {
		t.Fatal("parts of the old tree survived")
	}
	entries, _ = os.ReadDir(filepath.Join(dir, "out"))
	if len(entries) != 1 {
		t.Fatalf("leftovers next to the destination: %v", entries)
	}
	spec.Verified(t, rbundle.RBundleSafeWrite)
}

func TestFilesAreReplacedOnlyWithForce(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a", "file")
	f := New()
	if err := f.WriteFile(p, []byte("1"), 0o644, bundle.WriteOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := f.WriteFile(p, []byte("2"), 0o644, bundle.WriteOptions{}); !errors.Is(err, iofs.ErrExist) {
		t.Fatalf("got %v", err)
	}
	if err := f.WriteFile(p, []byte("2"), 0o755, bundle.WriteOptions{Force: true}); err != nil {
		t.Fatal(err)
	}
	st, _ := os.Stat(p)
	if st.Mode().Perm() != 0o755 {
		t.Fatalf("mode %v", st.Mode())
	}
	if info, _ := f.Stat(filepath.Join(filepath.Dir(p), "nope")); info.Exists {
		t.Fatal("a missing file exists")
	}
	list, err := f.List(filepath.Dir(p))
	if err != nil || len(list) != 1 || list[0].Path != "file" {
		t.Fatalf("%v %v", list, err)
	}
}
