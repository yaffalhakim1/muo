package muo

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

// The page's face.
//
// Muo sets its page in a serif, because that is what a page of prose has been
// set in for five hundred years and what the eye still expects: a serif's
// strokes swell and thin along the word, which gives the eye a running rhythm
// to ride, where a uniform sans gives it none. The window's own chrome stays in
// the system's sans, so the page reads as the document and everything around it
// as the app.
//
// The family is bundled rather than asked of the system, and the reason is the
// measure. The type model solves the column for a character count, and it does
// that by measuring a character of the face the page is actually drawn in. A
// system serif is Georgia on Windows and macOS and something else again on
// Linux, and those are not the same width — so the same document would read at
// a different number of characters per line on each, which is the one thing a
// reader that promises a measure must not do. Newsreader is one file, on every
// platform, under the SIL Open Font License (muo/fonts/OFL.txt).
//
// It is a variable font: one file carries every optical size and weight the
// page asks for, so the heading at 600 and the body at 400 cost nothing extra.

//go:embed fonts/newsreader.ttf
var newsreaderTTF []byte

//go:embed fonts/newsreader-italic.ttf
var newsreaderItalicTTF []byte

// pageFontErr is why the page's face could not be installed, if it could not.
// It is kept rather than returned to an init, which has nowhere to report to:
// the app reads it once it has somewhere to print.
var pageFontErr error

// The face is installed when the package loads, not when the app starts, because
// every path that lays out a page needs it — the app's window, and the tests that
// render the window without one. A family that is registered only by the app's
// entry point is a family the tests draw in the system's font, which is how a
// heading loses its serif in a screenshot and nobody notices.
func init() { pageFontErr = registerPageFont() }

// PageFontError reports why the page's face could not be installed, or nil.
func PageFontError() error { return pageFontErr }

// pageFamily is the family the page is set in, as Font and the Markdown theme
// take it. The fallbacks are there so that a registration that somehow failed
// still draws a serif rather than dropping to the window's sans.
const pageFamily = "Newsreader"

// registerPageFont installs the page's face: the upright cut and the italic.
//
// Both are registered under the same family, which is what lets a run of text
// ask for Italic and get the drawn italic rather than a slanted upright. The
// design leans on that: its tagline, its block quotes and its emphasis are set
// in the italic, and a synthesised slant in a serif is a visible fake.
func registerPageFont() error {
	if err := ui.RegisterFont(newsreaderTTF, pageFamily); err != nil {
		return err
	}
	return ui.RegisterFont(newsreaderItalicTTF, pageFamily)
}
