//go:build ignore

// Draws Muo's app icon at every size the platforms ask for.
//
// Run with `go run assets/gen-icon.go`. It writes icon.png at 1024, which
// mygo.json uses for the Windows and Linux icons, and icon.icns for macOS.
//
// The mark is drawn rather than rendered from SVG because the icon needs a
// gradient plate, a cast shadow and antialiased rounded corners, and the SVG
// parsers available here leave out filters. Drawing it directly gives exact
// control over each of those at every size.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const size = 1024

func main() {
	// The icon at every size the platforms ask for. Each is drawn rather than
	// downscaled, because the mark has two weights: the full one for large sizes
	// and a simplified one that survives a 32-pixel taskbar.
	sizes := []int{1024, 512, 256, 128, 64, 32, 16}
	for _, n := range sizes {
		img := draw(n)
		out := filepath.Join("assets", fmt.Sprintf("icon-%d.png", n))
		f, err := os.Create(out)
		if err != nil {
			fail(err)
		}
		if err := png.Encode(f, img); err != nil {
			fail(err)
		}
		f.Close()
		fmt.Printf("wrote %s (%d×%d)\n", out, n, n)
	}

	// icon.png is the 1024, which is what mygo.json names for Windows and Linux.
	if err := os.WriteFile(filepath.Join("assets", "icon.png"), mustRead(filepath.Join("assets", "icon-1024.png")), 0o644); err != nil {
		fail(err)
	}
}

// mustRead reads a file the run just wrote.
func mustRead(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	return data
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen-icon:", err)
	os.Exit(1)
}

// draw renders the icon at n by n pixels.
func draw(n int) *image.RGBA {
	// Everything is laid out in the 1024 space the icon was designed in, then
	// scaled, so one set of numbers serves every size.
	s := float64(n) / size

	img := image.NewRGBA(image.Rect(0, 0, n, n))

	// The plate: Apple's icon grid, 824 of 1024 with the artwork inside it. At
	// the sizes a taskbar shows, the margin eats the mark, so the plate grows to
	// nearly the whole canvas and the mark with it.
	plateX, plateY, plateW, plateR := 100.0, 100.0, 824.0, 186.0
	if n <= 64 {
		plateX, plateY, plateW, plateR = 8, 8, float64(size) - 16, 190
	}

	// The cast shadow, drawn first so the plate sits on it. It is a soft ellipse
	// offset down, which is what a lit surface on a desktop does.
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			dx := (float64(x)/s - (plateX + plateW/2)) / (plateW * 0.54)
			dy := (float64(y)/s - (plateY + plateW/2 + 30)) / (plateW * 0.46)
			d := math.Sqrt(dx*dx + dy*dy)
			if d < 1 {
				// A soft falloff, strongest just under the plate.
				a := math.Pow(1-d, 2.0) * 0.42
				img.Set(x, y, color.RGBA{52, 41, 30, uint8(a * 255)})
			}
		}
	}

	// The plate itself, rounded, with a vertical gradient from near-white to
	// warm paper, and a hairline of light along its top edge.
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			fx, fy := float64(x)/s, float64(y)/s
			if !inRoundRect(fx, fy, plateX, plateY, plateW, plateW, plateR) {
				continue
			}
			// Antialias the edge by how far inside the shape the pixel centre is.
			cov := coverage(fx, fy, plateX, plateY, plateW, plateW, plateR)
			if cov <= 0 {
				continue
			}
			t := (fy - plateY) / plateW
			top := mix(rgb(255, 253, 249), rgb(242, 236, 226), t)
			// The rim: a white line just inside the top edge.
			rim := 0.0
			if !inRoundRect(fx, fy, plateX+3, plateY+3, plateW-6, plateW-6, plateR-3) {
				rim = 0.9 * (1 - t*2)
				if rim < 0 {
					rim = 0
				}
			}
			c := mix(top, rgb(255, 255, 255), rim)
			blend(img, x, y, c, cov)
		}
	}

	// The border, one warm hairline just inside the plate's edge.
	strokeRoundRect(img, s, plateX, plateY, plateW, plateW, plateR, 3, rgb(224, 216, 203), 1)

	// The mark has three weights. The full one, with every line of text; a
	// simplified one for the sizes a taskbar shows; and at 16 pixels the letter
	// alone, because six hairlines and a stroke become one dark block.
	simplified := n <= 64
	tiny := n <= 16

	// The two columns of text, centred in the plate and drawn larger than the
	// SVG's numbers, so the mark fills the plate the way an app icon should: the
	// artwork reads at 32 pixels, where a small mark turns to mush.
	const dy = 74.0
	const k = 1.16 // the mark's scale, about its own centre
	at := func(v float64) float64 { return (v-460)*k + 460 + dy }
	w := func(v float64) float64 { return v * k }
	inkTop, inkBottom := rgb(43, 39, 36), rgb(20, 18, 16)

	// thick widens every stroke at the small sizes, so a 32-pixel icon keeps its
	// weight instead of fading.
	thick := 1.0
	if simplified {
		thick = 1.9
	}

	accentTop, accentBottom := rgb(74, 111, 150), rgb(47, 75, 104)

	if tiny {
		// The letter alone: two stems and the valley between them, at a weight
		// that survives eight pixels of stroke.
		stroke := w(74)
		line(img, s, at(300), at(300), at(300), at(560), stroke, func(t float64) color.RGBA {
			return mix(inkTop, inkBottom, t)
		})
		line(img, s, at(712), at(300), at(712), at(560), stroke, func(t float64) color.RGBA {
			return mix(inkTop, inkBottom, t)
		})
		line(img, s, at(300), at(300), at(506), at(524), stroke, func(t float64) color.RGBA {
			return mix(accentTop, accentBottom, t)
		})
		line(img, s, at(712), at(300), at(506), at(524), stroke, func(t float64) color.RGBA {
			return mix(accentTop, accentBottom, t)
		})
		return img
	}

	col := func(x float64) {
		// The heading rule.
		fillRoundRect(img, s, at(x), at(300), w(176), w(26*thick), w(13*thick), func(t float64) color.RGBA {
			return mix(inkTop, inkBottom, t)
		})
		lines := []struct {
			y, w float64
		}{
			{362, 176}, {404, 176}, {446, 140}, {488, 176}, {530, 112},
		}
		if simplified {
			// Three lines, evenly spaced, the last shorter: enough to read as a
			// column of text without crowding the pixels.
			lines = []struct {
				y, w float64
			}{
				{382, 176}, {452, 176}, {522, 124},
			}
		}
		for _, ln := range lines {
			fillRoundRect(img, s, at(x), at(ln.y), w(ln.w), w(16*thick), w(8*thick), func(t float64) color.RGBA {
				c := mix(inkTop, inkBottom, t)
				c.A = 210
				return c
			})
		}
	}
	col(256)
	col(592)

	// The valley: the M's middle, in the accent, between the columns.
	stroke := 30.0
	if simplified {
		stroke = 40
	}
	line(img, s, at(432), at(300), at(512), at(424), w(stroke), func(t float64) color.RGBA {
		return mix(accentTop, accentBottom, t)
	})
	line(img, s, at(512), at(424), at(592), at(300), w(stroke), func(t float64) color.RGBA {
		return mix(accentTop, accentBottom, t)
	})
	// Round the joints, as the SVG does with linecap and linejoin.
	disc(img, s, at(432), at(300), w(stroke/2), accentTop)
	disc(img, s, at(512), at(424), w(stroke/2), accentBottom)
	disc(img, s, at(592), at(300), w(stroke/2), accentTop)

	// The baseline under both columns.
	fillRoundRect(img, s, at(256), at(612), w(512), w(6*thick), w(3*thick), func(t float64) color.RGBA {
		c := mix(inkTop, inkBottom, t)
		c.A = 56
		return c
	})

	return img
}

// rgb is an opaque colour.
func rgb(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

// mix blends two colours in sRGB, t from 0 to 1.
func mix(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	m := func(x, y uint8) uint8 {
		return uint8(float64(x)*(1-t) + float64(y)*t + 0.5)
	}
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), m(a.A, b.A)}
}

// blend puts c over the pixel with coverage a.
func blend(img *image.RGBA, x, y int, c color.RGBA, a float64) {
	if a <= 0 {
		return
	}
	if a > 1 {
		a = 1
	}
	src := float64(c.A) / 255 * a
	if src <= 0 {
		return
	}
	i := img.PixOffset(x, y)
	d := img.Pix[i : i+4 : i+4]
	da := float64(d[3]) / 255
	out := src + da*(1-src)
	if out <= 0 {
		return
	}
	for k := 0; k < 3; k++ {
		sc := [...]float64{float64(c.R), float64(c.G), float64(c.B)}[k]
		dc := float64(d[k])
		d[k] = uint8((sc*src + dc*da*(1-src)) / out)
	}
	d[3] = uint8(out*255 + 0.5)
}

// inRoundRect reports whether a point is inside a rounded rectangle.
func inRoundRect(x, y, rx, ry, w, h, r float64) bool {
	if x < rx || y < ry || x > rx+w || y > ry+h {
		return false
	}
	// Distance into the corner, if the point is in one.
	cx, cy := x, y
	switch {
	case x < rx+r && y < ry+r:
		cx, cy = rx+r, ry+r
	case x > rx+w-r && y < ry+r:
		cx, cy = rx+w-r, ry+r
	case x < rx+r && y > ry+h-r:
		cx, cy = rx+r, ry+h-r
	case x > rx+w-r && y > ry+h-r:
		cx, cy = rx+w-r, ry+h-r
	default:
		return true
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy <= r*r
}

// coverage antialiases a rounded rectangle's edge: 1 well inside, 0 outside, and
// a fraction within a pixel of the boundary.
func coverage(x, y, rx, ry, w, h, r float64) float64 {
	if inRoundRect(x, y, rx, ry, w, h, r) {
		// How far from the edge, in the cheapest useful way: test a few
		// concentric offsets.
		for _, d := range []float64{0.25, 0.5, 0.75} {
			if !inRoundRect(x+d, y+d, rx, ry, w, h, r) ||
				!inRoundRect(x-d, y+d, rx, ry, w, h, r) ||
				!inRoundRect(x+d, y-d, rx, ry, w, h, r) ||
				!inRoundRect(x-d, y-d, rx, ry, w, h, r) {
				return 0.75
			}
		}
		return 1
	}
	return 0
}

// fillRoundRect fills a rounded rectangle in design space, the colour given by
// a function of the vertical position within it.
func fillRoundRect(img *image.RGBA, s, x, y, w, h, r float64, colour func(t float64) color.RGBA) {
	x0, y0 := int((x-1)*s), int((y-1)*s)
	x1, y1 := int((x+w+1)*s), int((y+h+1)*s)
	for py := y0; py <= y1; py++ {
		for px := x0; px <= x1; px++ {
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			fx, fy := (float64(px)+0.5)/s, (float64(py)+0.5)/s
			// A 4×4 sample for the edge, which is enough at these sizes.
			cov := 0.0
			for sy := 0; sy < 2; sy++ {
				for sx := 0; sx < 2; sx++ {
					if inRoundRect(fx+float64(sx)-0.5/s/2, fy+float64(sy)-0.5/s/2, x, y, w, h, r) {
						cov += 0.25
					}
				}
			}
			if cov > 0 {
				t := (fy - y) / h
				blend(img, px, py, colour(t), cov)
			}
		}
	}
}

// line draws a thick segment with round caps.
func line(img *image.RGBA, s, x0, y0, x1, y1, width float64, colour func(t float64) color.RGBA) {
	dx, dy := x1-x0, y1-y0
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	half := width / 2
	minX := math.Min(x0, x1) - half
	maxX := math.Max(x0, x1) + half
	minY := math.Min(y0, y1) - half
	maxY := math.Max(y0, y1) + half
	for py := int(minY*s) - 1; py <= int(maxY*s)+1; py++ {
		for px := int(minX*s) - 1; px <= int(maxX*s)+1; px++ {
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			fx, fy := (float64(px)+0.5)/s, (float64(py)+0.5)/s
			cov := 0.0
			for sy := 0; sy < 2; sy++ {
				for sx := 0; sx < 2; sx++ {
					sfx := fx + float64(sx)-0.5/s/2
					sfy := fy + float64(sy)-0.5/s/2
					if distToSegment(sfx, sfy, x0, y0, dx, dy, length) <= half {
						cov += 0.25
					}
				}
			}
			if cov > 0 {
				// The gradient runs along the segment's y.
				t := (fy - math.Min(y0, y1)) / math.Max(math.Abs(y1-y0), 1)
				blend(img, px, py, colour(t), cov)
			}
		}
	}
}

// distToSegment is the distance from a point to a segment.
func distToSegment(px, py, x0, y0, dx, dy, length float64) float64 {
	t := ((px-x0)*dx + (py-y0)*dy) / (length * length)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return math.Hypot(px-(x0+dx*t), py-(y0+dy*t))
}

// disc draws a filled circle.
func disc(img *image.RGBA, s, cx, cy, r float64, c color.RGBA) {
	for py := int((cy-r)*s) - 1; py <= int((cy+r)*s)+1; py++ {
		for px := int((cx-r)*s) - 1; px <= int((cx+r)*s)+1; px++ {
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			fx, fy := (float64(px)+0.5)/s, (float64(py)+0.5)/s
			cov := 0.0
			for sy := 0; sy < 2; sy++ {
				for sx := 0; sx < 2; sx++ {
					sfx := fx + float64(sx)-0.5/s/2
					sfy := fy + float64(sy)-0.5/s/2
					if math.Hypot(sfx-cx, sfy-cy) <= r {
						cov += 0.25
					}
				}
			}
			if cov > 0 {
				blend(img, px, py, c, cov)
			}
		}
	}
}

// strokeRoundRect draws a rounded rectangle's outline.
func strokeRoundRect(img *image.RGBA, s, x, y, w, h, r, width float64, c color.RGBA, alpha float64) {
	for py := int((y-width)*s) - 1; py <= int((y+h+width)*s)+1; py++ {
		for px := int((x-width)*s) - 1; px <= int((x+w+width)*s)+1; px++ {
			if px < 0 || py < 0 || px >= img.Bounds().Dx() || py >= img.Bounds().Dy() {
				continue
			}
			fx, fy := (float64(px)+0.5)/s, (float64(py)+0.5)/s
			// On the edge: inside the outer rect but not the inner one.
			if !inRoundRect(fx, fy, x-width/2, y-width/2, w+width, h+width, r+width/2) {
				continue
			}
			if inRoundRect(fx, fy, x+width/2, y+width/2, w-width, h-width, r-width/2) {
				continue
			}
			cc := c
			cc.A = uint8(float64(cc.A) * alpha)
			blend(img, px, py, cc, 1)
		}
	}
}
