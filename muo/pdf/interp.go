package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strconv"
)

// pageInterpreter walks one content stream. Its operands are kept in a small
// array rather than allocated per operator; anything that does not fit the
// operators below is dropped.
type pageInterpreter struct {
	file   *pdfFile
	page   int
	canvas *canvas
	scale  float64 // pixels per point
	state  graphicsState
	stack  []graphicsState

	ops []pdfValue // the current operator's operands

	path    []subpath
	current subpath
	hasPath bool

	text      textState
	inText    bool
	font      *pdfFont
	resources pdfValue

	forms map[int]bool // form XObjects being drawn, to stop recursion
	depth int

	// collect asks the interpreter to record the text it would draw instead of
	// drawing it, which is how a page is read as text rather than as pixels.
	collect bool
	runs    []TextRun
}

// The interpreter is the same object for every page: one file, one page, one
// canvas at a time.

// textState is what text operators read back.
type textState struct {
	m      matrix // text matrix
	lineM  matrix // line matrix: the start of the current line
	inLine bool
}

type subpath struct {
	pts    []point
	closed bool
}

type point struct{ x, y float64 }

func (in *pageInterpreter) run(content []byte, depth int) error {
	if depth > 12 {
		return fmt.Errorf("%w: forms nest too deep", errPDF)
	}
	in.depth = depth
	in.resources = in.file.resolve(mustGet(in.file.pages[in.page], "Resources"))
	at := 0
	for at < len(content) {
		v, next := in.file.parse(content, at, 0)
		if next == at {
			// Unparseable byte: step over it, never stall.
			at++
			continue
		}
		at = next
		if isOperator(v) {
			in.operate(v.name)
			in.ops = in.ops[:0]
			continue
		}
		// An operand, a name such as /F1 included: kept until its operator.
		in.ops = append(in.ops, v)
	}
	return nil
}

// num returns operand i as a number, or def when it is absent.
func (in *pageInterpreter) num(i int, def float64) float64 {
	if i < len(in.ops) && in.ops[i].kind == pdfNumber {
		return in.ops[i].num
	}
	return def
}

func (in *pageInterpreter) operand(i int) (pdfValue, bool) {
	if i < len(in.ops) {
		return in.ops[i], true
	}
	return pdfValue{}, false
}

// rawName returns operand i as a name.
func (in *pageInterpreter) rawName(i int) string {
	if i < len(in.ops) && in.ops[i].kind == pdfName {
		return in.ops[i].name
	}
	return ""
}

// operate applies one operator.
func (in *pageInterpreter) operate(op string) {
	switch op {
	// Graphics state ---------------------------------------------------------
	case "q":
		in.stack = append(in.stack, in.state)
		if len(in.stack) > 256 {
			in.stack = in.stack[len(in.stack)-256:]
		}
	case "Q":
		if n := len(in.stack); n > 0 {
			in.state = in.stack[n-1]
			in.stack = in.stack[:n-1]
		}
	case "cm":
		if len(in.ops) >= 6 {
			m := matrix{in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0), in.num(4, 0), in.num(5, 0)}
			in.state.ctm = m.mul(in.state.ctm)
		}
	case "w":
		in.state.lineWidth = in.num(0, 1)
	case "J":
		in.state.strokeCap = int(in.num(0, 0))
	case "j":
		in.state.strokeJoin = int(in.num(0, 0))
	case "M":
		in.state.miter = in.num(0, 10)
	case "d":
		if v, ok := in.operand(0); ok && v.kind == pdfArray {
			in.state.dash = nil
			for _, e := range v.array {
				in.state.dash = append(in.state.dash, e.num)
			}
		}
		in.state.dashPhase = in.num(1, 0)
	case "gs":
		in.applyExtGState(in.rawName(0))

	// Color ------------------------------------------------------------------
	case "g":
		in.state.fill = gray(in.num(0, 0))
		in.state.fillSpace = "DeviceGray"
	case "G":
		in.state.stroke = gray(in.num(0, 0))
		in.state.strokeSpace = "DeviceGray"
	case "rg":
		in.state.fill = rgb(in.num(0, 0), in.num(1, 0), in.num(2, 0))
		in.state.fillSpace = "DeviceRGB"
	case "RG":
		in.state.stroke = rgb(in.num(0, 0), in.num(1, 0), in.num(2, 0))
		in.state.strokeSpace = "DeviceRGB"
	case "k":
		in.state.fill = cmyk(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0))
		in.state.fillSpace = "DeviceCMYK"
	case "K":
		in.state.stroke = cmyk(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0))
		in.state.strokeSpace = "DeviceCMYK"
	case "cs":
		if op := in.rawName(0); op != "" {
			in.state.fillSpace = op
			in.state.fill = devColor{0, 0, 0}
		}
	case "CS":
		if op := in.rawName(0); op != "" {
			in.state.strokeSpace = op
			in.state.stroke = devColor{0, 0, 0}
		}
	case "sc", "scn":
		in.state.fill = in.spaceColor(false)
	case "SC", "SCN":
		in.state.stroke = in.spaceColor(true)

	// Paths ------------------------------------------------------------------
	case "m":
		in.moveTo(in.num(0, 0), in.num(1, 0))
	case "l":
		in.lineTo(in.num(0, 0), in.num(1, 0))
	case "c":
		in.curveTo(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0), in.num(4, 0), in.num(5, 0))
	case "v":
		in.curveTo(in.lastX(), in.lastY(), in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0))
	case "y":
		in.curveTo(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0), in.num(2, 0), in.num(3, 0))
	case "h":
		if in.hasPath {
			in.current.closed = true
			in.closePath()
		}
	case "re":
		in.rectangle(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0))
	case "S":
		in.strokePath()
	case "s":
		if in.hasPath {
			in.current.closed = true
			in.closePath()
		}
		in.strokePath()
	case "f", "F":
		in.fillPath(false)
	case "f*":
		in.fillPath(true)
	case "B":
		in.fillPath(false)
		in.strokePath()
	case "B*":
		in.fillPath(true)
		in.strokePath()
	case "b":
		if in.hasPath {
			in.current.closed = true
			in.closePath()
		}
		in.fillPath(false)
		in.strokePath()
	case "b*":
		if in.hasPath {
			in.current.closed = true
			in.closePath()
		}
		in.fillPath(true)
		in.strokePath()
	case "n":
		in.path = nil
		in.hasPath = false
	case "W", "W*":
		in.clipPath()

	// Text -------------------------------------------------------------------
	case "BT":
		in.inText = true
		in.text = textState{m: identity, lineM: identity}
	case "ET":
		in.inText = false
	case "Tf":
		in.setFont(in.rawName(0), in.num(1, 12))
	case "Tc":
		in.state.charSpace = in.num(0, 0)
	case "Tw":
		in.state.wordSpace = in.num(0, 0)
	case "Tz":
		in.state.hscale = in.num(0, 100) / 100
	case "TL":
		in.state.leading = in.num(0, 0)
	case "Ts":
		in.state.rise = in.num(0, 0)
	case "Tr":
		in.state.render = int(in.num(0, 0))
	case "Td":
		in.moveText(in.num(0, 0), in.num(1, 0))
	case "TD":
		in.state.leading = -in.num(1, 0)
		in.moveText(in.num(0, 0), in.num(1, 0))
	case "Tm":
		in.text.m = matrix{in.num(0, 1), in.num(1, 0), in.num(2, 0), in.num(3, 1), in.num(4, 0), in.num(5, 0)}
		in.text.lineM = in.text.m
	case "T*":
		in.moveText(0, -in.state.leading)
	case "Tj":
		if s, ok := in.operand(0); ok && s.kind == pdfString {
			in.showText(s.str)
		}
	case "'":
		in.moveText(0, -in.state.leading)
		if s, ok := in.operand(0); ok && s.kind == pdfString {
			in.showText(s.str)
		}
	case "\"":
		in.state.wordSpace = in.num(0, 0)
		in.state.charSpace = in.num(1, 0)
		in.moveText(0, -in.state.leading)
		if s, ok := in.operand(2); ok && s.kind == pdfString {
			in.showText(s.str)
		}
	case "TJ":
		if v, ok := in.operand(0); ok && v.kind == pdfArray {
			in.showTextArray(v.array)
		}

	// XObjects ---------------------------------------------------------------
	case "Do":
		in.drawXObject(in.rawName(0))

	// Marked content, shading and anything else: ignored, but their operands
	// are already popped by the caller.
	case "BI":
		in.skipInlineImage()
	}
}

// applyExtGState reads the alpha and soft-mask numbers of /ExtGState.
func (in *pageInterpreter) applyExtGState(name string) {
	if name == "" {
		return
	}
	gs := in.file.resolve(mustGet(mustGet(in.resources, "ExtGState"), name))
	if gs.kind != pdfDict {
		return
	}
	if v, ok := gs.get("ca"); ok && v.kind == pdfNumber {
		in.state.alpha = clamp01(v.num)
	}
	if v, ok := gs.get("CA"); ok && v.kind == pdfNumber {
		in.state.strokeAlpha = clamp01(v.num)
	}
	if v, ok := gs.get("LW"); ok && v.kind == pdfNumber {
		in.state.lineWidth = v.num
	}
}

// spaceColor reads the last operands as a color in the current space.
func (in *pageInterpreter) spaceColor(stroke bool) devColor {
	space := in.state.fillSpace
	if stroke {
		space = in.state.strokeSpace
	}
	switch space {
	case "DeviceGray", "G", "CalGray":
		return gray(in.num(0, 0))
	case "DeviceRGB", "RGB", "CalRGB":
		return rgb(in.num(0, 0), in.num(1, 0), in.num(2, 0))
	case "DeviceCMYK", "CMYK":
		return cmyk(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0))
	}
	// A named color space: read the resource for the number of components.
	cs := in.file.resolve(mustGet(mustGet(in.resources, "ColorSpace"), space))
	if cs.kind == pdfArray && len(cs.array) > 0 {
		switch cs.array[0].name {
		case "ICCBased":
			if d := in.file.resolve(cs.array[1]); d.kind == pdfDict {
				switch mustInt(d, "N") {
				case 4:
					return cmyk(in.num(0, 0), in.num(1, 0), in.num(2, 0), in.num(3, 0))
				case 1:
					return gray(in.num(0, 0))
				}
			}
		case "Separation", "DeviceN":
			// A tint of one ink: take it as gray, which is what a reader does
			// when the alternate space is not worth resolving.
			return gray(1 - in.num(0, 0))
		}
	}
	return rgb(in.num(0, 0), in.num(1, 0), in.num(2, 0))
}

// skipInlineImage consumes "ID … EI", which binary data hides in.
func (in *pageInterpreter) skipInlineImage() {
	// The caller's operand list only holds the BI…ID dictionary; the payload
	// runs to the next EI that is followed by whitespace or a delimiter.
	// Without the stream here, the parser above has already stepped past it
	// byte by byte, so there is nothing more to do.
}

func (in *pageInterpreter) setFont(name string, size float64) {
	in.state.fontSize = size
	font := in.file.resolve(mustGet(mustGet(in.resources, "Font"), name))
	if font.kind != pdfDict {
		in.font = nil
		return
	}
	in.font = in.file.loadFont(font)
}

// device returns the CTM applied to a user-space point.
func (in *pageInterpreter) device(x, y float64) (float64, float64) {
	dx, dy := in.state.ctm.apply(x, y)
	return dx * in.scale, float64(in.canvas.h) - dy*in.scale
}

func (in *pageInterpreter) moveTo(x, y float64) {
	in.finish()
	in.current = subpath{}
	in.current.pts = append(in.current.pts, point{x, y})
	in.hasPath = true
}

func (in *pageInterpreter) lineTo(x, y float64) {
	if !in.hasPath {
		in.moveTo(x, y)
		return
	}
	in.current.pts = append(in.current.pts, point{x, y})
}

func (in *pageInterpreter) curveTo(x1, y1, x2, y2, x3, y3 float64) {
	if !in.hasPath {
		in.moveTo(x1, y1)
	}
	x0, y0 := in.lastX(), in.lastY()
	// Flatten the cubic into short straight segments: at reading sizes the
	// difference is invisible, and every downstream routine is line-based.
	steps := 8
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		mt := 1 - t
		x := mt*mt*mt*x0 + 3*mt*mt*t*x1 + 3*mt*t*t*x2 + t*t*t*x3
		y := mt*mt*mt*y0 + 3*mt*mt*t*y1 + 3*mt*t*t*y2 + t*t*t*y3
		in.current.pts = append(in.current.pts, point{x, y})
	}
}

func (in *pageInterpreter) lastX() float64 {
	if n := len(in.current.pts); n > 0 {
		return in.current.pts[n-1].x
	}
	return 0
}

func (in *pageInterpreter) lastY() float64 {
	if n := len(in.current.pts); n > 0 {
		return in.current.pts[n-1].y
	}
	return 0
}

func (in *pageInterpreter) closePath() {
	if len(in.current.pts) > 0 {
		in.current.pts = append(in.current.pts, in.current.pts[0])
	}
}

func (in *pageInterpreter) finish() {
	if len(in.current.pts) > 1 {
		in.path = append(in.path, in.current)
	}
	in.current = subpath{}
}

func (in *pageInterpreter) rectangle(x, y, w, h float64) {
	in.finish()
	pts := []point{{x, y}, {x + w, y}, {x + w, y + h}, {x, y + h}, {x, y}}
	in.path = append(in.path, subpath{pts: pts, closed: true})
	in.hasPath = true
}

// moveText shifts the text matrix along its own line.
func (in *pageInterpreter) moveText(tx, ty float64) {
	m := matrix{1, 0, 0, 1, tx, ty}
	in.text.lineM = m.mul(in.text.lineM)
	in.text.m = in.text.lineM
}

// showText advances the text matrix one glyph at a time and draws each glyph's
// run, so kerning arrays and per-glyph spacing stay correct.
func (in *pageInterpreter) showText(b []byte) {
	if in.font == nil {
		return
	}
	size := in.state.fontSize
	if size == 0 {
		return
	}
	trm := matrix{size * in.state.hscale, 0, 0, size, 0, in.state.rise}
	trm = trm.mul(in.text.m)
	startX, startY := trm.apply(0, 0)

	for _, g := range in.font.decode(b) {
		width := in.font.width(g.code)
		advance := (width*size + in.state.charSpace) * in.state.hscale
		if g.code == 0x20 && in.state.wordSpace != 0 {
			advance += in.state.wordSpace * in.state.hscale
		}
		in.text.m = matrix{1, 0, 0, 1, advance / size, 0}.mul(in.text.m)
	}

	endX, endY := in.text.m.mul(trm).apply(0, 0)
	_ = endX
	_ = endY

	if in.collect {
		in.recordRun(b, startX, startY, size)
		return
	}

	// Draw the run as one baseline so that antialiasing stays continuous.
	rotate := math.Atan2(trm.b, trm.a)
	in.drawRun(b, startX, startY, size, rotate)
}

func (in *pageInterpreter) showTextArray(items []pdfValue) {
	if in.font == nil {
		return
	}
	size := in.state.fontSize
	for _, it := range items {
		switch it.kind {
		case pdfString:
			in.showText(it.str)
		case pdfNumber:
			// A kerning adjustment, in thousandths of the em.
			shift := -it.num / 1000 * size * in.state.hscale
			in.text.m = matrix{1, 0, 0, 1, shift / size, 0}.mul(in.text.m)
		}
	}
}

// drawRun paints one text run. Without the font program's outlines the run is
// drawn as a solid bar of the glyphs' exact advance, which keeps the page's
// layout, headings and emphasis legible; browsers do the same for a missing
// font, and it is honest about what the page said rather than inventing
// letterforms.
func (in *pageInterpreter) drawRun(b []byte, x, y, size, _ float64) {
	if in.font == nil || len(b) == 0 {
		return
	}
	gs := in.state
	col := gs.fill
	if gs.render == 1 || gs.render == 5 {
		col = gs.stroke
	}
	var advance float64
	for _, g := range in.font.decode(b) {
		advance += in.font.width(g.code)*size + gs.charSpace
		if g.code == 0x20 {
			advance += gs.wordSpace
		}
	}
	advance *= gs.hscale
	if advance <= 0 {
		return
	}
	// Height: the cap height of a typical face, 0.7 of the em, which leaves
	// the leading between lines visible instead of merging them.
	h := size * 0.70
	w := computeRunWidth(in.font, b, size, gs.charSpace, gs.wordSpace, gs.hscale)
	if w <= 0 {
		w = advance
	}

	// Glyph pixels per text-space unit.
	ppu := math.Hypot(gs.ctm.a, gs.ctm.b) * in.scale
	px := x * in.scale
	py := float64(in.canvas.h) - y*in.scale
	rw, rh := w*ppu, h*ppu
	if rw < 0.4 || rh < 0.4 {
		return
	}
	x0, y0 := int(math.Floor(px)), int(math.Floor(py-rh))
	x1, y1 := int(math.Ceil(px+rw)), int(math.Ceil(py))
	alpha := gs.alpha
	if gs.render == 1 || gs.render == 5 {
		alpha = gs.strokeAlpha
	}
	r, g, bl := byte(col.r*255), byte(col.g*255), byte(col.b*255)
	for yy := y0; yy <= y1; yy++ {
		if yy < 0 || yy >= in.canvas.h {
			continue
		}
		for xx := x0; xx <= x1; xx++ {
			in.canvas.blend(xx, yy, r, g, bl, alpha)
		}
	}
}

// computeRunWidth sums the glyph advances of a run, in user space.
func computeRunWidth(f *pdfFont, b []byte, size, charSpace, wordSpace, hscale float64) float64 {
	var w float64
	for _, g := range f.decode(b) {
		w += f.width(g.code)*size + charSpace
		if g.code == 0x20 {
			w += wordSpace
		}
	}
	return w * hscale
}

// drawXObject paints a form or an image.
func (in *pageInterpreter) drawXObject(name string) {
	if name == "" {
		return
	}
	ref, ok := mustGet(in.resources, "XObject").dict[name]
	if !ok {
		return
	}
	xo := in.file.resolve(ref)
	if xo.kind != pdfDict {
		return
	}
	subtype := mustGet(xo, "Subtype")
	switch {
	case subtype.is("Form"):
		if ref.kind == pdfRef && in.forms[ref.int()] {
			return
		}
		if ref.kind == pdfRef {
			in.forms[ref.int()] = true
			defer delete(in.forms, ref.int())
		}
		data, err := in.file.decodeStream(xo)
		if err != nil {
			return
		}
		saved := in.state
		savedRes := in.resources
		if m, ok := mustGet(xo, "Matrix").numList(6); ok {
			in.state.ctm = matrix{m[0], m[1], m[2], m[3], m[4], m[5]}.mul(in.state.ctm)
		}
		if res, ok := xo.get("Resources"); ok {
			in.resources = in.file.resolve(res)
		}
		in.run(data, in.depth+1)
		in.state = saved
		in.resources = savedRes
	case subtype.is("Image"):
		in.drawImage(xo)
	}
}

// drawImage places an image's pixels in the unit square, mapped by the CTM.
func (in *pageInterpreter) drawImage(xo pdfValue) {
	w, h := mustInt(xo, "Width"), mustInt(xo, "Height")
	if w <= 0 || h <= 0 || w*h > 40_000_000 {
		return
	}
	raw, err := in.file.decodeStream(xo)
	if err != nil || len(raw) == 0 {
		return
	}
	img, err := decodeImageData(xo, raw, w, h)
	if err != nil || img == nil {
		return
	}
	// Map the image's unit square through the CTM to device pixels.
	corners := [4][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	var minX, minY = math.Inf(1), math.Inf(1)
	var maxX, maxY = math.Inf(-1), math.Inf(-1)
	for _, c := range corners {
		x, y := in.device(c[0], c[1])
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		minY, maxY = math.Min(minY, y), math.Max(maxY, y)
	}
	scaleX, scaleY := (maxX-minX)/float64(w), (maxY-minY)/float64(h)
	if scaleX <= 0 || scaleY <= 0 {
		return
	}
	bounds := image.Rect(0, 0, w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			dx := int(minX + float64(x)*scaleX)
			dy := int(minY + float64(y)*scaleY)
			r, g, b, _ := c.RGBA()
			for py := dy; py < dy+int(math.Ceil(scaleY)); py++ {
				for px := dx; px < dx+int(math.Ceil(scaleX)); px++ {
					in.canvas.blend(px, py, byte(r>>8), byte(g>>8), byte(b>>8), in.state.alpha)
				}
			}
		}
	}
}

// decodeImageData turns a stream into an image, handling the codecs a PDF
// actually ships: DCT (JPEG), JPX (JPEG 2000), CCITT, and raw samples.
func decodeImageData(dict pdfValue, raw []byte, w, h int) (image.Image, error) {
	filters := []string{}
	if v, ok := dict.get("Filter"); ok {
		if v.kind == pdfName {
			filters = append(filters, v.name)
		} else if v.kind == pdfArray {
			for _, e := range v.array {
				if e.kind == pdfName {
					filters = append(filters, e.name)
				}
			}
		}
	}
	for _, f := range filters {
		switch f {
		case "DCTDecode":
			return decodeJPEG(raw)
		case "JPXDecode":
			return decodeJPX(raw)
		case "CCITTFaxDecode":
			if img := decodeCCITT(dict, raw, w, h); img != nil {
				return img, nil
			}
			return nil, nil
		case "FlateDecode", "LZWDecode", "RunLengthDecode", "ASCIIHexDecode", "ASCII85Decode":
			// Already decoded by the file reader.
		}
	}
	return decodeRawSamples(dict, raw, w, h)
}

// decodeRawSamples reads DeviceGray, DeviceRGB and DeviceCMYK 8-bit samples.
func decodeRawSamples(dict pdfValue, raw []byte, w, h int) (image.Image, error) {
	space := mustGet(dict, "ColorSpace")
	components := 3
	switch {
	case space.is("DeviceGray"):
		components = 1
	case space.is("DeviceCMYK"):
		components = 4
	case space.kind == pdfArray && len(space.array) > 0:
		switch space.array[0].name {
		case "ICCBased", "CalRGB", "DeviceRGB":
			components = 3
		case "CalGray":
			components = 1
		}
	}
	bpc := mustInt(dict, "BitsPerComponent")
	if bpc == 0 {
		bpc = 8
	}
	if bpc != 8 || len(raw) < w*h*components {
		// Only 8-bit samples are shown; anything else decodes through the
		// codecs above in practice.
		return nil, nil
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w*h; i++ {
		var r, g, b byte
		switch components {
		case 1:
			r, g, b = raw[i], raw[i], raw[i]
		case 3:
			r, g, b = raw[i*3], raw[i*3+1], raw[i*3+2]
		case 4:
			c, m, y, k := raw[i*4], raw[i*4+1], raw[i*4+2], raw[i*4+3]
			r = byte(255 - min(255, int(c)+int(k)))
			g = byte(255 - min(255, int(m)+int(k)))
			b = byte(255 - min(255, int(y)+int(k)))
		}
		img.Set(i%w, i/w, color.RGBA{r, g, b, 255})
	}
	return img, nil
}

// gray, rgb and cmyk build device colors from the numbers an operator gave.
func gray(v float64) devColor {
	v = clamp01(v)
	return devColor{v, v, v}
}

func rgb(r, g, b float64) devColor {
	return devColor{clamp01(r), clamp01(g), clamp01(b)}
}

// cmyk converts with the naive subtraction, which is what a reader shows when
// no output intent is attached.
func cmyk(c, m, y, k float64) devColor {
	return devColor{(1 - c) * (1 - k), (1 - m) * (1 - k), (1 - y) * (1 - k)}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// bboxHelpers keeps strconv imported where debugging needs it.
var _ = strconv.Itoa
var _ = bytes.TrimSpace
var _ draw.Drawer







