package muo

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

// Headings are the page's landmarks, so the reader builds them itself instead
// of leaving them to the renderer: it has to bind each one to a handle to
// remember where the reader stopped, and scroll to it when an outline entry is
// chosen. Everything else on the page stays the renderer's.
//
// Elements live for one frame (ui/input.md), so the handle, not the element, is
// what survives.

// markdownOptions returns the renderer's options for this frame. The options
// carry the renderer's own state (a sync.Map of images), so the reader's value
// is used rather than a copy.
//
// The document the renderer is given is the marked one, so every block's Start
// and End are offsets into the *marked* source. Find's matches, though, are
// offsets into the unmarked source, because that is what the reader's file says.
// The two must be brought into one space, which markShift does: an anchor's range
// is translated back to the unmarked source before it is compared with a match.
func (r *Reader) markdownOptions(c *ui.Context, pal pagePalette) *markdown.Options {
	opts := &r.md
	opts.Theme = r.pageTheme(c, pal)
	opts.CodeBlock = r.codeBlockView
	opts.Block = func(c *ui.Context, d *parser.Document, i int) bool {
		b := d.Block(i)
		if b.Kind == parser.Heading {
			r.headingBlock(c, d, i, pal)
			return true
		}
		// A callout is a quote the renderer knows nothing about. The marker is
		// stripped from the source before the page is drawn, so by the time this
		// runs the quote is an ordinary one; the block's index is what tells the
		// hook which kind it was.
		if kind, ok := r.calloutAt(i); ok {
			r.calloutBlock(c, pal, d, i, kind)
			return true
		}
		// The block itself stays the renderer's to draw. What is added is a
		// marker: a row with no height, built just before the block in the same
		// parent, and bound to a handle so Find can scroll to it.
		//
		// The hook runs in document order while the page is built, so creating
		// an element here puts it immediately above the block it belongs to.
		row := ui.Column(c).Key(fmt.Sprint("find:", i)).FillWidth().Height(0)
		var h ui.Handle
		row.Bind(&h)
		start, end := r.markShift(b.Start, b.End)
		r.anchors = append(r.anchors, anchor{start: start, end: end, handle: h})
		r.followFind(row, start, end)
		return false
	}
	return opts
}

// markShift translates a range in the marked source back to the unmarked one.
//
// The two differ only by the tags Find inserted, so the shift is how many bytes of
// tags lie before the offset. It is computed by walking the marks in order rather
// than by counting them, because a match inside code is not marked — the tags and
// the matches are not one-to-one, and assuming they were put the anchors in the
// wrong blocks.
func (r *Reader) markShift(start, end int) (int, int) {
	if !r.finding || r.query == "" || len(r.matchList) == 0 {
		return start, end
	}
	shift := r.tagsBefore(start)
	return start - shift, end - shift
}

// tagsBefore is how many bytes of mark tags the marked source holds before an
// offset of the unmarked one.
func (r *Reader) tagsBefore(offset int) int {
	n := len(r.query)
	skip := unmarkable(r.doc.Source)
	tags := 0
	for _, m := range r.matchList {
		if m >= offset {
			break
		}
		// Only a marked match contributes tags.
		if overlaps(skip, m, m+n) {
			continue
		}
		tags++
	}
	return tags * (len(markOpen) + len(markClose))
}

// headingBlock builds one heading and keeps the handle that remembers it.
func (r *Reader) headingBlock(c *ui.Context, d *parser.Document, i int, pal pagePalette) {
	if r.doc == nil {
		return
	}
	b := d.Block(i)
	size := r.headingPx(b.Level)
	row := ui.Column(c).Key(i).FillWidth()
	var handle ui.Handle
	row.Bind(&handle)
	r.headingHandles[i] = handle
	row.Children(func() {
		// The heading is drawn as spans rather than as one text, so that a word
		// Find is looking for can be marked inside it. It used to be plain text,
		// which made a match in a heading count, scroll into view and show no
		// highlight at all — the worst kind of bug, because the count said the
		// search had worked.
		text := ui.RichText(c, r.headingSpans(d, i, pal)...)
		text.Font(pageFamily).FontSize(size).
			FontWeight(headingWeight(b.Level)).
			// Large type wants its letters closer, small type wants them
			// further apart; one tracking value for every size is wrong at one
			// end or the other, so it follows the size. The design sets its
			// titles at -0.04em; this approaches that at the title's size and
			// relaxes as the heading gets smaller, which is what optical sizing
			// does for a face that carries it.
			//
			// A serif carries less tracking than a sans: its sidebearings are
			// already tighter at display sizes, and past about -0.02em the
			// letters begin to touch, which at 55 DIPs is visible as a collision.
			LetterSpacing(-size * 0.016).
			TextColor(uiRGB(triple(palette(pal.dark).ink)))
	})
	// A heading chosen in the outline comes into view in the frame it is
	// chosen, which is what ScrollIntoView is for.
	if r.jumpToBlock == i {
		row.ScrollIntoView()
		r.jumpToBlock = -1
	}
}

// headingSpans splits a heading's words around the marks Find put in them.
//
// The source it is given is the marked one, so a heading can hold <mark> tags;
// they are turned into spans with the theme's Mark behind them rather than
// shown as text. Splitting on the tags is enough because the app wrote them: a
// document's own "<mark>" is escaped by the parser before this point, so the
// only tags here are ours.
func (r *Reader) headingSpans(d *parser.Document, i int, pal pagePalette) []ui.Span {
	text := d.Lines(i)
	text = strings.TrimSpace(text)
	text = strings.TrimLeft(text, "#")
	text = strings.TrimSpace(strings.TrimRight(text, "#"))
	if !strings.Contains(text, markOpen) {
		return []ui.Span{{Text: headingText(d.Lines(i))}}
	}
	var spans []ui.Span
	for {
		open := strings.Index(text, markOpen)
		if open < 0 {
			break
		}
		if open > 0 {
			spans = append(spans, ui.Span{Text: text[:open]})
		}
		rest := text[open+len(markOpen):]
		close := strings.Index(rest, markClose)
		if close < 0 {
			spans = append(spans, ui.Span{Text: rest})
			break
		}
		spans = append(spans, ui.Span{Text: rest[:close], Background: r.markColor()})
		text = rest[close+len(markClose):]
	}
	if len(text) > 0 {
		spans = append(spans, ui.Span{Text: text})
	}
	if len(spans) == 0 {
		return []ui.Span{{Text: headingText(d.Lines(i))}}
	}
	return spans
}

// headingWeight is how much of the hierarchy weight carries: the first level
// is heavier than the text under it, and below that a heading is told apart by
// size alone, so that a page of nested headings does not turn into a wall of
// bold.
func headingWeight(level int) int {
	if level < 1 || level > 6 {
		return headingWeights[5]
	}
	return headingWeights[level-1]
}

// headingScale is the one place the page's heading sizes live, in multiples of
// the body size. It was copied in two places — the renderer's theme and the
// reader's own heading drawing — and the two drifted the moment one was changed.
//
// The shape is the design's: an article's title is much larger than what
// follows, and the rest step down quickly, because a serif carries hierarchy at
// a smaller size difference than a sans does. The scale stops at the body size
// and never goes below it: a heading smaller than the paragraph it introduces
// reads as a footnote, so past level four weight and space carry the hierarchy
// and the size stays where the text is.
var headingScale = [6]float32{2.35, 1.45, 1.22, 1.08, 1.0, 1.0}

// headingWeights are the heading weights, likewise in one place. A serif at 700
// is a slab of ink; 500 holds the hierarchy and keeps the page calm, which is
// the posture the design asks for.
var headingWeights = [6]int{500, 500, 500, 500, 500, 500}

// headingPx is a heading's type size, from the body size: the page's hierarchy
// stays proportional as the reader changes size.
func (r *Reader) headingPx(level int) float32 {
	if level < 1 || level > 6 {
		return r.typ.FontPx
	}
	return r.typ.FontPx * headingScale[level-1]
}

// resolveHeadings turns each heading's rectangle into an offset from the top of
// the page, which is what the saved position needs and what Find interpolates
// between. The anchor is the reading column, so an offset is measured from the
// first line of the page rather than from the window.
func (r *Reader) resolveHeadings(c *ui.Context) {
	if r.doc == nil || len(r.headingHandles) == 0 {
		return
	}
	column := r.column.Bounds(c)
	if column.H == 0 {
		return
	}
	for k := range r.doc.headings {
		h := &r.doc.headings[k]
		handle, ok := r.headingHandles[h.Index]
		if !ok {
			continue
		}
		box := handle.Bounds(c)
		if box.H == 0 {
			continue
		}
		// The anchor is the reading column, which is itself inside the scroll, so
		// the distance between the two is already the heading's offset in the
		// page's content — the scroll has moved both of them together.
		//
		// Adding the scroll offset here was wrong, and it was wrong in a way that
		// hid: at the top of a page the scroll is zero, so every heading measured
		// correctly and every test that never scrolled passed. Scroll down and each
		// offset grew by the distance scrolled, so the outline's idea of where the
		// reader was drifted further from the page the further they read, and a
		// saved position pointed past the end of the document.
		if offset := box.Y - column.Y; offset >= 0 {
			h.Offset = offset
		}
	}
}




