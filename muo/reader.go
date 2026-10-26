package muo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"example.com/muo/muo/code"
	"example.com/muo/muo/pdf"
)

// Reader is the whole application: the open documents, the display calibration
// and the type model derived from it. Its view is a pure function of these
// fields, rebuilt every frame.
type Reader struct {
	window *mygo.Window

	documents []*Document
	active    int
	doc       *Document // the active one, for readability in the view

	manual float32 // the user's own type size, DIPs; 0 follows the model

	// Calibration, by monitor. The key changes when the window moves to a
	// screen of another size, so the 27" and the 16" each keep their numbers.
	displays map[string]*Display
	display  *Display

	typ      Typography
	settings bool
	finding  bool
	query    string
	match    int // which match the find bar shows

	// What the find bar's count was made from, so that it is counted again
	// only when the query or the document changed.
	matchList   []int
	matchQuery  string
	matchSource string
	shownMatch  int // the match the page was last scrolled to, -1 for none

	// Find's landing places: one anchor per block the page draws, and whether the
	// page still owes a scroll to the current match.
	anchors       []anchor
	pendingScroll bool
	// pdfFound is a PDF's matches with their places on their pages, since a PDF
	// has no source offsets to search.
	pdfFound []pdfMatch
	// pageScale is the ratio the PDF page is drawn at, which turns a position in
	// points into a scroll offset.
	pageScale float32
	// contentScrollAt and pdfScrollAt are where Find last put each page, so that
	// choosing a match the page has already moved away from scrolls again.
	contentScrollAt float32

	cfg settingsState

	// Rendering. file is the parsed PDF, raster the page on screen.
	file    *pdf.Document
	viewing int
	raster  *pdfPageRaster

	md markdown.Options

	// The page's place: the scroll container's state, which also saves the
	// reader's position, and the outline built from the document's headings.
	// Handles, not elements, carry a heading's rectangle across frames.
	contentScroll  ui.ScrollState
	column         ui.Handle
	headingHandles map[int]ui.Handle
	// showLibrary is whether the left pane — the list of open files — is on
	// screen. showOutline is the right pane, which is the document's own
	// headings. They are separate because they answer different questions:
	// which file, and where in this file.
	showLibrary bool
	showOutline bool
	// zen folds both panes away and hides the status pill, leaving the page and
	// the header. It is a reading mode, not a pane state: coming out of it
	// restores the panes the reader had, so it never costs them a decision.
	zen bool
	// narrowLibrary and narrowOutline are the reader's explicit request for a
	// pane at a width where it would otherwise yield to the page.
	narrowLibrary bool
	narrowOutline bool
	outline       ui.OutlineState[string]

	// Where to put the reader when a document opens, until the heading it
	// stopped at has been measured.
	restorePending bool
	restoreTitle   string

	// sidebarChoice is the id of the row the library shows as chosen, and
	// outlineChoice the outline's. They are separate lists, so each keeps its own
	// choice: one id chosen in the library must not mark a heading in the outline.
	sidebarChoice string
	outlineChoice string
	// outlineBounds is where the outline's list was laid out last frame. It is
	// kept for the tests that click its rows by position, since a heading's
	// words appear twice on screen and a click by text finds the page's.
	outlineBounds ui.Rect
	// outlineRows are the outline's row boxes, one per heading, for the same
	// reason.
	outlineRows []ui.Rect
	// callouts maps a block index to the callout kind it was, for the quotes the
	// page drew as callouts. It is rebuilt every frame with the page.
	callouts map[int]calloutKind
	// outlineX is the outline pane's left edge this frame, which slides as it
	// folds. The tests read it to check the pane glides rather than vanishing.
	outlineX float32

	// sidebarWidth is the shell's source-list column, which the content view
	// cannot see from its own context: the shell sizes that context to the
	// window.
	sidebarWidth float32

	// sidebarW is the width the reader has given the library, which the divider
	// between the panes sets. It is kept across launches, so a reader who makes
	// room for long file names keeps that room.
	sidebarW float32

	// outlineW is the width of the right pane, the document's outline.
	outlineW float32

	// libReveal and outReveal are how much of each panel is on screen this frame,
	// from 0 to 1, animated by the row that holds them. They are kept here because
	// the measure is solved from the page's width before the row is built.
	libReveal float32
	outReveal float32

	// page is the content's own box, measured to learn the real page width.
	page ui.Handle

	// chosenHeading is true after the reader picks an entry, so that the choice
	// is what gets saved rather than whatever sits at the top of the page.
	chosenHeading bool
	jumpToBlock       int // the heading an outline entry asked for, -1 for none

	mu   sync.Mutex
	work int // frames are single-threaded; kept for symmetry with goroutines
}

// settingsState is presentation state the settings surface owns.
type settingsState struct {

}

// New returns a reader with no document. The appearance is the desktop's, not
// the reader's: a page that ignores the system's light and dark choice reads as
// broken (dark-mode.md › Best practices).
func New() *Reader {
	// The window comes back as the reader left it: the panes' widths, and which of
	// them were open. A layout saved before the panes' visibility was kept has no
	// answer for it, so those default to open.
	layout := loadLayout()
	library, outline := true, true
	if layout.Library != nil {
		library = *layout.Library
	}
	if layout.Outline != nil {
		outline = *layout.Outline
	}
	return &Reader{
		displays:       map[string]*Display{},
		active:         -1,
		viewing:        -1,
		headingHandles: map[int]ui.Handle{},
		showLibrary:    library,
		showOutline:    outline,
		jumpToBlock:   -1,
		shownMatch:    -1,
		sidebarW:      clampPanel(layout.SidebarWidth),
		outlineW:      clampPanel(layout.OutlineWidth),
	}
}

// Attach gives the reader its window, which it needs for async work and for
// the display it is on.
func (r *Reader) Attach(w *mygo.Window) { r.window = w }

// ---------------------------------------------------------------------------
// Opening files

// OpenIn opens a file as a document, or shows it if it is already open.
func (r *Reader) OpenIn(path string) {
	if path == "" {
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return
	}
	for i, d := range r.documents {
		if sameFile(d.Path, abs) {
			r.active = i
			r.documentChanged()
			return
		}
	}
	doc := &Document{Path: abs, Title: titleOf(abs), Format: classify(abs)}
	if data, err := os.ReadFile(abs); err != nil {
		doc.Err = err.Error()
	} else if doc.Format == FormatPDF {
		if _, ok := pdf.Load(abs); !ok {
			doc.Err = "This PDF couldn't be read. It may be damaged, encrypted, or use a format the reader doesn't support."
		}
	} else {
		// Tabs become spaces before anything reads the source: the text engine
		// draws a tab at zero width, so a tab-indented Go file would lose every
		// block in it. See tabs.go.
		doc.Source = code.ExpandTabs(plainText(string(data)))
		doc.headings = outlineOf(doc.Source)
	}
	r.documents = append(r.documents, doc)
	r.active = len(r.documents) - 1
	// The file goes into the recent list, which the empty page and the Open
	// Recent menu both read. Only once a window is up: recording one reaches
	// the app's own menu, which a headless test has no thread to answer on, and
	// a failure there is not worth interrupting a reader for in any case.
	if r.window != nil {
		go func() { _ = mygo.App.AddRecentDocument(abs) }()
	}
	touchRecent(abs)
	r.documentChanged()
}

// Open asks for a file.
func (r *Reader) Open() {
	files, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Title: "Open a File",
		Filters: []mygo.FileFilter{
			{Name: "Documents", Extensions: []string{"md", "markdown", "pdf", "txt", "text"}},
			{Name: "Markdown", Extensions: []string{"md", "markdown"}},
			{Name: "PDF", Extensions: []string{"pdf"}},
			{Name: "All Files", Extensions: []string{"*"}},
		},
	})
	if err != nil || len(files) == 0 {
		return
	}
	for _, f := range files {
		r.OpenIn(f)
	}
}

// Reload re-reads the active document, for a file that changed on disk.
func (r *Reader) Reload() {
	d := r.doc
	if d == nil {
		return
	}
	pdf.Forget(d.Path)
	data, err := os.ReadFile(d.Path)
	if err != nil {
		d.Err = err.Error()
		return
	}
	d.Err = ""
	if d.Format == FormatPDF {
		r.file = nil
		r.raster = nil
		r.loadDoc()
		return
	}
	d.Source = code.ExpandTabs(plainText(string(data)))
}

func (r *Reader) closeDocument(i int) {
	if i < 0 || i >= len(r.documents) {
		return
	}
	r.documents = append(r.documents[:i], r.documents[i+1:]...)
	if r.active >= len(r.documents) {
		r.active = len(r.documents) - 1
	}
	r.documentChanged()
}

func (r *Reader) selectDocument(i int) {
	if i < 0 || i >= len(r.documents) {
		return
	}
	r.active = i
	r.documentChanged()
}

// documentChanged points the view at the new active document, prepares its
// renderer, and puts the reader back where they stopped in it.
func (r *Reader) documentChanged() {
	if r.active >= 0 && r.active < len(r.documents) {
		r.doc = r.documents[r.active]
	} else {
		r.doc = nil
	}
	r.raster = nil
	r.file = nil
	r.viewing = 0
	r.headingHandles = map[int]ui.Handle{}
	// The new document has its own matches, and the bar's count must not carry
	// the last document's.
	r.matchList, r.matchQuery, r.matchSource = nil, "", ""
	r.shownMatch = -1
	if r.doc != nil {
		r.loadDoc()
		r.restorePosition()
	}
}

// restorePosition puts the reader back where they left the document: at the
// heading they stopped at, or failing that at the same distance down the page.
// The heading's offset is not known until the page has been laid out once, so
// the stored offset is used first and the heading is found on the next frame.
func (r *Reader) restorePosition() {
	doc := r.doc
	s, ok := savedPosition(doc.Path)
	if !ok {
		return
	}
	doc.heading = s.Heading
	r.contentScroll.Y = s.Position
	r.restoreTitle = s.Title
	r.restorePending = true
}

func (r *Reader) loadDoc() {
	if r.doc == nil {
		return
	}
	switch r.doc.Format {
	case FormatPDF:
		if f, ok := pdf.Load(r.doc.Path); ok {
			r.file = f
			if r.viewing >= f.PageCount() {
				r.viewing = 0
			}
		}
	case FormatMarkdown:
		// Images resolve next to the document, as in any Markdown editor.
		r.md.LoadImage = markdown.FileImages(filepath.Dir(r.doc.Path))
		r.md.OnLink = func(url string) {
			if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "mailto:") {
				mygo.Shell.OpenExternal(url)
			}
		}
	}
}

func sameFile(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	return err1 == nil && err2 == nil && strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}

// ---------------------------------------------------------------------------
// Displays

// displayKey identifies the monitor under the window, so that calibration
// follows the screen rather than the document.
func displayKey(w *mygo.Window) string {
	if w == nil {
		return "windowless"
	}
	d := mygo.Screen.DisplayMatching(mygo.Rectangle(w.Bounds()))
	return fmt.Sprintf("%dx%d@%d,%d", int(d.WorkArea.Width), int(d.WorkArea.Height), int(d.WorkArea.X), int(d.WorkArea.Y))
}

// currentDisplay returns the calibration for the window's monitor, creating a
// sensible default the first time that monitor is seen.
func (r *Reader) currentDisplay(w *mygo.Window) *Display {
	key := displayKey(w)
	if d, ok := r.displays[key]; ok {
		r.display = d
		return d
	}
	d := &Display{Key: key, Label: labelFor(key), WidthCm: 60, Distance: defaultDistance(key)}
	r.displays[key] = d
	r.display = d
	return d
}

// labelFor names a monitor the way a person would, from its size class: a
// 2560-wide work area is a 27", a 1512-wide one a 16" laptop.
func labelFor(key string) string {
	var w int
	fmt.Sscanf(key, "%dx", &w)
	switch {
	case w <= 0:
		return "This Display"
	case w < 1200:
		return "13-inch Laptop"
	case w < 1600:
		return "14-inch Laptop"
	case w < 1900:
		return "16-inch Laptop"
	case w < 2400:
		return "24-inch Display"
	case w < 3400:
		return "27-inch Display"
	case w < 4000:
		return "32-inch Display"
	}
	return "Large Display"
}

// defaultDistance guesses from the size class, which is a better starting point
// than one number for every screen.
func defaultDistance(key string) float64 {
	var w int
	fmt.Sscanf(key, "%dx", &w)
	switch {
	case w > 0 && w < 1900:
		return 42
	case w < 2600:
		return 55
	}
	return 65
}

// ---------------------------------------------------------------------------
// Type model

// recompute derives the type from the display and the window. It runs every
// frame, which is cheap: it measures one character and one cap height.
func (r *Reader) recompute(c *ui.Context) {
	w := r.window
	d := r.currentDisplay(w)
	manual := r.manual
	if d.FontPx > 0 {
		manual = d.FontPx
	}
	cw, _ := c.Size()
	// The measure is computed from the page's width, not the window's: the shell
	// sizes its content context to the whole window, so the panes must be taken
	// off before the column is solved — and only the ones actually on screen, or a
	// hidden pane would keep costing the page its width. Each takes its width plus
	// the divider's own line.
	//
	// It is solved from the panels' destinations, not from their animated reveal,
	// so the column is one width for the whole of a slide. Feeding it the reveal
	// re-solved the measure every frame, which re-broke every paragraph every
	// frame: a document's worth of text layout between each pair of frames, which
	// is a stutter on a 60 Hz display and unusable on a 240 Hz one.
	//
	// The pane counted is the one that will be shown, which is not the same as the
	// reader's preference: a pane that yields at a narrow width must not be
	// subtracted, or the page would be solved for a width it does not have.
	pageW := float64(cw)
	if r.libraryShown(c) {
		pageW -= float64(leftPanelWidth)
	}
	if r.outlineShown(c) {
		pageW -= float64(rightPanelWidth)
	}
	if pageW < 240 {
		pageW = float64(cw)
	}
	// The column is solved in the family the page is drawn in, so the measure
	// the readout reports is the one the reader is actually looking at.
	r.typ = Model(*d, pageW, r.ppiFor(w), pageFamily, manual)
}

// ppiFor estimates the density the window is drawn at. MyGo reports DIPs, so
// the window's own numbers already carry the scale; the physical density is the
// display's, which the window's DIP size reveals: a laptop panel drawn at 2×
// reports about half the DIPs of its pixels, a desktop monitor at 1× reports
// them all.
func (r *Reader) ppiFor(w *mygo.Window) float64 { return 96 * displayScale(w) }

func displayScale(w *mygo.Window) float64 {
	if w == nil {
		return 1
	}
	d := mygo.Screen.DisplayMatching(mygo.Rectangle(w.Bounds()))
	switch {
	case d.WorkArea.Width > 2200:
		return 1
	case d.WorkArea.Width > 1400:
		return 1.5
	}
	return 2
}






