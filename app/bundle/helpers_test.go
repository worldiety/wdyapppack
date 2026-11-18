package bundle

import (
	"bytes"
	"errors"
	"image"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/worldiety/wdyapppack/foundation/problem"
)

func mustDecode(t *testing.T, data []byte) image.Image {
	t.Helper()
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// messagesOf renders every message of an error, for assertions on wording.
func messagesOf(err error) string {
	var e *problem.E
	if !errors.As(err, &e) {
		if err == nil {
			return ""
		}
		return err.Error()
	}
	var b strings.Builder
	for _, m := range e.Messages {
		b.WriteString(m.Code + " " + m.What + " " + m.Why + " " + m.How + "\n")
	}
	return b.String()
}

// sources returns the non-test Go files of this package.
func sources(t *testing.T) string {
	t.Helper()
	files, _ := filepath.Glob("*.go")
	var b strings.Builder
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(data)
	}
	return b.String()
}

var codeLiteral = regexp.MustCompile(`msg(?:How)?\("(WAP-[A-Z0-9-]+)"`)

func quotedCodes(src string) []string {
	var out []string
	for _, m := range codeLiteral.FindAllStringSubmatch(src, -1) {
		out = append(out, m[1])
	}
	return out
}
