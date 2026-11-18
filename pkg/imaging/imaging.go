// Package imaging scales images and cuts them into the shape of a macOS icon.
//
// The standard library decodes and encodes images but cannot resize them. A
// Lanczos filter is the usual choice for icons: it keeps edges crisp when an
// image is made much smaller, which is the common case here — one master
// image becomes every size from 1024 down to 16 pixels.
//
// Scaling works on premultiplied colour, so that transparent pixels do not
// bleed their (invisible) colour into the edges of what is visible.
package imaging

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Resize scales an image to the given size with a Lanczos-3 filter.
func Resize(src image.Image, w, h int) *image.NRGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	in := premultiplied(src)

	// Horizontal pass: sw×sh → w×sh, then vertical: w×sh → w×h.
	tmp := make([]float64, w*sh*4)
	hw := weights(sw, w)
	for y := 0; y < sh; y++ {
		for x := 0; x < w; x++ {
			var acc [4]float64
			for _, c := range hw[x] {
				p := (y*sw + c.i) * 4
				for k := 0; k < 4; k++ {
					acc[k] += in[p+k] * c.w
				}
			}
			copy(tmp[(y*w+x)*4:], acc[:])
		}
	}

	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	vw := weights(sh, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var acc [4]float64
			for _, c := range vw[y] {
				p := (c.i*w + x) * 4
				for k := 0; k < 4; k++ {
					acc[k] += tmp[p+k] * c.w
				}
			}
			a := clamp(acc[3])
			var r, g, bl float64
			if a > 0 {
				r, g, bl = clamp(acc[0]/a), clamp(acc[1]/a), clamp(acc[2]/a)
			}
			o := out.PixOffset(x, y)
			out.Pix[o+0] = uint8(math.Round(r * 255))
			out.Pix[o+1] = uint8(math.Round(g * 255))
			out.Pix[o+2] = uint8(math.Round(bl * 255))
			out.Pix[o+3] = uint8(math.Round(a * 255))
		}
	}
	return out
}

func clamp(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// premultiplied returns the pixels as premultiplied RGBA in [0,1].
func premultiplied(src image.Image) []float64 {
	b := src.Bounds()
	out := make([]float64, b.Dx()*b.Dy()*4)
	if n, ok := src.(*image.NRGBA); ok {
		// The common case, read directly rather than through At, which
		// allocates a colour per pixel.
		i := 0
		for y := b.Min.Y; y < b.Max.Y; y++ {
			row := n.Pix[n.PixOffset(b.Min.X, y):]
			for x := 0; x < b.Dx(); x++ {
				a := float64(row[x*4+3]) / 255
				out[i+0] = float64(row[x*4+0]) / 255 * a
				out[i+1] = float64(row[x*4+1]) / 255 * a
				out[i+2] = float64(row[x*4+2]) / 255 * a
				out[i+3] = a
				i += 4
			}
		}
		return out
	}
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA() // already premultiplied, 16 bit
			out[i+0] = float64(r) / 0xffff
			out[i+1] = float64(g) / 0xffff
			out[i+2] = float64(bl) / 0xffff
			out[i+3] = float64(a) / 0xffff
			i += 4
		}
	}
	return out
}

type contrib struct {
	i int
	w float64
}

func lanczos(x float64) float64 {
	const a = 3
	if x == 0 {
		return 1
	}
	if x <= -a || x >= a {
		return 0
	}
	px := math.Pi * x
	return a * math.Sin(px) * math.Sin(px/a) / (px * px)
}

// weights computes, for every target pixel, which source pixels contribute
// and how much.
func weights(src, dst int) [][]contrib {
	scale := float64(src) / float64(dst)
	support := 3.0
	filterScale := 1.0
	if scale > 1 {
		support *= scale
		filterScale = scale
	}
	out := make([][]contrib, dst)
	for i := 0; i < dst; i++ {
		center := (float64(i)+0.5)*scale - 0.5
		lo := int(math.Floor(center - support))
		hi := int(math.Ceil(center + support))
		var cs []contrib
		sum := 0.0
		for j := lo; j <= hi; j++ {
			w := lanczos((float64(j) - center) / filterScale)
			if w == 0 {
				continue
			}
			k := min(max(j, 0), src-1)
			cs = append(cs, contrib{k, w})
			sum += w
		}
		for k := range cs {
			cs[k].w /= sum
		}
		out[i] = cs
	}
	return out
}

// Grid describes Apple's macOS icon grid: a canvas, and the rounded square
// inside it that the artwork fills.
const (
	GridCanvas = 1024
	GridShape  = 824
	GridMargin = (GridCanvas - GridShape) / 2
)

// Squircle places the image inside Apple's macOS icon grid and cuts it into
// the rounded square macOS 26 expects: a 824 pixel shape centred on a 1024
// pixel canvas, transparent outside. The shape is a superellipse, which is the
// continuous curvature corner Apple draws.
func Squircle(src image.Image) *image.NRGBA {
	art := Resize(ToNRGBA(src), GridShape, GridShape)
	out := image.NewNRGBA(image.Rect(0, 0, GridCanvas, GridCanvas))
	const n = 5.0 // exponent of the superellipse
	const ss = 4  // samples per axis for smooth edges
	half := float64(GridShape) / 2
	for y := 0; y < GridShape; y++ {
		for x := 0; x < GridShape; x++ {
			inside := 0
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					px := (float64(x) + (float64(sx)+0.5)/ss - half) / half
					py := (float64(y) + (float64(sy)+0.5)/ss - half) / half
					if math.Pow(math.Abs(px), n)+math.Pow(math.Abs(py), n) <= 1 {
						inside++
					}
				}
			}
			if inside == 0 {
				continue
			}
			c := art.NRGBAAt(x, y)
			c.A = uint8(math.Round(float64(c.A) * float64(inside) / (ss * ss)))
			out.SetNRGBA(x+GridMargin, y+GridMargin, c)
		}
	}
	return out
}

// OpaqueCorners reports whether any corner pixel of the image is visible. An
// icon whose corners are visible does not follow the rounded shape, and macOS
// 26 puts such an icon into a grey frame.
func OpaqueCorners(img image.Image) bool {
	b := img.Bounds()
	for _, p := range []image.Point{
		{b.Min.X, b.Min.Y}, {b.Max.X - 1, b.Min.Y},
		{b.Min.X, b.Max.Y - 1}, {b.Max.X - 1, b.Max.Y - 1},
	} {
		if _, _, _, a := img.At(p.X, p.Y).RGBA(); a > 0 {
			return true
		}
	}
	return false
}

// ToNRGBA converts any image to NRGBA with its origin at zero.
func ToNRGBA(src image.Image) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), src, b.Min, draw.Src)
	return out
}

var _ color.Model = color.NRGBAModel
