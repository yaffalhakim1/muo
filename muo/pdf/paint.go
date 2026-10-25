package pdf

import (
	"math"
	"os"
)

// readFile reads a file whole. Reading happens off the UI thread: builders
// never touch the disk.
func readFile(path string) ([]byte, error) { return os.ReadFile(path) }

// fillPath rasterizes the accumulated path with a nonzero or even-odd rule.
// Every subpath is flattened to device lines first, so the same scanline code
// serves painting and clipping.
func (in *pageInterpreter) fillPath(evenOdd bool) {
	if in.collect || !in.hasPath {
		return
	}
	in.finish()
	edges := in.deviceEdges()
	if len(edges) == 0 {
		return
	}
	col := in.state.fill
	r, g, b := byte(col.r*255+.5), byte(col.g*255+.5), byte(col.b*255+.5)
	rows := map[int][]float64{}
	for _, e := range edges {
		for y := int(math.Ceil(minf(e.y0, e.y1))); y <= int(math.Floor(math.Max(e.y0, e.y1))); y++ {
			yc := float64(y) + 0.5
			if (e.y0 <= yc && e.y1 > yc) || (e.y1 <= yc && e.y0 > yc) {
				t := (yc - e.y0) / (e.y1 - e.y0)
				rows[y] = append(rows[y], e.x0+(e.x1-e.x0)*t)
			}
		}
	}
	alpha := in.state.alpha
	for y, xs := range rows {
		if len(xs) < 2 {
			continue
		}
		segments := scanSegments(xs, evenOdd)
		for _, seg := range segments {
			for x := int(math.Ceil(seg[0] - 0.5)); x < int(math.Floor(seg[1])); x++ {
				in.canvas.blend(x, y, r, g, b, alpha)
			}
		}
	}
	in.path = nil
	in.hasPath = false
}

// strokePath draws the path's sides with a one-pixel-wide line, which is what
// reading needs: table rules, underlines and borders.
func (in *pageInterpreter) strokePath() {
	if in.collect || !in.hasPath {
		return
	}
	in.finish()
	edges := in.deviceEdges()
	if len(edges) == 0 {
		in.path = nil
		in.hasPath = false
		return
	}
	col := in.state.stroke
	r, g, b := byte(col.r*255+.5), byte(col.g*255+.5), byte(col.b*255+.5)
	alpha := in.state.strokeAlpha
	width := math.Max(0.6, math.Abs(in.state.lineWidth)*in.scale*math.Hypot(in.state.ctm.c, in.state.ctm.d))
	half := width / 2
	for _, e := range edges {
		dx, dy := e.x1-e.x0, e.y1-e.y0
		length := math.Hypot(dx, dy)
		if length < 0.01 {
			continue
		}
		steps := int(length*2) + 1
		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			cx, cy := e.x0+dx*t, e.y0+dy*t
			for oy := -half; oy <= half; oy += 0.5 {
				for ox := -half; ox <= half; ox += 0.5 {
					in.canvas.blend(int(cx+ox), int(cy+oy), r, g, b, alpha)
				}
			}
		}
	}
	in.path = nil
	in.hasPath = false
}

// clipPath keeps the path as the clip for what follows. A reader cares about
// clips at page margins and inside table cells; a polygon comparison per pixel
// is the simple, correct way to honour them.
func (in *pageInterpreter) clipPath() {
	if in.collect {
		return
	}
	in.finish()
	in.state.clip = append(in.state.clip, in.deviceEdges()...)
	in.path = nil
	in.hasPath = false
}

// deviceEdges flattens every subpath to device-space lines.
func (in *pageInterpreter) deviceEdges() []edge {
	var out []edge
	for _, sp := range in.path {
		for i := 1; i < len(sp.pts); i++ {
			x0, y0 := in.device(sp.pts[i-1].x, sp.pts[i-1].y)
			x1, y1 := in.device(sp.pts[i].x, sp.pts[i].y)
			out = append(out, edge{x0, y0, x1, y1})
		}
	}
	return out
}

// scanSegments turns the crossings of one scanline into filled spans.
func scanSegments(xs []float64, evenOdd bool) [][2]float64 {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j] < xs[j-1]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
	var out [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		if evenOdd {
			out = append(out, [2]float64{xs[i], xs[i+1]})
			continue
		}
		out = append(out, [2]float64{xs[i], xs[i+1]})
	}
	return out
}

