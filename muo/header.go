package muo

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// The header: one row over the whole window, carrying the document's identity on
// the left, what the type model computed in the middle, and the commands on the
// right.
//
// It is a sibling of the body rather than an overlay, so nothing scrolls under
// it and nothing needs a fade to hide the seam. That is the difference from the
// window this replaces: the old title row floated over the page and the page
// scrolled beneath it, which cost a scroll-edge material and a whole set of
// insets to keep text from running into the bar. A header in the layout costs
// none of that.
//
// The left end starts clear of the window controls. With an inset title bar they
// are drawn over the content at the top left, which is exactly where a header
// wants to put its first button; the inset is read from the title bar rather
// than guessed, so it follows the platform.

// headerView is gone: the header is an overlay now, built by headerOverlay in
// the frame. It was a row in the layout for one revision, and that is what made
// its backdrop blur a grey band: there was nothing behind it to blur.
func (r *Reader) header(c *ui.Context, bar ui.TitleBar, left, right bool) {
	t := c.Theme()

	padLeft, padRight := headerInsets(bar)
	row := ui.Row(c).FillWidth().Height(headerHeight).Shrink(0).
		Padding(0, padRight, 0, padLeft).
		AlignItems(ui.Center).DragWindow()
	// The header is a surface, not a material.
	//
	// The design frosts it with a backdrop blur, and that works there because its
	// header is sticky over scrolling content: there is always something behind
	// the blur. A material over the window's own paper is a material doing nothing
	// but tinting — the regular glass maps black to 54% and white to 100%, so on
	// warm paper it came out as a grey band, with a 28-DIP bezel on a 56-DIP bar
	// bending empty space.
	//
	// The paper itself is the right answer, and it is what the design's own 85%
	// wash approximates: the header is the page's colour, a shade separated from
	// the page by a hairline.
	row.Background(t.Surface)
	row.BorderWidth(0, 0, 1, 0).BorderColor(t.Border.Alpha(0.7))

	row.Children(func() {
		if r.doc == nil {
			r.emptyHeaderLeading(c, left)
		} else {
			r.headerLeading(c, left, right)
		}
		ui.Spacer(c)
		r.headerMeta(c)
		ui.Spacer(c)
		r.headerCommands(c, left, right)
	})
}

// headerInsets is the room the window's own controls take on each side of the
// header, plus the gap the header keeps from them.
//
// It is not a constant, and getting it wrong is not a cosmetic error. macOS puts
// the traffic lights at the top left and Windows puts its three buttons at the
// top right, so a header padded on the left alone puts every command on Windows
// underneath the close button. Both sides come from the title bar, which reports
// the room per platform, and the gap is added to whichever side is occupied.
//
// The inset is read from the bar rather than from the OS, so a frameless window
// or a full screen — where the bar reports nothing — needs no special case.
func headerInsets(bar ui.TitleBar) (left, right float32) {
	const gap = 8
	left, right = 12, 12
	if bar.Left > 0 {
		left = bar.Left + gap
	}
	if bar.Right > 0 {
		right = bar.Right + gap
	}
	return left, right
}

// emptyHeaderLeading is the header before a document is open: the toggle, and
// the wordmark in the page's own italic.
//
// The wordmark stands in for the file name, in the place the name will be. That
// is the design's own move, and it is the right one: the empty window is the
// app's frontispiece, and it should carry the app's name rather than a blank
// where a document's would go.
func (r *Reader) emptyHeaderLeading(c *ui.Context, left bool) {
	t := c.Theme()
	ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Shrink(0).Children(func() {
		label := "Show Files"
		if left {
			label = "Hide Files"
		}
		if b := headerButton(c, iconSidebar, label, left); b.Clicked() {
			r.toggleLibrary(c)
		}
		ui.Text(c, "Muo").Font(pageFamily()).Italic().FontSize(15).
			LetterSpacing(0.2).TextColor(t.Text).SingleLine()
	})
}

// headerLeading is the explorer toggle, the document's name, and the mark that
// says the file is the one on disk.
func (r *Reader) headerLeading(c *ui.Context, left, right bool) {
	t := c.Theme()

	ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Shrink(0).Children(func() {
		label := "Show Files"
		if left {
			label = "Hide Files"
		}
		b := headerButton(c, iconSidebar, label, left)
		if b.Clicked() {
			r.toggleLibrary(c)
		}

		if r.doc == nil {
			return
		}
		ui.Text(c, filepath.Base(r.doc.Path)).
			Font("monospace").FontSize(typeCallout).FontWeight(500).
			TextColor(t.Text).SingleLine().MaxWidth(220)
		// A dot rather than a word: it reports the file's state at a glance and
		// takes no room from the name. It is read-only, and the tooltip says so
		// rather than claiming a save that never happens.
		ui.Box(c).Size(6, 6).Radius(3).Shrink(0).
			Background(t.Success.Alpha(0.7)).
			Label("Read from disk").Tooltip("Read from disk")
	})
}

// headerMeta is the reading metadata: the measure, the apparent size, and how
// long the document takes to read.
//
// It is the app's own claim stated quietly in the one place a reader passes on
// the way in. It is not a control, so it is muted and it does not respond to the
// pointer; and it is hidden below the width where three numbers stop fitting
// beside two sets of commands.
func (r *Reader) headerMeta(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(t.Space(1.5)).AlignItems(ui.Center).Shrink(0).
		Label("Reading metadata").
		Description(fmt.Sprintf(
			"%.0f characters per line, an apparent character height of %.2f degrees.",
			r.typ.ChPerLine, r.typ.Angle)).
		Children(func() {
			// The measure is the promise, and the angle is the evidence for it.
			// A read-time used to sit third here, counted from the source's
			// whitespace — which counts code, table cells and link targets as
			// prose, so on a document with 70 code blocks it reported a number
			// with no meaning. Two facts about the type are a claim; three,
			// where the third is wrong, is a header nobody reads.
			metaText(c, fmt.Sprintf("%.0fch", r.typ.ChPerLine))
			metaDot(c)
			metaText(c, fmt.Sprintf("%.2f° angle", r.typ.Angle))
		})
}

func metaText(c *ui.Context, s string) {
	t := c.Theme()
	ui.Text(c, s).Font("monospace").FontSize(typeSubhead).
		FontFeatures("tnum").TextColor(t.TextMuted).SingleLine()
}

func metaDot(c *ui.Context) {
	t := c.Theme()
	ui.Text(c, "·").FontSize(typeSubhead).
		TextColor(t.Border).SingleLine()
}

// headerCommands is the right-hand set: search, the outline toggle, the reading
// settings, and Zen, with a hairline between the commands that act on the
// document and the one that acts on the window.
func (r *Reader) headerCommands(c *ui.Context, left, right bool) {
	t := c.Theme()

	ui.Row(c).Gap(t.Space(1)).AlignItems(ui.Center).Shrink(0).Children(func() {
		if r.doc != nil {
			if headerButton(c, iconSearch, "Find in Document", r.finding).Clicked() {
				r.finding = !r.finding
			}
		}

		outLabel := "Show Page Outline"
		if right {
			outLabel = "Hide Page Outline"
		}
		if b := headerButton(c, iconOutline, outLabel, right); b.Clicked() {
			r.toggleOutline(c)
		}

		if headerButton(c, iconSettings, "Reading Settings", r.settings).Clicked() {
			r.settings = !r.settings
		}

		ui.Box(c).Width(1).Height(t.Rem(1.1)).Shrink(0).
			Background(t.Border).Margin(t.Space(1), t.Space(1))

		if zenButton(c, r.zen).Clicked() {
			r.zen = !r.zen
			// Leaving Zen must not restore a panel the reader had folded before
			// entering it, so the reader's own choice is left untouched: Zen is
			// a mode over the top of it, not a change to it.
		}
	})
}

// headerButton is an icon-only command in the header. It carries the active
// state of the thing it toggles, because a toggle that looks the same either way
// makes the reader press it to find out.
func headerButton(c *ui.Context, mark *ui.SVG, label string, active bool) ui.Element {
	t := c.Theme()
	b := ui.Button(c, "").Size(30, 28).Radius(6).Padding(0).
		Background(ui.Transparent).BorderWidth(0).
		TextColor(t.TextMuted).
		Label(label).Tooltip(label)
	switch {
	case active:
		b.Background(t.SurfaceHover)
		b.BorderWidth(1).BorderColor(t.Border)
		b.TextColor(t.Text)
	case b.Hovered():
		b.Background(t.SurfaceHover)
		b.TextColor(t.Text)
	}
	b.Children(func() { ui.Icon(c, mark).Size(16, 16) })
	return b
}

// zenButton is the one command in the header that carries a word. Zen is a mode,
// and a mode a reader cannot name is a mode they will not find.
func zenButton(c *ui.Context, on bool) ui.Element {
	t := c.Theme()
	b := ui.Button(c, "").Height(28).Radius(6).Padding(0, t.Space(2)).
		Background(ui.Transparent).BorderWidth(0).
		TextColor(t.TextMuted).
		Label("Zen Mode").Tooltip("Zen Mode")
	switch {
	case on:
		b.Background(t.SurfaceHover)
		b.BorderWidth(1).BorderColor(t.Border)
		b.TextColor(t.Text)
	case b.Hovered():
		b.Background(t.SurfaceHover)
		b.TextColor(t.Text)
	}
	b.Children(func() {
		ui.Icon(c, iconZen).Size(14, 14)
		ui.Text(c, "Zen").FontSize(typeCallout).SingleLine()
	})
	return b
}

// toggleLibrary and toggleOutline are the two panel commands. Both remember an
// explicit request at a narrow width, because a command that appears to do
// nothing is worse than a cramped page.
func (r *Reader) toggleLibrary(c *ui.Context) {
	shown := r.libraryShown(c)
	r.showLibrary = !shown
	r.narrowLibrary = !shown
	r.refreshMenu()
}

func (r *Reader) toggleOutline(c *ui.Context) {
	shown := r.outlineShown(c)
	r.showOutline = !shown
	r.narrowOutline = !shown
	r.refreshMenu()
}

// readingMinutes is how long the document takes to read at a stated pace.
//
// 230 words a minute is the middle of the range the reading literature reports
// for silent reading of prose on a screen by an adult reading for comprehension.
// It is stated here rather than inline so the number has one home.
const readingWordsPerMinute = 230

func (r *Reader) readingMinutes() int {
	if r.doc == nil || r.doc.Source == "" {
		return 0
	}
	words := len(strings.FieldsFunc(r.doc.Source, func(ch rune) bool {
		return unicode.IsSpace(ch)
	}))
	minutes := (words + readingWordsPerMinute - 1) / readingWordsPerMinute
	if minutes < 1 && words > 0 {
		return 1
	}
	return minutes
}

var _ = utf8.RuneLen
