package imaging

import (
	"image"
	"image/color"
	"testing"
)

func filled(n int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func TestResizeKeepsSizeAndColour(t *testing.T) {
	src := filled(100, color.NRGBA{200, 100, 50, 255})
	for _, n := range []int{16, 64, 300} {
		out := Resize(src, n, n)
		if out.Bounds().Dx() != n || out.Bounds().Dy() != n {
			t.Fatalf("size %v", out.Bounds())
		}
		c := out.NRGBAAt(n/2, n/2)
		if c.R < 198 || c.R > 202 || c.A != 255 {
			t.Fatalf("colour drifted to %v", c)
		}
	}
}

// A transparent pixel's colour must not bleed into its visible neighbour.
func TestResizeDoesNotBleedTransparentColour(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if x < 32 {
				src.SetNRGBA(x, y, color.NRGBA{255, 0, 0, 0}) // invisible red
			} else {
				src.SetNRGBA(x, y, color.NRGBA{0, 0, 255, 255})
			}
		}
	}
	out := Resize(src, 16, 16)
	for x := 0; x < 16; x++ {
		if c := out.NRGBAAt(x, 8); c.A > 0 && c.R > 8 {
			t.Fatalf("red bled into pixel %d: %v", x, c)
		}
	}
}

func TestSquircleFollowsTheGrid(t *testing.T) {
	src := filled(1024, color.NRGBA{10, 200, 10, 255})
	if !OpaqueCorners(src) {
		t.Fatal("a full square has opaque corners")
	}
	out := Squircle(src)
	if out.Bounds().Dx() != GridCanvas {
		t.Fatalf("canvas %v", out.Bounds())
	}
	if OpaqueCorners(out) {
		t.Fatal("the shaped icon still has opaque corners")
	}
	if a := out.NRGBAAt(GridMargin+5, GridMargin+5).A; a != 0 {
		t.Fatalf("the corner of the shape is not cut: alpha %d", a)
	}
	if a := out.NRGBAAt(512, 512).A; a != 255 {
		t.Fatalf("the centre is not opaque: alpha %d", a)
	}
	if a := out.NRGBAAt(512, GridMargin+1).A; a != 255 {
		t.Fatalf("the middle of the top edge is cut: alpha %d", a)
	}
	if a := out.NRGBAAt(512, GridMargin-2).A; a != 0 {
		t.Fatalf("the margin is not transparent: alpha %d", a)
	}
}
