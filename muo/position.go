package muo

import (
	"math"

	"github.com/egoist/mygo"
)

// Remembering the reader's place. The page records its scroll offset every
// frame, the app writes the positions when it closes, and opening a document
// again puts the reader back at the heading they stopped at.

// rememberWhereWeAre records the current document's place. It runs each frame
// while a Markdown document is open: cheap, because it only copies two numbers,
// and correct even if the app is killed before its close handler runs.
func (r *Reader) rememberWhereWeAre() {
	doc := r.doc
	if doc == nil || doc.Format != FormatMarkdown {
		return
	}
	heading := doc.heading
	title := ""
	for _, h := range doc.headings {
		if h.Index == heading {
			title = h.Text
			break
		}
	}
	// The heading under the top of the page is where the reader is, unless they
	// chose one in the outline and have not scrolled since.
	if len(doc.headings) > 0 && !r.chosenHeading {
		best := -1
		bestOffset := float32(-1)
		for _, h := range doc.headings {
			if h.Offset <= r.contentScroll.Y+8 && h.Offset >= bestOffset {
				best, bestOffset = h.Index, h.Offset
			}
		}
		if best >= 0 {
			heading = best
			for _, h := range doc.headings {
				if h.Index == best {
					title = h.Text
					break
				}
			}
		}
	}
	if heading != doc.heading {
		doc.heading = heading
	}
	rememberPosition(doc.Path, readingState{
		Position: r.contentScroll.Y,
		Heading:  heading,
		Title:    title,
	})
}

// applyRestore puts the reader back once the heading it stopped at has been
// measured: the stored pixel offset was only a first guess, because the window
// may be a different size than when the app closed.
func (r *Reader) applyRestore() {
	if !r.restorePending || r.doc == nil {
		return
	}
	doc := r.doc
	for _, h := range doc.headings {
		if h.Offset <= 0 {
			return // the page has not been laid out far enough yet
		}
		if h.Index == doc.heading {
			r.contentScroll.Y = h.Offset
			r.restorePending = false
			return
		}
	}
	// The heading is gone; the pixel offset stands.
	r.restorePending = false
}

// Save closes out the app, writing the reader's positions. A failure is
// reported to the caller rather than swallowed: silently losing the place is
// the bug these positions exist to prevent.
func (r *Reader) Save() error {
	r.rememberWhereWeAre()
	if err := saveState(); err != nil {
		return err
	}
	return saveRecentTimes()
}

// ScrolledToEnd reports whether the page is at its bottom, for the callers that
// want to remember the end rather than an offset.
func (r *Reader) ScrolledToEnd() bool {
	return r.contentScroll.MaxY > 0 && r.contentScroll.Y >= r.contentScroll.MaxY-1 ||
		r.contentScroll.Y >= math.MaxFloat32/2
}

var _ = mygo.App
