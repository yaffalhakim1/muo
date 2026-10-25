package pdf

import "image"

// The rasterizer. A page's content stream is a small stack language of path
// construction, painting, text and images, all in the page's own coordinate
// space. This interpreter covers the operators that decide what a page looks
// like — the ones Acrobat, Preview and every library writer emit — and skips
// the rest, which is what makes it safe: an operator it does not know leaves
// the canvas alone.

// canvas is a premultiplied-free RGB raster, three bytes per pixel.
type canvas struct {
	w, h int
	pix  []byte
}

func newCanvas(w, h int, r, g, b byte) *canvas {
	c := &canvas{w: w, h: h, pix: make([]byte, w*h*3)}
	for i := 0; i < len(c.pix); i += 3 {
		c.pix[i], c.pix[i+1], c.pix[i+2] = r, g, b
	}
	return c
}

func (c *canvas) image() image.Image {
	return &image.RGBA{Pix: c.rgba(), Stride: c.w * 4, Rect: image.Rect(0, 0, c.w, c.h)}
}

// rgba widens the raster to four bytes for MyGo's bitmap.
func (c *canvas) rgba() []byte {
	out := make([]byte, c.w*c.h*4)
	for i, j := 0, 0; i < len(c.pix); i, j = i+3, j+4 {
		out[j], out[j+1], out[j+2], out[j+3] = c.pix[i], c.pix[i+1], c.pix[i+2], 0xff
	}
	return out
}

// blend mixes one pixel towards a color with the given coverage (0–1),
// clipped by the canvas.
func (c *canvas) blend(x, y int, r, g, b byte, a float64) {
	if x < 0 || y < 0 || x >= c.w || y >= c.h || a <= 0 {
		return
	}
	if a > 1 {
		a = 1
	}
	i := (y*c.w + x) * 3
	c.pix[i] = byte(float64(c.pix[i])*(1-a) + float64(r)*a)
	c.pix[i+1] = byte(float64(c.pix[i+1])*(1-a) + float64(g)*a)
	c.pix[i+2] = byte(float64(c.pix[i+2])*(1-a) + float64(b)*a)
}

// graphicsState is everything an operator can change that a later operator
// reads back: the CTM, the colors, the line width and the clip.
type graphicsState struct {
	ctm        matrix
	stroke     devColor
	fill       devColor
	fillSpace  string
	strokeSpace string
	lineWidth  float64
	strokeCap  int
	strokeJoin int
	miter      float64
	dash       []float64
	dashPhase  float64
	alpha      float64 // fill alpha
	strokeAlpha float64
	fontSize   float64
	charSpace  float64
	wordSpace  float64
	hscale     float64
	leading    float64
	rise       float64
	render     int
	clip       []edge
}

func defaultGraphicsState() graphicsState {
	return graphicsState{
		ctm:         identity,
		stroke:      devColor{0, 0, 0},
		fill:        devColor{0, 0, 0},
		fillSpace:   "DeviceGray",
		strokeSpace: "DeviceGray",
		lineWidth:   1,
		miter:       10,
		alpha:       1,
		strokeAlpha: 1,
		hscale:      1,
	}
}

// matrix is the PDF's six-element transform: x' = a·x + c·y + e.
type matrix struct{ a, b, c, d, e, f float64 }

var identity = matrix{1, 0, 0, 1, 0, 0}

func (m matrix) mul(n matrix) matrix {
	return matrix{
		a: m.a*n.a + m.b*n.c,
		b: m.a*n.b + m.b*n.d,
		c: m.c*n.a + m.d*n.c,
		d: m.c*n.b + m.d*n.d,
		e: m.e*n.a + m.f*n.c + n.e,
		f: m.e*n.b + m.f*n.d + n.f,
	}
}

func (m matrix) apply(x, y float64) (float64, float64) {
	return m.a*x + m.c*y + m.e, m.b*x + m.d*y + m.f
}

// edge is one side of a clip polygon in device space.
type edge struct{ x0, y0, x1, y1 float64 }

// devColor is a device color: three channels, 0–1.
type devColor struct{ r, g, b float64 }
