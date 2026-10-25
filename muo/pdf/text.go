package pdf

import (
	"math"
	"strings"
	"unicode/utf8"
)

// The text of a page, and where each run of it sits.
//
// A PDF has no lines, paragraphs or words: it has instructions to put glyphs at
// coordinates. Reading a page as text means running the same interpreter that
// paints it and keeping what it would have drawn, which is what this file does —
// the interpreter is asked to record runs instead of drawing them, so what is
// searched is exactly what is shown.
//
// The result is a flat list of runs in the order the page draws them, each with
// its baseline in points from the page's top-left corner. That is enough for
// Find: it needs to know which page a word is on and where on the page, and both
// follow from the run's own position.

// TextRun is one run of text on a page: the words, and where they are.
type TextRun struct {
	Text string
	// X, Y are the baseline's start, in points, from the page's top-left
	// corner, with Y increasing downward as the screen's does.
	X, Y float64
	// Size is the font size the run was set at, in points.
	Size float64
}

// PageText returns the text of a page, run by run, in the order it is drawn.
//
// Drawing order is what a PDF producer writes and is right for the overwhelming
// majority of documents. A page that draws its text out of reading order — a
// two-column layout built column by column is the common case — comes back in the
// order it was drawn, which is still the order its own reader shows it.
func (d *Document) PageText(index int) ([]TextRun, error) {
	if !d.HasPage(index) {
		return nil, errNoSuchPage(index)
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	in := &pageInterpreter{
		file:    d.file,
		page:    index,
		collect: true,
		state:   defaultGraphicsState(),
		forms:   map[int]bool{},
	}
	_ = in.run(d.file.Contents(index), 0)
	return in.runs, nil
}

// Text returns every page's text, for a search across the file. The pages are in
// the file's own order.
func (d *Document) Text() []string {
	out := make([]string, 0, d.PageCount())
	for i := 0; i < d.PageCount(); i++ {
		runs, err := d.PageText(i)
		if err != nil {
			out = append(out, "")
			continue
		}
		out = append(out, runsText(runs))
	}
	return out
}

// runsText joins runs into the text a reader would see, one line per baseline.
func runsText(runs []TextRun) string {
	var b strings.Builder
	lastY := math.Inf(-1)
	for _, r := range runs {
		if !math.IsInf(lastY, -1) {
			// A baseline far enough from the last one starts a new line. Half a
			// run's own size is the threshold: anything closer is a superscript
			// or a font change on the same line.
			gap := math.Abs(r.Y - lastY)
			limit := math.Max(r.Size, 1) * 0.5
			if gap > limit {
				b.WriteByte('\n')
			}
		}
		b.WriteString(r.Text)
		lastY = r.Y
	}
	return b.String()
}

// recordRun keeps a run of text for PageText. The interpreter calls it instead of
// painting when it is collecting.
func (in *pageInterpreter) recordRun(b []byte, x, y, size float64) {
	if in.font == nil || len(b) == 0 {
		return
	}
	text := in.font.text(b)
	if text == "" {
		return
	}
	// A PDF's coordinate space has its origin at the bottom-left corner with Y
	// increasing upward; the app's has it at the top-left with Y increasing
	// downward. Flipping here means every caller works in screen coordinates and
	// never has to know.
	in.runs = append(in.runs, TextRun{
		Text: text,
		X:    x,
		Y:    in.heightPoints() - y,
		Size: math.Abs(size),
	})
}

// heightPoints is the page's height in points, which the flip needs.
func (in *pageInterpreter) heightPoints() float64 {
	if in.file == nil {
		return 0
	}
	_, h := in.file.PageSize(in.page)
	return h
}

// text turns a string of glyph codes into the characters it shows, and drops what
// a reader would not see: control codes, and a glyph the font gives no character
// for.
func (f *pdfFont) text(b []byte) string {
	if f == nil {
		return ""
	}
	var sb strings.Builder
	for _, g := range f.decode(b) {
		r := g.rune()
		if r == 0 || r < 0x20 || r == utf8.RuneError {
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// rune is the character a decoded glyph shows: the font's own mapping when it has
// one, else the code read as Latin-1, which is what a simple font without a
// character map means.
func (g decodedGlyph) rune() rune {
	if g.text != "" {
		r := []rune(g.text)
		if len(r) > 0 {
			return r[0]
		}
	}
	if g.code >= 32 && g.code < 127 {
		return rune(g.code)
	}
	if g.code >= 160 && g.code < 256 {
		return rune(g.code)
	}
	return 0
}

// errNoSuchPage names the page a caller asked for and the file does not have.
func errNoSuchPage(index int) error {
	return &pageError{index: index}
}

// pageError is a request for a page outside the file.
type pageError struct{ index int }

func (e *pageError) Error() string {
	return "the file has no page " + itoa(e.index+1)
}

// itoa is strconv.Itoa without the import, so this file needs nothing beyond what
// the parser already uses.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
