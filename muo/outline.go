package muo

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

// The outline is the document's own headings, in order. The window's source list
// shows them, choosing one scrolls the page to it, and the heading the reader
// stopped at is saved per document so reopening the app returns to it.

// heading is one entry of the outline.
type heading struct {
	Index      int     // the block's index in the parsed document
	Level      int     // 1 to 6
	Text       string  // its words, as shown
	Key        string  // the outline's identity for its row
	Offset     float32 // its distance from the top of the page, DIPs
	CharOffset int     // its offset in the source, bytes
}

// outlineOf reads the headings of a Markdown source in document order.
func outlineOf(src string) []heading {
	doc := parser.Parse(src)
	out := make([]heading, 0, 16)
	// The document's blocks start at its first child: block 0 is the root, whose
	// Next is the end of the document.
	for i := range doc.Children(0) {
		b := doc.Block(i)
		if b.Kind != parser.Heading {
			continue
		}
		text := headingText(doc.Lines(i))
		if text == "" {
			continue
		}
		out = append(out, heading{
			Index:      i,
			Level:      b.Level,
			Text:       text,
			Key:        text + "\x00" + markdown.Slug(text),
			CharOffset: b.Start,
		})
	}
	return out
}

// headingText is a heading line without its hashes.
func headingText(line string) string {
	text := strings.TrimSpace(line)
	text = strings.TrimLeft(text, "#")
	return strings.TrimSpace(strings.TrimRight(text, "#"))
}

// jumpTo scrolls the page to a heading and remembers it as the reader's place.
// The heading's element is asked to come into view in the frame it is chosen,
// which is how ui.Element.ScrollIntoView is meant to be used.
func (r *Reader) jumpTo(h heading) {
	if r.doc == nil {
		return
	}
	r.doc.heading = h.Index
	r.chosenHeading = true
	r.jumpToBlock = h.Index
}

// MinWidth is the narrowest window that still shows the source list and a
// readable measure: the column, the list and the page's margins.
func (r *Reader) MinWidth() float32 {
	column := r.typ.ColumnW
	if column <= 0 {
		column = 420
	}
	return sidebarWidth + column + pageMargin*2
}

// pageMargin is the space between the page's text and the pane's edges.
const pageMargin = 32

var _ = ui.Rect{}
