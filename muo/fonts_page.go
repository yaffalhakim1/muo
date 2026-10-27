package muo

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

// The page's face.
//
// Muo sets its page in Inter, and the window's chrome in the same family: one
// voice for the whole window, so a document reads as the thing the app is for
// rather than as content inside a frame. Inter is chosen for the same reason a
// modern reading app picks it — its open apertures and even rhythm hold up at
// small sizes and at a long measure, where a serif's thick-to-thin strokes start
// to shimmer into the background on a high-density panel.
//
// A serif is bundled too, and the page can be set in it: see pageFamilyName.
// Newsreader is there because the choice between the two is a judgement about
// what the window is, and the only way to make it is to look.
//
// Either family is bundled rather than asked of the system, and the reason is
// the measure. The type model solves the column for a character count, and it
// does that by measuring a character of the face the page is actually drawn in.
// A system sans is Segoe UI on Windows, SF on macOS and something else again on
// Linux, and those are not the same width — so the same document would read at a
// different number of characters per line on each, which is the one thing a
// reader that promises a measure must not do. Both files are on every platform,
// under the SIL Open Font License (muo/fonts/OFL.txt, muo/fonts/OFL-Inter.txt).
//
// Both are variable fonts: one file carries every weight the page asks for, so a
// heading at 600 and a body at 400 cost nothing extra.

//go:embed fonts/newsreader.ttf
var newsreaderTTF []byte

//go:embed fonts/newsreader-italic.ttf
var newsreaderItalicTTF []byte

// Inter, the sans the page is set in. Both families ship: which one the page
// uses is pageFamilyName, so the two can be compared without a rebuild.
//
//go:embed fonts/inter.ttf
var interTTF []byte

//go:embed fonts/inter-italic.ttf
var interItalicTTF []byte

// pageFontErr is why the page's face could not be installed, if it could not.
// It is kept rather than returned to an init, which has nowhere to report to:
// the app reads it once it has somewhere to print.
var pageFontErr error

// The faces are installed when the package loads, not when the app starts,
// because every path that lays out a page needs them — the app's window, and the
// tests that render the window without one. A family that is registered only by
// the app's entry point is a family the tests draw in the system's font, which is
// how a heading loses its face in a screenshot and nobody notices.
func init() { pageFontErr = registerPageFont() }

// PageFontError reports why the page's face could not be installed, or nil.
func PageFontError() error { return pageFontErr }

// The two families the page can be set in.
const (
	familyNewsreader = "Newsreader"
	familyInter      = "Inter"
)

// pageFamilyName is the family the page is drawn in. It is a variable rather
// than a constant so a test can render both without a second build, and so the
// choice is one line when it is made.
//
// The page was set in Newsreader, a serif. Inter is the alternative: the sans
// that Muo's own design notes name, and the face the app's chrome already
// implied. A serif gives prose a rhythm to ride; a sans gives it uniformity,
// which reads as software rather than as a page. Both are defensible, and which
// one is right is a decision about what the window is, so it is a decision to
// make by looking.
var pageFamilyName = familyInter

// pageFamily is the family the page is set in, as Font and the Markdown theme
// take it.
func pageFamily() string { return pageFamilyName }

// registerPageFont installs both faces: the upright cut and the italic of each.
//
// Both cuts of a family are registered under one family name, which is what lets
// a run of text ask for Italic and get the drawn italic rather than a slanted
// upright. The page leans on that: block quotes and emphasis are set in the
// italic, and a synthesised slant in either face is a visible fake.
func registerPageFont() error {
	if err := ui.RegisterFont(newsreaderTTF, familyNewsreader); err != nil {
		return err
	}
	if err := ui.RegisterFont(newsreaderItalicTTF, familyNewsreader); err != nil {
		return err
	}
	if err := ui.RegisterFont(interTTF, familyInter); err != nil {
		return err
	}
	return ui.RegisterFont(interItalicTTF, familyInter)
}
