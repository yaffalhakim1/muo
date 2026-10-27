package muo

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown/parser"
)

// Finding in the page.
//
// The renderer draws the document and does not say where a word ended up on
// screen. So the app puts an anchor on the block that holds each match: every
// block the reader can see gets an element, that element is bound to a handle,
// and ui.Element.ScrollIntoView scrolls it into view in the frame it is asked
// for. That is exact — the layout engine knows where the block is, and the app
// does not have to estimate it.
//
// The first version of this interpolated a match's position from the headings
// around it. It was wrong in the case that matters: a match below the last
// heading, or in a document with no headings at all, landed at a guess, and a
// match outside the viewport could scroll nowhere. An anchor cannot be wrong,
// because it is the block itself.
//
// Everything else in Find is counting and marking: the bar reports how many
// matches there are and which one is current, and the page paints them.

// anchor is one block the reader can land on: the range of the source it covers,
// and the handle of the marker built just above it.
type anchor struct {
	start, end int
	handle     ui.Handle
}

// findMatches returns the source offsets of every occurrence of query, in
// order, ignoring case. An empty query matches nothing.
func findMatches(src, query string) []int {
	if query == "" {
		return nil
	}
	lower := strings.ToLower(src)
	q := strings.ToLower(query)
	var out []int
	for from := 0; from+len(q) <= len(lower); {
		i := strings.Index(lower[from:], q)
		if i < 0 {
			break
		}
		out = append(out, from+i)
		from += i + len(q)
	}
	return out
}

// pdfMatch is one occurrence of the query in a PDF: the page it is on, and where
// on that page.
type pdfMatch struct {
	page int
	// Y is the run's baseline in points from the page's top, which is what the
	// view turns into a scroll offset.
	Y float64
}

// pdfMatches returns every occurrence of the query in a PDF, page by page.
//
// A PDF has no source text to search, so the interpreter is asked to read the
// pages: it records the runs it would draw, each with its baseline, and the
// query is matched against their words. That is why a PDF's Find is a page-level
// search with a position rather than a character offset — there is no character
// offset to have, only a place on a page.
func (r *Reader) pdfMatches() []pdfMatch {
	if r.file == nil || r.query == "" {
		return nil
	}
	q := strings.ToLower(r.query)
	var out []pdfMatch
	for page := 0; page < r.file.PageCount(); page++ {
		runs, err := r.file.PageText(page)
		if err != nil {
			continue
		}
		for _, run := range runs {
			if strings.Contains(strings.ToLower(run.Text), q) {
				out = append(out, pdfMatch{page: page, Y: run.Y})
			}
		}
	}
	return out
}

// findBar is the bar itself, under the title row. The page has already made
// room for it, so it sits in the chrome rather than over the text.
func (r *Reader) findBar(c *ui.Context) {
	if !r.finding || r.doc == nil {
		return
	}
	matches := r.matches()
	// The bar counts what the page can show. A PDF marks every match it finds,
	// so it counts them all.
	count := len(matches)
	if r.doc.Format != FormatPDF {
		count = r.paintableMatches()
	}
	if count > 0 {
		if r.match < 0 || r.match >= count {
			r.match = 0
		}
	} else {
		r.match = 0
	}
	ui.Row(c).Grow(1).MaxWidth(560).Children(func() {
		bar := ui.FindBar(c, &r.finding, &r.query, count, &r.match)
		// The bar is the app's, not the toolkit's default: it sits on the page's
		// own paper with the page's rule, so a search is something the page does
		// rather than a control pasted over it.
		t := c.Theme()
		bar.Radius(8).Background(t.Surface).
			Border(1, t.Border.Alpha(0.8))
	})
	if !r.finding {
		return
	}
	// The bar moved the choice: the page follows it.
	//
	// The request is made whenever the chosen match changes, and also when the
	// reader has scrolled the page away from where Find left it — otherwise
	// Previous, which lands on the match the page was last scrolled to, would
	// appear to do nothing on the way back. It is not made every frame, which
	// would fight the reader's own scrolling.
	//
	// The first request is made even when the choice has not changed, which is
	// the case a query with a single match falls into: there is nowhere to move
	// to, so `match` stays 0 and an equality test on it never fires. Typing a
	// query that matches once and pressing Enter did nothing, because nothing
	// had asked the page to go anywhere.
	//
	// The bar is drawn after the page in a frame, so a request made here is served
	// by the next frame's hook, not this one.
	if count > 0 {
		if r.match != r.shownMatch || r.scrollMoved() || r.shownMatch < 0 {
			r.shownMatch = r.match
			r.pendingScroll = true
			r.askFrame()
		}
	}
}

// askFrame makes sure another frame runs, so that a scroll requested after the
// page was built is served.
func (r *Reader) askFrame() {
	if r.window != nil {
		r.window.Invalidate()
	}
}

// scrollMoved reports whether the reader has scrolled the page away from where
// Find last put it.
func (r *Reader) scrollMoved() bool {
	return r.contentScrollAt > 0 && r.contentScroll.Y != r.contentScrollAt
}
// matches is the current document's matches, recomputed only when the query or
// the source changed: the source is re-read every frame, so counting it every
// frame would be wasted work on a long page.
//
// A PDF has no source, so its matches are found a different way and the count
// comes from that. The bar shows the same number either way.
func (r *Reader) matches() []int {
	if r.doc == nil {
		return nil
	}
	if r.doc.Format == FormatPDF {
		return r.pdfMatchOffsets()
	}
	if r.matchQuery == r.query && r.matchSource == r.doc.Source {
		return r.matchList
	}
	r.matchList = findMatches(r.doc.Source, r.query)
	// Which of them the page can paint. A match inside a fence or an inline
	// span is counted but never marked, because marking it would show the tags
	// themselves. Counting those against the bar's "1 of 2" made a search look
	// broken: the reader walks to a match and finds nothing lit, because the
	// match is in code and the page cannot show it.
	skip := unmarkable(r.doc.Source)
	r.matchPaintable = make([]bool, len(r.matchList))
	for i, at := range r.matchList {
		r.matchPaintable[i] = !overlaps(skip, at, at+len(r.query))
	}
	r.matchQuery = r.query
	r.matchSource = r.doc.Source
	r.shownMatch = -1
	r.pendingScroll = false
	return r.matchList
}

// paintableMatches is how many of the current matches the page can actually
// mark. The bar counts these, because a match a reader cannot see is not a
// match they can walk to.
func (r *Reader) paintableMatches() int {
	n := 0
	for _, ok := range r.matchPaintable {
		if ok {
			n++
		}
	}
	return n
}

// matchAt is the source offset of the chosen match.
//
// The bar's choice counts only matches the page can paint, because those are the
// ones it offers. The offsets list holds every occurrence, including the ones in
// code, so the chosen index has to be translated: walking it directly would land
// on a fence every time one sits between two visible matches.
func (r *Reader) matchAt(chosen int) (int, bool) {
	if chosen < 0 || len(r.matchPaintable) == 0 {
		return 0, false
	}
	seen := 0
	for i, ok := range r.matchPaintable {
		if !ok {
			continue
		}
		if seen == chosen {
			if i >= len(r.matchList) {
				return 0, false
			}
			return r.matchList[i], true
		}
		seen++
	}
	return 0, false
}

// pdfMatchOffsets is the PDF's matches as indices into a list, so that the bar
// can count them and choose among them with the same code as a page's. The
// positions themselves live in pdfFound.
func (r *Reader) pdfMatchOffsets() []int {
	if r.matchQuery == r.query && r.matchSource == r.doc.Path && r.matchSource != "" {
		return r.matchList
	}
	r.pdfFound = r.pdfMatches()
	r.matchList = make([]int, len(r.pdfFound))
	for i := range r.pdfFound {
		r.matchList[i] = i
	}
	r.matchQuery = r.query
	r.matchSource = r.doc.Path
	r.shownMatch = -1
	r.pendingScroll = false
	return r.matchList
}

// followFindPdf turns a PDF match into a page and a scroll.
//
// A PDF's page is a picture, so a match has no element to scroll to: the page it
// is on is chosen, and the run's baseline — a distance down the page in points —
// becomes a scroll offset by the ratio the page is drawn at. That is exact
// enough to land on the line, which is what a reader wants from Find in a PDF.
// followFind asks the block that holds the current match to come into view.
//
// It runs while the page is being built, because that is the only moment the
// element exists: elements live for one frame, and ScrollIntoView asks the layout
// engine where the element it is called on actually is. That is what makes the
// landing exact — the engine knows, so the app never estimates a position, and a
// match outside the viewport lands where it truly is rather than at a guess.
func (r *Reader) followFind(row ui.Element, start, end int) {
	if !r.pendingScroll || !r.finding {
		return
	}
	at, ok := r.matchAt(r.match)
	if !ok {
		return
	}
	if at < start || at >= end {
		return
	}
	r.pendingScroll = false
	row.ScrollIntoView()	// Where the scroll is headed, remembered so that a later Previous to the same
	// match knows whether the reader has moved the page since.
	r.contentScrollAt = r.contentScroll.Y + 1
	// The request is served; a frame is needed for the layout engine to move the
	// page, and asking for one here is what makes the scroll arrive in the app
	// rather than only in a test that advances frames.
	r.askFrame()
}

func (r *Reader) followFindPdf() {
	if !r.pendingScroll || !r.finding || r.doc == nil || r.doc.Format != FormatPDF {
		return
	}
	matches := r.matches()
	if r.match < 0 || r.match >= len(matches) || r.match >= len(r.pdfFound) {
		r.pendingScroll = false
		return
	}
	m := r.pdfFound[r.match]
	r.pendingScroll = false
	if m.page != r.viewing {
		r.goToPage(m.page)
		return
	}
	if r.pageScale <= 0 {
		return
	}
	// The run's Y is in points from the page's top; the view draws the page at a
	// scale of its own, so the scroll is that Y in the page's drawn pixels. The
	// title bar floats over the page, so the target is moved down by its height
	// to keep the match clear of it.
	//
	// It arrives at once rather than gliding. A PDF's match can be on another
	// page, and a smooth scroll across a page change fights the new page's own
	// layout; a jump is also what a reader expects from turning to a result.
	r.contentScrollAt = float32(m.Y)*r.pageScale + titleBarHeight
	r.contentScroll.Y = r.contentScrollAt
}
// resetAnchors empties the anchors at the start of a frame's page build, so that
// the list is this frame's blocks and not the last one's.
func (r *Reader) resetAnchors() {
	r.anchors = r.anchors[:0]
}

// highlightedSource is the document's source with the query's matches marked, or
// the source unchanged when nothing is being found.
//
// Marking is what makes Find visible. Counting alone tells the reader a word
// exists; a mark on the page tells them where, and it is the difference between
// a search that works and one that appears to do nothing.
func (r *Reader) highlightedSource() string {
	if r.doc == nil || !r.finding || r.query == "" {
		return r.doc.Source
	}
	matches := r.matches()
	if len(matches) == 0 {
		return r.doc.Source
	}
	return markMatches(r.doc.Source, matches, len(r.query))
}

// markMatches wraps every match in <mark>, which the renderer draws with the
// theme's Mark colour.
//
// Two kinds of text are left alone, and each for a reason. A fenced or indented
// code block and an inline code span show their markup literally, so a <mark>
// inside one would appear as the tags themselves rather than as a highlight. A
// match in code is still counted and still scrolled to; it is simply not
// painted, which is better than painting it wrong.
//
// Headings are marked: this app draws them itself, and headingSpans turns the
// tags back into spans.
//
// The insertions run from the end backwards, so each match's offsets are still
// the source's own when it is used.
func markMatches(src string, matches []int, n int) string {
	if n <= 0 || len(matches) == 0 {
		return src
	}
	skip := unmarkable(src)
	out := src
	for i := len(matches) - 1; i >= 0; i-- {
		start := matches[i]
		end := start + n
		if start < 0 || end > len(src) {
			continue
		}
		if overlaps(skip, start, end) {
			continue
		}
		out = out[:start] + markOpen + out[start:end] + markClose + out[end:]
	}
	return out
}

const (
	markOpen  = "<mark>"
	markClose = "</mark>"
)

// markColor is the highlight behind a found word: the same tint the renderer
// gives a document's own <mark>, so the app does not invent a second colour for
// the same idea.
func (r *Reader) markColor() ui.Color {
	return ui.RGBA(250, 204, 21, 0.35)
}

// unmarkable returns the byte ranges where a mark would not render: code blocks
// and code spans.
func unmarkable(src string) [][2]int {
	doc := parser.Parse(src)
	var out [][2]int
	var walk func(parent int)
	walk = func(parent int) {
		for i := range doc.Children(parent) {
			b := doc.Block(i)
			switch b.Kind {
			case parser.CodeBlock, parser.HTMLBlock:
				out = append(out, [2]int{b.Start, b.End})
				continue
			}
			for _, in := range doc.Inlines(i) {
				if in.Kind == parser.Code {
					out = append(out, [2]int{in.Start, in.End})
				}
			}
			walk(i)
		}
	}
	walk(0)
	return out
}

// overlaps reports whether [start, end) meets any of the ranges.
func overlaps(ranges [][2]int, start, end int) bool {
	for _, r := range ranges {
		if start < r[1] && end > r[0] {
			return true
		}
	}
	return false
}
