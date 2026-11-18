package plist

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func sample() map[string]any {
	return map[string]any{
		"CFBundleName":            "My <Tool> & Co",
		"NSHighResolutionCapable": true,
		"LSUIElement":             false,
		"Count":                   int64(42),
		"Ratio":                   1.5,
		"When":                    time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		"Blob":                    []byte{0, 1, 2, 250},
		"Schemes":                 []any{"a", "b"},
		"Empty":                   map[string]any{},
		"Nested":                  map[string]any{"Inner": []any{map[string]any{"K": "v"}}},
	}
}

func TestRoundTrip(t *testing.T) {
	data, err := Encode(sample())
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, sample()) {
		t.Fatalf("round trip changed the values:\n%#v\n%s", got, data)
	}
	if !strings.Contains(string(data), "My &lt;Tool&gt; &amp; Co") {
		t.Fatalf("text not escaped:\n%s", data)
	}
}

func TestEncodingIsStable(t *testing.T) {
	a, _ := Encode(sample())
	b, _ := Encode(sample())
	if !bytes.Equal(a, b) {
		t.Fatal("two encodings of the same values differ")
	}
	if strings.Index(string(a), "Blob") > strings.Index(string(a), "When") {
		t.Fatal("keys are not sorted")
	}
}

func TestBinaryIsRefusedWithAHint(t *testing.T) {
	if _, err := Decode([]byte("bplist00....")); err != ErrBinary {
		t.Fatalf("got %v", err)
	}
}

func TestMalformedReportsTheLine(t *testing.T) {
	_, err := Decode([]byte("<plist>\n<dict>\n<key>a</key>\n<bogus/>\n</dict></plist>"))
	if err == nil || !strings.Contains(err.Error(), "line") {
		t.Fatalf("got %v", err)
	}
}

// On macOS, Apple's own parser must accept what is written here.
func TestPlutilAccepts(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("plutil exists only on macOS")
	}
	if _, err := exec.LookPath("plutil"); err != nil {
		t.Skip("plutil not found")
	}
	data, _ := Encode(sample())
	p := filepath.Join(t.TempDir(), "Info.plist")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("plutil", "-lint", p).CombinedOutput(); err != nil {
		t.Fatalf("plutil refused the file: %s", out)
	}
}
