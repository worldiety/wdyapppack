package distribution

import (
	"errors"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/plist"
)

type memWS struct {
	files map[string]fileEntry
	zips  []string
}

type fileEntry struct {
	data []byte
	dir  bool
}

func newWS() *memWS { return &memWS{files: map[string]fileEntry{}} }

func (m *memWS) put(p string, data []byte) {
	p = filepath.Clean(p)
	m.files[p] = fileEntry{data: data}
	for d := filepath.Dir(p); d != "." && d != "/"; d = filepath.Dir(d) {
		m.files[d] = fileEntry{dir: true}
	}
}

func (m *memWS) Read(p string) ([]byte, error) {
	f, ok := m.files[filepath.Clean(p)]
	if !ok || f.dir {
		return nil, fs.ErrNotExist
	}
	return f.data, nil
}

func (m *memWS) Stat(p string) (FileInfo, error) {
	f, ok := m.files[filepath.Clean(p)]
	return FileInfo{Path: p, Exists: ok, Dir: f.dir}, nil
}

func (m *memWS) List(dir string) ([]FileInfo, error) {
	prefix := filepath.Clean(dir) + "/"
	var out []FileInfo
	for k, f := range m.files {
		if rel, ok := strings.CutPrefix(k, prefix); ok {
			out = append(out, FileInfo{Path: rel, Exists: true, Dir: f.dir})
		}
	}
	slices.SortFunc(out, func(a, b FileInfo) int { return strings.Compare(a.Path, b.Path) })
	return out, nil
}

func (m *memWS) Zip(dir, out string, o ZipOptions) error {
	if st, _ := m.Stat(out); st.Exists && !o.Force {
		return fs.ErrExist
	}
	m.zips = append(m.zips, dir+" -> "+out)
	m.put(out, []byte("zip"))
	return nil
}

func (m *memWS) TempDir() (string, func(), error) { return "tmp", func() {}, nil }

// fakeSigner records what it signs.
type fakeSigner struct {
	unavailable error
	ids         []Identity
	signed      []string
	opts        []SignOptions
	verifyErr   error
	sig         Signature
}

func (f *fakeSigner) Ready() error                    { return f.unavailable }
func (f *fakeSigner) Identities() ([]Identity, error) { return f.ids, nil }
func (f *fakeSigner) Sign(p string, o SignOptions) error {
	f.signed = append(f.signed, p)
	f.opts = append(f.opts, o)
	return nil
}
func (f *fakeSigner) Verify(string) error                { return f.verifyErr }
func (f *fakeSigner) Describe(string) (Signature, error) { return f.sig, nil }

type fakeNotary struct {
	unavailable error
	verdict     Submission
	log         []byte
	submitted   []string
	creds       []Credentials
	stapled     []string
}

func (f *fakeNotary) Ready() error { return f.unavailable }
func (f *fakeNotary) Submit(p string, c Credentials) (Submission, error) {
	f.submitted = append(f.submitted, p)
	f.creds = append(f.creds, c)
	return f.verdict, nil
}
func (f *fakeNotary) Log(string, Credentials) ([]byte, error) { return f.log, nil }
func (f *fakeNotary) Staple(p string) error {
	f.stapled = append(f.stapled, p)
	return nil
}

type fakeImager struct {
	unavailable error
	specs       []DiskImageSpec
}

func (f *fakeImager) Ready() error { return f.unavailable }
func (f *fakeImager) Create(s DiskImageSpec) error {
	f.specs = append(f.specs, s)
	return nil
}

const app = "dist/My Tool.app"

func bundleWS(t *testing.T) *memWS {
	t.Helper()
	m := newWS()
	info, err := plist.Encode(map[string]any{
		"CFBundleName": "My Tool", "CFBundleShortVersionString": "1.2.0",
		"CFBundleIdentifier": "com.example.mytool", "CFBundleExecutable": "mytool",
	})
	if err != nil {
		t.Fatal(err)
	}
	m.put(app+"/Contents/Info.plist", info)
	m.put(app+"/Contents/MacOS/mytool", []byte("prog"))
	m.put(app+"/Contents/MacOS/tool2", []byte("prog"))
	m.put(app+"/Contents/Helpers/helper", []byte("prog"))
	m.put(app+"/Contents/Frameworks/libfoo.dylib", []byte("lib"))
	m.put(app+"/Contents/Frameworks/Bar.framework/Versions/A/Bar", []byte("lib"))
	m.put(app+"/Contents/Resources/AppIcon.icns", []byte("icns"))
	return m
}

var devID = Identity{Hash: "ABCDEF0123456789ABCDEF0123456789ABCDEF01", Name: "Developer ID Application: Example GmbH (TEAM123456)"}

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

func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	if !slices.Contains(codes(err), code) {
		t.Fatalf("want %s, got %v (%v)", code, codes(err), err)
	}
}

func exitKind(err error) problem.Kind {
	var e *problem.E
	if errors.As(err, &e) {
		return e.Kind
	}
	return -1
}
