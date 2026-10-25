//go:build ignore

// Draws docs/og.png, the 1200 by 630 image social previews show.
//
// Run with `go run docs/gen-og.go`. It draws its own letterforms rather than
// loading a font engine, because the strings are fixed and the alternative is a
// build tag file that drags a second dependency into the app's go.mod for a
// marketing image.
//
// The typeface it imitates is the site's: a geometric sans, tight tracking, two
// weights. Each glyph below is a set of strokes on a 1em grid.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

const (
	w = 1200
	h = 630
)

var (
	paper  = color.RGBA{0xfa, 0xf8, 0xf5, 0xff}
	ink    = color.RGBA{0x0a, 0x0a, 0x0a, 0xff}
	muted  = color.RGBA{0x66, 0x66, 0x66, 0xff}
	accent = color.RGBA{0x35, 0x57, 0x7a, 0xff}
	rule   = color.RGBA{0xe5, 0xe5, 0xe1, 0xff}
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{paper}, image.Point{}, draw.Src)

	if icon := loadIcon(); icon != nil {
		draw.Draw(img, image.Rect(96, 96, 272, 272), resize(icon, 176, 176), image.Point{}, draw.Over)
	}

	// The wordmark beside the icon.
	write(img, 300, 210, 34, 0.72, "Muo", ink)

	// The headline, three lines, broken where the thought breaks. The cap height
	// is set so the longest line, "without your eyes", ends before the subhead
	// column begins at x 700.
	write(img, 96, 396, 50, 0.66, "Read for hours", ink)
	write(img, 96, 462, 50, 0.66, "without your eyes", ink)
	write(img, 96, 528, 50, 0.66, "paying for it", ink)

	// The subhead, on the right, balancing the headline.
	write(img, 700, 396, 24, 0.52, "Type solved from your", muted)
	write(img, 700, 430, 24, 0.52, "screen and your distance.", muted)
	write(img, 700, 464, 24, 0.52, "Every line held at", muted)
	write(img, 700, 498, 24, 0.52, "72 characters.", muted)

	// The one figure, in the accent, as the page's gauge does.
	write(img, 700, 566, 20, 0.5, "0.32 degrees, in the band", accent)

	line(img, 96, 600, w-96, 1, rule)

	out, err := os.Create("docs/og.png")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := png.Encode(out, img); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out.Close()
	fmt.Println("wrote docs/og.png")
}

// glyphs maps a rune to the strokes that draw it on a grid where the cap height
// is 1, the baseline is 0, and the advance is about 0.62. Each stroke is
// {x0, y0, x1, y1} in that space; curves are approximated by their endpoints
// because at this size a rounded join reads as a curve.
var glyphs = map[rune][][4]float64{
	'A': {{0, 0, 0.31, 1}, {0.31, 1, 0.62, 0}, {0.12, 0.36, 0.5, 0.36}},
	'B': {{0, 0, 0, 1}, {0, 1, 0.42, 1}, {0.42, 1, 0.55, 0.86}, {0.55, 0.86, 0.55, 0.62}, {0.55, 0.62, 0.42, 0.5}, {0.42, 0.5, 0, 0.5}, {0.42, 0.5, 0.56, 0.38}, {0.56, 0.38, 0.56, 0.14}, {0.56, 0.14, 0.42, 0}, {0.42, 0, 0, 0}},
	'C': {{0.58, 0.86, 0.42, 1}, {0.42, 1, 0.16, 1}, {0.16, 1, 0, 0.8}, {0, 0.8, 0, 0.2}, {0, 0.2, 0.16, 0}, {0.16, 0, 0.42, 0}, {0.42, 0, 0.58, 0.14}},
	'D': {{0, 0, 0, 1}, {0, 1, 0.36, 1}, {0.36, 1, 0.56, 0.8}, {0.56, 0.8, 0.56, 0.2}, {0.56, 0.2, 0.36, 0}, {0.36, 0, 0, 0}},
	'E': {{0.54, 1, 0, 1}, {0, 1, 0, 0}, {0, 0, 0.54, 0}, {0, 0.5, 0.44, 0.5}},
	'F': {{0.54, 1, 0, 1}, {0, 1, 0, 0}, {0, 0.5, 0.44, 0.5}},
	'G': {{0.58, 0.86, 0.42, 1}, {0.42, 1, 0.16, 1}, {0.16, 1, 0, 0.8}, {0, 0.8, 0, 0.2}, {0, 0.2, 0.16, 0}, {0.16, 0, 0.42, 0}, {0.42, 0, 0.58, 0.16}, {0.58, 0.16, 0.58, 0.44}, {0.58, 0.44, 0.34, 0.44}},
	'H': {{0, 0, 0, 1}, {0.56, 0, 0.56, 1}, {0, 0.5, 0.56, 0.5}},
	'I': {{0.1, 1, 0.1, 0}, {0, 1, 0.2, 1}, {0, 0, 0.2, 0}},
	'K': {{0, 0, 0, 1}, {0.56, 1, 0, 0.42}, {0.14, 0.56, 0.56, 0}},
	'L': {{0, 1, 0, 0}, {0, 0, 0.5, 0}},
	'M': {{0, 0, 0, 1}, {0, 1, 0.3, 0.42}, {0.3, 0.42, 0.6, 1}, {0.6, 1, 0.6, 0}},
	'N': {{0, 0, 0, 1}, {0, 1, 0.56, 0}, {0.56, 0, 0.56, 1}},
	'O': {{0.16, 1, 0.42, 1}, {0.42, 1, 0.58, 0.8}, {0.58, 0.8, 0.58, 0.2}, {0.58, 0.2, 0.42, 0}, {0.42, 0, 0.16, 0}, {0.16, 0, 0, 0.2}, {0, 0.2, 0, 0.8}, {0, 0.8, 0.16, 1}},
	'P': {{0, 0, 0, 1}, {0, 1, 0.4, 1}, {0.4, 1, 0.56, 0.86}, {0.56, 0.86, 0.56, 0.62}, {0.56, 0.62, 0.4, 0.48}, {0.4, 0.48, 0, 0.48}},
	'R': {{0, 0, 0, 1}, {0, 1, 0.4, 1}, {0.4, 1, 0.56, 0.86}, {0.56, 0.86, 0.56, 0.62}, {0.56, 0.62, 0.4, 0.48}, {0.4, 0.48, 0, 0.48}, {0.3, 0.48, 0.58, 0}},
	'S': {{0.56, 0.86, 0.4, 1}, {0.4, 1, 0.16, 1}, {0.16, 1, 0, 0.86}, {0, 0.86, 0, 0.66}, {0, 0.66, 0.16, 0.54}, {0.16, 0.54, 0.4, 0.48}, {0.4, 0.48, 0.56, 0.36}, {0.56, 0.36, 0.56, 0.14}, {0.56, 0.14, 0.4, 0}, {0.4, 0, 0.16, 0}, {0.16, 0, 0, 0.14}},
	'T': {{0, 1, 0.56, 1}, {0.28, 1, 0.28, 0}},
	'U': {{0, 1, 0, 0.2}, {0, 0.2, 0.16, 0}, {0.16, 0, 0.42, 0}, {0.42, 0, 0.58, 0.2}, {0.58, 0.2, 0.58, 1}},
	'V': {{0, 1, 0.29, 0}, {0.29, 0, 0.58, 1}},
	'W': {{0, 1, 0.14, 0}, {0.14, 0, 0.3, 0.66}, {0.3, 0.66, 0.46, 0}, {0.46, 0, 0.62, 1}},
	'X': {{0, 1, 0.56, 0}, {0.56, 1, 0, 0}},
	'Y': {{0, 1, 0.28, 0.5}, {0.56, 1, 0.28, 0.5}, {0.28, 0.5, 0.28, 0}},
	'a': {{0.12, 0.66, 0.34, 0.72}, {0.34, 0.72, 0.5, 0.64}, {0.5, 0.64, 0.5, 0}, {0.5, 0.12, 0.34, 0}, {0.34, 0, 0.12, 0.08}, {0.12, 0.08, 0.12, 0.3}, {0.12, 0.3, 0.5, 0.36}},
	'b': {{0, 1, 0, 0}, {0, 0.12, 0.16, 0}, {0.16, 0, 0.38, 0}, {0.38, 0, 0.5, 0.16}, {0.5, 0.16, 0.5, 0.56}, {0.5, 0.56, 0.38, 0.72}, {0.38, 0.72, 0.16, 0.72}, {0.16, 0.72, 0, 0.6}},
	'c': {{0.5, 0.6, 0.36, 0.72}, {0.36, 0.72, 0.14, 0.72}, {0.14, 0.72, 0, 0.58}, {0, 0.58, 0, 0.14}, {0, 0.14, 0.14, 0}, {0.14, 0, 0.36, 0}, {0.36, 0, 0.5, 0.12}},
	'd': {{0.5, 1, 0.5, 0}, {0.5, 0.12, 0.34, 0}, {0.34, 0, 0.12, 0}, {0.12, 0, 0, 0.16}, {0, 0.16, 0, 0.56}, {0, 0.56, 0.12, 0.72}, {0.12, 0.72, 0.34, 0.72}, {0.34, 0.72, 0.5, 0.6}},
	'e': {{0, 0.38, 0.5, 0.38}, {0.5, 0.38, 0.5, 0.58}, {0.5, 0.58, 0.36, 0.72}, {0.36, 0.72, 0.14, 0.72}, {0.14, 0.72, 0, 0.58}, {0, 0.58, 0, 0.14}, {0, 0.14, 0.14, 0}, {0.14, 0, 0.36, 0}, {0.36, 0, 0.5, 0.1}},
	'f': {{0.42, 1, 0.26, 1}, {0.26, 1, 0.16, 0.88}, {0.16, 0.88, 0.16, 0}, {0.04, 0.7, 0.34, 0.7}},
	'g': {{0.5, 0.72, 0.5, -0.18}, {0.5, -0.18, 0.36, -0.3}, {0.36, -0.3, 0.12, -0.3}, {0.5, 0.12, 0.34, 0}, {0.34, 0, 0.12, 0}, {0.12, 0, 0, 0.14}, {0, 0.14, 0, 0.5}, {0, 0.5, 0.12, 0.66}, {0.12, 0.66, 0.34, 0.66}, {0.34, 0.66, 0.5, 0.54}},
	'h': {{0, 1, 0, 0}, {0, 0.56, 0.16, 0.72}, {0.16, 0.72, 0.38, 0.72}, {0.38, 0.72, 0.5, 0.58}, {0.5, 0.58, 0.5, 0}},
	'i': {{0.06, 0.72, 0.06, 0}, {0.06, 0.94, 0.06, 0.99}},
	'k': {{0, 1, 0, 0}, {0.46, 0.72, 0, 0.32}, {0.14, 0.44, 0.46, 0}},
	'l': {{0.06, 1, 0.06, 0}},
	'm': {{0, 0.72, 0, 0}, {0, 0.56, 0.12, 0.72}, {0.12, 0.72, 0.28, 0.72}, {0.28, 0.72, 0.36, 0.58}, {0.36, 0.58, 0.36, 0}, {0.36, 0.56, 0.48, 0.72}, {0.48, 0.72, 0.64, 0.72}, {0.64, 0.72, 0.72, 0.58}, {0.72, 0.58, 0.72, 0}},
	'n': {{0, 0.72, 0, 0}, {0, 0.56, 0.16, 0.72}, {0.16, 0.72, 0.38, 0.72}, {0.38, 0.72, 0.5, 0.58}, {0.5, 0.58, 0.5, 0}},
	'o': {{0.14, 0.72, 0.36, 0.72}, {0.36, 0.72, 0.5, 0.58}, {0.5, 0.58, 0.5, 0.14}, {0.5, 0.14, 0.36, 0}, {0.36, 0, 0.14, 0}, {0.14, 0, 0, 0.14}, {0, 0.14, 0, 0.58}, {0, 0.58, 0.14, 0.72}},
	'p': {{0, 0.72, 0, -0.32}, {0, 0.12, 0.16, 0}, {0.16, 0, 0.38, 0}, {0.38, 0, 0.5, 0.16}, {0.5, 0.16, 0.5, 0.56}, {0.5, 0.56, 0.38, 0.72}, {0.38, 0.72, 0.16, 0.72}, {0.16, 0.72, 0, 0.6}},
	'r': {{0, 0.72, 0, 0}, {0, 0.52, 0.18, 0.72}, {0.18, 0.72, 0.4, 0.72}},
	's': {{0.48, 0.6, 0.36, 0.72}, {0.36, 0.72, 0.12, 0.72}, {0.12, 0.72, 0, 0.62}, {0, 0.62, 0.06, 0.44}, {0.06, 0.44, 0.24, 0.38}, {0.24, 0.38, 0.42, 0.3}, {0.42, 0.3, 0.48, 0.12}, {0.48, 0.12, 0.36, 0}, {0.36, 0, 0.12, 0}, {0.12, 0, 0, 0.12}},
	't': {{0.16, 1, 0.16, 0.16}, {0.16, 0.16, 0.28, 0}, {0.28, 0, 0.42, 0}, {0.04, 0.7, 0.36, 0.7}},
	'u': {{0, 0.72, 0, 0.16}, {0, 0.16, 0.12, 0}, {0.12, 0, 0.34, 0}, {0.34, 0, 0.5, 0.16}, {0.5, 0.72, 0.5, 0}},
	'v': {{0, 0.72, 0.24, 0}, {0.24, 0, 0.48, 0.72}},
	'w': {{0, 0.72, 0.12, 0}, {0.12, 0, 0.26, 0.5}, {0.26, 0.5, 0.4, 0}, {0.4, 0, 0.52, 0.72}},
	'x': {{0, 0.72, 0.46, 0}, {0.46, 0.72, 0, 0}},
	'y': {{0, 0.72, 0.26, 0.1}, {0.5, 0.72, 0.16, -0.3}},
	'z': {{0, 0.72, 0.46, 0.72}, {0.46, 0.72, 0, 0}, {0, 0, 0.46, 0}},
	'0': {{0.14, 1, 0.36, 1}, {0.36, 1, 0.5, 0.8}, {0.5, 0.8, 0.5, 0.2}, {0.5, 0.2, 0.36, 0}, {0.36, 0, 0.14, 0}, {0.14, 0, 0, 0.2}, {0, 0.2, 0, 0.8}, {0, 0.8, 0.14, 1}},
	'2': {{0, 0.8, 0.14, 1}, {0.14, 1, 0.4, 1}, {0.4, 1, 0.5, 0.86}, {0.5, 0.86, 0.5, 0.66}, {0.5, 0.66, 0, 0}, {0, 0, 0.52, 0}},
	'7': {{0, 1, 0.5, 1}, {0.5, 1, 0.18, 0}},
	'3': {{0, 0.84, 0.14, 1}, {0.14, 1, 0.4, 1}, {0.4, 1, 0.5, 0.86}, {0.5, 0.86, 0.5, 0.66}, {0.5, 0.66, 0.4, 0.56}, {0.4, 0.56, 0.14, 0.56}, {0.4, 0.56, 0.5, 0.44}, {0.5, 0.44, 0.5, 0.16}, {0.5, 0.16, 0.4, 0}, {0.4, 0, 0.14, 0}, {0.14, 0, 0, 0.16}},
	'.': {{0.06, 0.1, 0.06, 0}},
	',': {{0.1, 0.1, 0.02, -0.16}},
	' ': {},
}

// write draws s at a baseline, with the cap height given in pixels.
func write(img *image.RGBA, x, y int, capPx float64, weight float64, s string, c color.RGBA) {
	pen := float64(x)
	// The stroke width scales with the cap height, which is what makes the
	// 64 pixel headline read as a headline rather than a bolded caption.
	stroke := capPx * 0.11 * (weight / 0.66)
	for _, r := range s {
		strokes, ok := glyphs[r]
		if !ok {
			pen += capPx * 0.5
			continue
		}
		for _, st := range strokes {
			seg(img,
				pen+st[0]*capPx, float64(y)-st[1]*capPx,
				pen+st[2]*capPx, float64(y)-st[3]*capPx,
				stroke, c)
		}
		pen += capPx * 0.62
	}
}

// seg draws a thick segment with round caps.
func seg(img *image.RGBA, x0, y0, x1, y1, width float64, c color.RGBA) {
	dx, dy := x1-x0, y1-y0
	length := math.Hypot(dx, dy)
	half := width / 2
	minX := int(math.Min(x0, x1) - half - 1)
	maxX := int(math.Max(x0, x1) + half + 1)
	minY := int(math.Min(y0, y1) - half - 1)
	maxY := int(math.Max(y0, y1) + half + 1)
	for py := minY; py <= maxY; py++ {
		for px := minX; px <= maxX; px++ {
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			var d float64
			if length < 0.01 {
				d = math.Hypot(float64(px)-x0, float64(py)-y0)
			} else {
				t := ((float64(px)-x0)*dx + (float64(py)-y0)*dy) / (length * length)
				t = math.Max(0, math.Min(1, t))
				d = math.Hypot(float64(px)-(x0+dx*t), float64(py)-(y0+dy*t))
			}
			// A one pixel feather, so the strokes are not jagged.
			if d <= half-0.5 {
				img.Set(px, py, c)
			} else if d <= half+0.5 {
				a := (half + 0.5 - d)
				img.Set(px, py, mix(img.RGBAAt(px, py), c, a))
			}
		}
	}
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	t = math.Max(0, math.Min(1, t))
	f := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + 0.5) }
	return color.RGBA{f(a.R, b.R), f(a.G, b.G), f(a.B, b.B), 0xff}
}

func line(img *image.RGBA, x0, x1, y, width int, c color.RGBA) {
	for py := y; py < y+width; py++ {
		for px := x0; px < x1; px++ {
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			img.Set(px, py, c)
		}
	}
}

func loadIcon() image.Image {
	f, err := os.Open("assets/icon.png")
	if err != nil {
		return nil
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil
	}
	return src
}

func resize(src image.Image, nw, nh int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	sb := src.Bounds()
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			dst.Set(x, y, src.At(sb.Min.X+x*sb.Dx()/nw, sb.Min.Y+y*sb.Dy()/nh))
		}
	}
	return dst
}
