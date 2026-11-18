// Package icns reads and writes macOS icon files.
//
// An .icns file is a header followed by tagged entries, each holding one image
// size. Every entry written here holds PNG data, which macOS has read in all of
// them since 10.7. The set of entries is the one Apple's iconutil produces from
// an .iconset, so the result is indistinguishable in use.
package icns

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/png"
	"sort"

	"github.com/worldiety/wdyapppack/pkg/imaging"
)

// Entry is one tagged image.
type Entry struct {
	// Type is the four character tag, e.g. "ic10".
	Type string
	// Data is the PNG.
	Data []byte
}

// Slot is one image of an icon: its tag, its size in pixels and its name in an
// .iconset directory.
type Slot struct {
	Type    string
	Pixels  int
	Iconset string
}

// Slots are the images of a complete icon, smallest first.
func Slots() []Slot {
	return []Slot{
		{"icp4", 16, "icon_16x16.png"},
		{"ic11", 32, "icon_16x16@2x.png"},
		{"icp5", 32, "icon_32x32.png"},
		{"ic12", 64, "icon_32x32@2x.png"},
		{"ic07", 128, "icon_128x128.png"},
		{"ic13", 256, "icon_128x128@2x.png"},
		{"ic08", 256, "icon_256x256.png"},
		{"ic14", 512, "icon_256x256@2x.png"},
		{"ic09", 512, "icon_512x512.png"},
		{"ic10", 1024, "icon_512x512@2x.png"},
	}
}

// Render scales a square master image to every slot and encodes each as PNG.
// It returns the PNGs by slot, in the order of Slots.
func Render(master image.Image) ([][]byte, error) {
	b := master.Bounds()
	if b.Dx() != b.Dy() {
		return nil, fmt.Errorf("the image is %d×%d pixels; an icon must be square", b.Dx(), b.Dy())
	}
	if _, ok := master.(*image.NRGBA); !ok {
		master = imaging.ToNRGBA(master)
	}
	// Each size is scaled from the next larger one rather than from the
	// master: halving with a Lanczos filter loses nothing visible, and it
	// makes the whole icon cost little more than its largest reduction.
	sizes := map[int]bool{}
	for _, s := range Slots() {
		sizes[s.Pixels] = true
	}
	order := make([]int, 0, len(sizes))
	for n := range sizes {
		order = append(order, n)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(order)))

	enc := png.Encoder{CompressionLevel: png.DefaultCompression}
	cache := map[int][]byte{}
	src := master
	for _, n := range order {
		img := src
		if img.Bounds().Dx() != n {
			img = imaging.Resize(src, n, n)
		}
		var buf bytes.Buffer
		if err := enc.Encode(&buf, img); err != nil {
			return nil, err
		}
		cache[n] = buf.Bytes()
		if n <= b.Dx() {
			src = img
		}
	}
	var out [][]byte
	for _, s := range Slots() {
		out = append(out, cache[s.Pixels])
	}
	return out, nil
}

// FromImage builds the entries of a complete icon from a master image.
func FromImage(master image.Image) ([]Entry, error) {
	pngs, err := Render(master)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for i, s := range Slots() {
		out = append(out, Entry{Type: s.Type, Data: pngs[i]})
	}
	return out, nil
}

// FromIconset builds entries from the PNG files of an .iconset directory,
// keyed by file name. Files with names iconutil does not know are an error,
// because iconutil would refuse them too.
func FromIconset(files map[string][]byte) ([]Entry, error) {
	byName := map[string]Slot{}
	for _, s := range Slots() {
		byName[s.Iconset] = s
	}
	var out []Entry
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		s, ok := byName[n]
		if !ok {
			return nil, fmt.Errorf("%s is not a name an icon set may contain", n)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(files[n]))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", n, err)
		}
		if cfg.Width != s.Pixels || cfg.Height != s.Pixels {
			return nil, fmt.Errorf("%s is %d×%d pixels, expected %d×%d", n, cfg.Width, cfg.Height, s.Pixels, s.Pixels)
		}
		out = append(out, Entry{Type: s.Type, Data: files[n]})
	}
	if len(out) == 0 {
		return nil, errors.New("the icon set contains no images")
	}
	return order(out), nil
}

func order(es []Entry) []Entry {
	rank := map[string]int{}
	for i, s := range Slots() {
		rank[s.Type] = i
	}
	sort.SliceStable(es, func(i, j int) bool { return rank[es[i].Type] < rank[es[j].Type] })
	return es
}

// Encode writes an .icns file.
func Encode(entries []Entry) []byte {
	size := 8
	for _, e := range entries {
		size += 8 + len(e.Data)
	}
	out := make([]byte, 0, size)
	out = append(out, "icns"...)
	out = binary.BigEndian.AppendUint32(out, uint32(size))
	for _, e := range entries {
		out = append(out, e.Type...)
		out = binary.BigEndian.AppendUint32(out, uint32(8+len(e.Data)))
		out = append(out, e.Data...)
	}
	return out
}

// Decode reads an .icns file.
func Decode(data []byte) ([]Entry, error) {
	if len(data) < 8 || string(data[:4]) != "icns" {
		return nil, errors.New("not an .icns file: it does not start with \"icns\"")
	}
	total := int(binary.BigEndian.Uint32(data[4:]))
	if total > len(data) {
		return nil, fmt.Errorf("the file claims %d bytes but has %d", total, len(data))
	}
	var out []Entry
	for p := 8; p < total; {
		if p+8 > total {
			return nil, errors.New("truncated entry header")
		}
		typ := string(data[p : p+4])
		n := int(binary.BigEndian.Uint32(data[p+4:]))
		if n < 8 || p+n > total {
			return nil, fmt.Errorf("entry %q has an impossible length %d", typ, n)
		}
		out = append(out, Entry{Type: typ, Data: data[p+8 : p+n]})
		p += n
	}
	return out, nil
}

// Size returns the pixel size of a PNG entry, 0 for entries in older formats.
func (e Entry) Size() int {
	cfg, err := png.DecodeConfig(bytes.NewReader(e.Data))
	if err != nil {
		return 0
	}
	return cfg.Width
}
