package muo

import "time"

import "github.com/egoist/mygo/ui"

// The reading band, in degrees of apparent capital height: the visual angle a
// capital letter subtends. Below 0.26° the eye has to lean in; above 0.38° the
// return sweep to the next line gets unreliable and lines stop reading as a
// block. The target sits mid-band and agrees with the sizes the reading
// literature converges on: 16–18 pt on a 109 ppi desktop at 65 cm, 18–20 pt on
// a 224 ppi laptop panel at 42 cm.
const (
	targetAngle  = 0.32
	smallAngle   = 0.26
	largeAngle   = 0.38
	charsPerLine = 72.0
)

// Comfort grades the apparent character height against the reading band.
// Below Small the eye leans in; above Large the return sweep to the next line
// gets unreliable.
type Comfort int

const (
	ComfortSmall Comfort = iota
	ComfortGood
	ComfortLarge
)

func (c Comfort) String() string {
	switch c {
	case ComfortSmall:
		return "small"
	case ComfortLarge:
		return "large"
	}
	return "good"
}

// Typography is the model's output: the type size for the current monitor and
// viewing distance, and the column that keeps the measure in characters.
// Everything else in the reader derives from these two.
type Typography struct {
	FontPx        float32 // body size in DIPs
	Line          float32 // line height, a multiple of FontPx
	ColumnW       float32 // reading column width in DIPs
	Ch            float32 // advance of an average character at FontPx, DIPs
	ChPerLine     float32 // characters that fit in ColumnW
	Angle         float64 // apparent capital height, degrees
	Comfort       Comfort
	Distance      float64 // cm, from the display calibration
	EffectivePPI  float64 // pixels per inch the window is rendered at
	PreferredSize float32 // DIPs before the user's manual adjustment
}

// Display is a monitor as the reader knows it: its own width and the distance
// the user sits from it. Calibration is per display, so the 27" and the 16"
// each keep their own numbers.
type Display struct {
	Key      string  // identity of the monitor, stable across launches
	Label    string  // readable name the user can edit
	WidthCm  float64 // physical, measured once
	Distance float64 // cm, measured once
	FontPx   float32 // 0 follows the model; otherwise a manual size
}

// Format is how a document is shown.
type Format uint8

const (
	FormatMarkdown Format = iota
	FormatPDF
	FormatText
)

// Document is one open file and where the reader is in it.
type Document struct {
	Path   string
	Title  string
	Format Format
	Source string
	Scroll float32 // offset of the reading column, DIPs
	Err    string

	headings []heading // the outline, for a Markdown document
	heading  int       // the block index last visited, -1 when at the top
}

// PDFPage is one page of a PDF, rendered at a size that suits the screen.
type PDFPage struct {
	Page   int // 1-based
	Width  int // pixels
	Height int
	RGB    []byte // three bytes per pixel
}

// readingSurface returns the page colors for an appearance. Warm off-white and
// warm charcoal, not #fff and #000: the halation of maximum contrast is a
// documented fatigue source over long sessions, and a reader that never
// switches tones does not need it.
//
// The values come from the same palette the chrome is built from, so the page
// and the window around it can never drift into two different neutrals.
func readingSurface(dark bool) (bg, fg, muted, rule, selection rgbColor) {
	p := palette(dark)
	return triple(p.paper), triple(p.ink), triple(p.muted), triple(p.border), triple(p.accent)
}

// triple drops a palette colour to the plain three bytes the page mixes.
func triple(c ui.Color) rgbColor { return rgbColor{R: c.R, G: c.G, B: c.B} }

// readingLineHeight tightens measured leading a little as lines get physically
// shorter: loose leading on a short measure makes paragraphs feel floaty.
func readingLineHeight(fontPx float32) float32 {
	switch {
	case fontPx < 17:
		return 1.55
	case fontPx < 21:
		return 1.58
	case fontPx < 26:
		return 1.62
	}
	return 1.66
}

// sessionStarted keeps time out of the view while leaving the package's
// timestamping to the caller.
var sessionStarted = time.Now

// page is the content's own box, measured to learn the real page width.


// SetSidebarWidth tells the reader how much of the window the shell's source
// list takes, so the measure is computed from the page's real width rather than
// the whole window's. The shell hands its content a context sized to the
// window, so this is the one fact it cannot learn from the context itself.
func (r *Reader) SetSidebarWidth(w float32) {
	r.sidebarWidth = w
}

// SidebarWidth reports the source list's width, for sizing the window before
// the first frame.
func (r *Reader) SidebarWidth() float32 { return r.sidebarWidth }




