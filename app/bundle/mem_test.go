package bundle

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/wdyapppack/foundation/auth"
	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/macho/machotest"
	"github.com/worldiety/wdyapppack/pkg/plist"
)

// memFiles is the Files port in memory. The whole point of declaring the port
// in the context is that a test can put this behind it.
type memFiles struct {
	files  map[string]memFile
	writes int
	// corrupt makes every written Info.plist unreadable, to test the self-check.
	corrupt bool
}

type memFile struct {
	data []byte
	mode fs.FileMode
	dir  bool
	link string
}

func newMem() *memFiles { return &memFiles{files: map[string]memFile{}} }

func (m *memFiles) put(path string, data []byte, mode fs.FileMode) {
	m.files[filepath.Clean(path)] = memFile{data: data, mode: mode}
}

func (m *memFiles) Read(path string) ([]byte, error) {
	f, ok := m.files[filepath.Clean(path)]
	if !ok || f.dir {
		return nil, fs.ErrNotExist
	}
	return f.data, nil
}

func (m *memFiles) Stat(path string) (FileInfo, error) {
	p := filepath.Clean(path)
	if f, ok := m.files[p]; ok {
		return FileInfo{Path: path, Exists: true, Dir: f.dir, Mode: f.mode, Size: int64(len(f.data))}, nil
	}
	for k := range m.files {
		if strings.HasPrefix(k, p+"/") {
			return FileInfo{Path: path, Exists: true, Dir: true, Mode: 0o755}, nil
		}
	}
	return FileInfo{Path: path}, nil
}

func (m *memFiles) List(dir string) ([]FileInfo, error) {
	p := filepath.Clean(dir) + "/"
	var out []FileInfo
	for k, f := range m.files {
		if rel, ok := strings.CutPrefix(k, p); ok {
			out = append(out, FileInfo{Path: rel, Exists: true, Dir: f.dir, Mode: f.mode, Link: f.link})
		}
	}
	slices.SortFunc(out, func(a, b FileInfo) int { return strings.Compare(a.Path, b.Path) })
	return out, nil
}

func (m *memFiles) WriteFile(path string, data []byte, mode uint32, o WriteOptions) error {
	if st, _ := m.Stat(path); st.Exists && !o.Force {
		return fmt.Errorf("%s: %w", path, fs.ErrExist)
	}
	m.writes++
	m.put(path, data, fs.FileMode(mode))
	return nil
}

func (m *memFiles) WriteTree(dir string, nodes []Node, o WriteOptions) error {
	if st, _ := m.Stat(dir); st.Exists && !o.Force {
		return fmt.Errorf("%s: %w", dir, fs.ErrExist)
	}
	m.writes++
	for _, n := range nodes {
		p := filepath.Join(dir, n.Path)
		switch {
		case n.Dir:
			m.files[p] = memFile{dir: true, mode: n.Mode}
		case n.Link != "":
			m.files[p] = memFile{link: n.Link, mode: fs.ModeSymlink | 0o777}
		case n.From != "":
			data, err := m.Read(n.From)
			if err != nil {
				return err
			}
			m.put(p, data, n.Mode)
		default:
			data := n.Data
			if m.corrupt && strings.HasSuffix(p, "Info.plist") {
				data = []byte("garbage")
			}
			m.put(p, data, n.Mode)
		}
	}
	return nil
}

// fixtures

func pngImage(t *testing.T, w, h int, corners bool) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			inside := x > w/8 && x < w*7/8 && y > h/8 && y < h*7/8
			if inside || corners {
				img.SetNRGBA(x, y, color.NRGBA{30, 120, 200, 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func jpegImage(t *testing.T, n int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func arm64(opts ...func(*machotest.Options)) []byte {
	o := machotest.Options{Arch: "arm64", MinOS: "12.0", Signed: true, Libraries: []string{"/usr/lib/libSystem.B.dylib"}}
	for _, f := range opts {
		f(&o)
	}
	return machotest.Program(o)
}

func x86(opts ...func(*machotest.Options)) []byte {
	o := machotest.Options{Arch: "x86_64", MinOS: "10.15", Libraries: []string{"/usr/lib/libSystem.B.dylib"}}
	for _, f := range opts {
		f(&o)
	}
	return machotest.Program(o)
}

// workspace is a typical project: a program, an icon, some resources.
func workspace(t *testing.T) *memFiles {
	m := newMem()
	m.put("dist/mytool-darwin-arm64", arm64(), 0o755)
	m.put("dist/mytool-darwin-amd64", x86(), 0o755)
	m.put("assets/icon.png", pngImage(t, 1024, 1024, false), 0o644)
	m.put("assets/data/a.txt", []byte("a"), 0o644)
	m.put("assets/data/sub/b.txt", []byte("b"), 0o644)
	return m
}

func validConfig() Config {
	return Config{
		Name:       "My Tool",
		Identifier: "com.example.mytool",
		Version:    "1.2.0",
		Binaries:   []Binary{{Path: "dist/mytool-darwin-arm64"}},
		Icon:       &Icon{Path: "assets/icon.png"},
	}
}

func buildWith(m *memFiles, cfg Config, mod ...func(*BuildRequest)) (BuildResult, error) {
	req := BuildRequest{Config: cfg}
	for _, f := range mod {
		f(&req)
	}
	return NewBuildBundle(m)(auth.System(), req)
}

func infoOf(t *testing.T, m *memFiles, app string) map[string]any {
	t.Helper()
	data, err := m.Read(filepath.Join(app, "Contents/Info.plist"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := plist.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	return v.(map[string]any)
}

// codes returns the message codes of an error.
func codes(err error) []string {
	var e *problem.E
	if !errors.As(err, &e) {
		return nil
	}
	var out []string
	for _, m := range e.Messages {
		out = append(out, m.Code)
	}
	return out
}

func noteCodes(r Report) []string {
	var out []string
	for _, n := range r.Notes {
		out = append(out, n.Code)
	}
	return out
}

func wantCodes(t *testing.T, err error, want ...string) {
	t.Helper()
	got := codes(err)
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Errorf("missing %s in %v (%v)", w, got, err)
		}
	}
}
