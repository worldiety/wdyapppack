package bundle

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // icons may be given as JPEG
	_ "image/png"
	"path"
	"path/filepath"
	"strings"

	"github.com/worldiety/wdyapppack/foundation/problem"
	"github.com/worldiety/wdyapppack/pkg/icns"
	"github.com/worldiety/wdyapppack/pkg/imaging"
)

// Shapes an icon can be given.
const (
	ShapeNone     = "none"
	ShapeSquircle = "squircle"
)

// makeIcon turns whatever the user gave into the entries of an icon: an image
// is scaled to every size, an .icns file or an icon set is checked and taken
// as it is.
func makeIcon(files Files, p, shape string, c *problem.Collector) ([]icns.Entry, bool) {
	st, err := files.Stat(p)
	if err == nil && !st.Exists {
		c.Error(problem.Failed, msg("WAP-FILE-NOT-FOUND", fmt.Sprintf("The icon %s does not exist.", p)))
		return nil, false
	}

	isImage := !st.Dir && !strings.EqualFold(filepath.Ext(p), ".icns")
	if shape == ShapeSquircle && !isImage {
		c.Error(problem.Usage, msg("WAP-ICON-SHAPE-ICNS", fmt.Sprintf("--icon-shape squircle was given, but %s is not an image.", p)))
		return nil, false
	}

	switch {
	case st.Dir:
		return iconFromSet(files, p, c)
	case strings.EqualFold(filepath.Ext(p), ".icns"):
		data, ok := readInput(files, p, "icon", c)
		if !ok {
			return nil, false
		}
		entries, err := icns.Decode(data)
		if err != nil {
			c.Error(problem.Failed, msg("WAP-ICONSET-INVALID", fmt.Sprintf("%s: %v.", p, err)))
			return nil, false
		}
		return entries, true
	}

	data, ok := readInput(files, p, "icon", c)
	if !ok {
		return nil, false
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") {
		c.Error(problem.Failed, msg("WAP-ICON-FORMAT", fmt.Sprintf("%s is not a PNG or JPEG image.", p)))
		return nil, false
	}
	b := img.Bounds()
	if b.Dx() != b.Dy() {
		c.Error(problem.Failed, msg("WAP-ICON-NOT-SQUARE", fmt.Sprintf("%s is %d×%d pixels.", p, b.Dx(), b.Dy())))
		return nil, false
	}
	if b.Dx() < 512 {
		c.Error(problem.Failed, msg("WAP-ICON-TOO-SMALL", fmt.Sprintf("%s is %d×%d pixels.", p, b.Dx(), b.Dy())))
		return nil, false
	}
	if b.Dx() < 1024 {
		c.Note(msg("WAP-ICON-SMALL", fmt.Sprintf("%s is %d×%d pixels.", p, b.Dx(), b.Dy())))
	}

	if shape == ShapeSquircle {
		img = imaging.Squircle(img)
	} else if imaging.OpaqueCorners(img) {
		c.Note(msg("WAP-ICON-SHAPE", fmt.Sprintf("The corners of %s are not transparent.", p)))
	}

	entries, err := icns.FromImage(img)
	if err != nil {
		c.Error(problem.Failed, msg("WAP-ICON-FORMAT", fmt.Sprintf("%s: %v.", p, err)))
		return nil, false
	}
	return entries, true
}

func iconFromSet(files Files, dir string, c *problem.Collector) ([]icns.Entry, bool) {
	list, err := files.List(dir)
	if err != nil {
		c.Error(problem.Failed, msg("WAP-FILE-UNREADABLE", fmt.Sprintf("The icon set %s cannot be listed: %v.", dir, err)))
		return nil, false
	}
	pngs := map[string][]byte{}
	for _, f := range list {
		if f.Dir || strings.HasPrefix(path.Base(f.Path), ".") {
			continue
		}
		data, ok := readInput(files, filepath.Join(dir, filepath.FromSlash(f.Path)), "icon set image", c)
		if !ok {
			return nil, false
		}
		pngs[f.Path] = data
	}
	entries, err := icns.FromIconset(pngs)
	if err != nil {
		c.Error(problem.Failed, msg("WAP-ICONSET-INVALID", fmt.Sprintf("%s: %v.", dir, err)))
		return nil, false
	}
	return entries, true
}

// iconSizes lists the pixel sizes of the entries, smallest first, without
// repetition.
func iconSizes(entries []icns.Entry) []int {
	seen := map[int]bool{}
	var out []int
	for _, e := range entries {
		if s := e.Size(); s > 0 && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
