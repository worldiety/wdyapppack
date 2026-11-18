package icns

import (
	"image"
	"image/color"
	"testing"

	"github.com/worldiety/speclink/spec"
	"github.com/worldiety/wdyapppack/requirements/dec"
	"github.com/worldiety/wdyapppack/requirements/fun/icon"
)

func master(n int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(x), uint8(y), 128, 255})
		}
	}
	return img
}

// The icon is produced in Go alone, and contains every size from 16 to 1024.
func TestEveryIconSizeInGo(t *testing.T) {
	entries, err := FromImage(master(1024))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(Encode(entries))
	if err != nil {
		t.Fatal(err)
	}
	sizes := map[int]bool{}
	for i, e := range decoded {
		if e.Type != Slots()[i].Type || e.Size() != Slots()[i].Pixels {
			t.Fatalf("entry %d is %s at %d px", i, e.Type, e.Size())
		}
		sizes[e.Size()] = true
	}
	for _, want := range []int{16, 32, 64, 128, 256, 512, 1024} {
		if !sizes[want] {
			t.Errorf("no %d px image", want)
		}
	}
	spec.Verified(t, icon.RIconSizes, dec.RDecPureGoIcns)
}

func TestNonSquareIsRefused(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 50))
	if _, err := FromImage(img); err == nil {
		t.Fatal("a non square image was accepted")
	}
}

func TestIconsetRoundTrip(t *testing.T) {
	pngs, err := Render(master(512))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for i, s := range Slots() {
		files[s.Iconset] = pngs[i]
	}
	entries, err := FromIconset(files)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(Slots()) {
		t.Fatalf("%d entries", len(entries))
	}
	files["icon_99x99.png"] = pngs[0]
	if _, err := FromIconset(files); err == nil {
		t.Fatal("an unknown file name was accepted")
	}
}

func TestDecodeRefusesGarbage(t *testing.T) {
	if _, err := Decode([]byte("nope")); err == nil {
		t.Fatal("accepted")
	}
	bad := Encode([]Entry{{Type: "ic07", Data: []byte("x")}})
	bad[12] = 0xff // corrupt the entry length
	if _, err := Decode(bad); err == nil {
		t.Fatal("accepted a corrupt length")
	}
}
