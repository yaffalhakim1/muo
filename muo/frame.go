package muo

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

// The window's frame: one header, two panels, one page.
//
// The shape is the design's, and it is worth stating because it is the app's
// whole structure. The header carries the document's identity and the commands
// that act on the window. The panels answer the two questions a reader has
// while reading — which file, and where in this file — and each folds away so
// that the page can have the room instead. The page is one column of prose,
// centred, never squeezed: it is the only thing on screen that is the reader's
// reason for being here.
//
// The pieces are built from the toolkit's own widgets and nothing else:
//
//   - the panels are ui.Columns of ui.ButtonBases, not hand-drawn rows, so they
//     get the platform's hit areas, hover, focus rings and keyboard activation;
//   - the header's controls are ui.Buttons with accessible labels, so an
//     icon-only command still says what it does;
//   - the panels animate by width, and a folded panel is not built at all, so a
//     hidden pane costs no work and leaves nothing in the accessibility tree.

const (
	// headerHeight is the row above everything: the document's name, the
	// reading metadata, and the commands. It is tall enough for a 32-point
	// control with air above and below, and it is the same height whether a
	// panel is open or not, so the page never jumps as one folds.
	headerHeight = 56
	// leftPanelWidth is the workspace column: wide enough for a file name of
	// the length a real project has without truncating it to four words.
	leftPanelWidth = 256
	// rightPanelWidth is the outline. It is narrower than the workspace because
	// it holds headings, which are read at a glance, not file names, which are
	// read before opening one.
	rightPanelWidth = 240
	// pageBottomMargin is the room below the last line: enough that the text
	// never runs under the window's bottom edge, which is where a reader's eye
	// stops trusting the scroll.
	pageBottomMargin = 64
	// findBarHeight is the row Find takes under the header.
	findBarHeight = 44
)

// narrowWidth is the window width below which the workspace yields to the page.
// A panel that keeps its column at a narrow width leaves the reading measure so
// short that the page stops being a page, so below this the panels hide
// themselves, as a macOS window does when it is narrowed, and they come back
// when it is widened.
const narrowWidth = 820

// outlineNarrowWidth is the width below which the outline yields first: it is
// the pane a reader can most easily do without, since the page's own headings
// are still in front of them.
const outlineNarrowWidth = 1200

// panelSlide is the spring a panel travels on when it is shown or hidden. A
// spring rather than a curve because the command can be given again mid-flight:
// the width keeps its velocity and turns around where it is. The desktop's
// Reduce Motion setting makes it arrive at once.
func panelSlide(c *ui.Context) ui.Animation {
	if c.Preferences().ReduceMotion {
		return ui.Animation{Duration: animationUnit}
	}
	return ui.Animation{Spring: ui.SpringOptions{
		Duration: 260 * animationUnit,
		Bounce:   0,
	}}
}

// libraryShown is whether the workspace is on screen this frame: the reader's
// choice, unless the window is too narrow to afford it — and a reader who asks
// for it anyway at that width gets it, because a command that appears to do
// nothing is worse than a cramped page.
func (r *Reader) libraryShown(c *ui.Context) bool {
	if r.zen || !r.showLibrary {
		return false
	}
	if c == nil {
		return true
	}
	w, _ := c.Size()
	return float32(w) >= narrowWidth || r.narrowLibrary
}

// outlineShown is the same for the outline, which is wider to afford: it sits
// beside the page, so it competes with the reading measure on both sides at once.
func (r *Reader) outlineShown(c *ui.Context) bool {
	if r.zen || !r.showOutline || r.doc == nil || len(r.doc.headings) == 0 {
		return false
	}
	if c == nil {
		return true
	}
	w, _ := c.Size()
	return float32(w) >= outlineNarrowWidth || r.narrowOutline
}

// FrameView builds the whole window: header, panels and page.
func (r *Reader) FrameView(c *ui.Context) {
	// The window installs the theme before any of its own widgets are built, so
	// the controls and the page agree on one accent and one scale.
	c.SetTheme(ShellTheme(c))
	r.recompute(c)
	r.applyRestore()
	r.shortcuts(c)
	if r.settings {
		r.settingsView(c)
		return
	}
	r.frameView(c)
	r.rememberWhereWeAre()
}

// frameView is the window: the header floating over the page, the panels beside
// it, and the progress bar over all of it.
//
// The header is an overlay, not a row in the layout, and that is the design's own
// structure: its header is `sticky` with a backdrop blur, which is only a
// backdrop blur because content scrolls under it. Putting the header back in the
// flow made that blur a grey band over nothing, and it also left no way to hide
// the header without the page jumping up by its height.
//
// As an overlay it can slide away and leave the page exactly where it was, which
// is what Zen needs: the page keeps its place, and the reader's eye does not have
// to re-find the line it was on.
func (r *Reader) frameView(c *ui.Context) {
	if c.Vibrancy() {
		// The material shows where the view draws no background, so the root
		// paints none and the panes leave their own transparent.
		c.Root().Background(ui.Transparent)
	}

	bar := c.TitleBar()
	left := r.libraryShown(c)
	right := r.outlineShown(c)

	// The panels slide, and the page does not reflow while they do.
	//
	// This is the whole of the performance design, and it is worth stating because
	// the obvious version is wrong. Animating the panels' width and letting the
	// page take the freed room means the reading column is a different width every
	// frame, and every paragraph in the document re-breaks every frame: at 60 Hz
	// that is 60 full text layouts per gesture, and at 240 Hz it is 240. The
	// symptom is a stutter that looks like a slow animation and is really the text
	// engine doing a document's worth of work between two frames.
	//
	// So the measure is solved from where the panels will be, not where they are.
	// The page's column stays at its final width for the whole slide, and what
	// moves is the panels' own clipping edge over it. The page does not reflow, so
	// the frames cost their painting and nothing else.
	row := ui.Row(c).Fill().AlignItems(ui.Stretch)
	r.libReveal = row.AnimateTo("library", boolToFloat(left), panelSlide(c))
	r.outReveal = row.AnimateTo("outline", boolToFloat(right), panelSlide(c))

	row.Children(func() {
		r.workspace(c, r.libReveal)
		r.readingColumnView(c, bar, headerHeight)
		r.outlinePane(c, r.outReveal)
	})

	// The header floats over the page, so it is built after it.
	r.headerOverlay(c, bar, left, right)

	// The footer is the window's, not the page's.
	r.footer(c)

	// The bar is over everything: it is the one mark that belongs to the document
	// rather than to the window.
	r.readingProgressBar(c)
	// The pill floats over the page's foot, and is drawn after it so it is on
	// top of what it reports on.
	r.pill(c)
}

// headerOverlay draws the header over the page.
//
// It is an overlay rather than a row in the layout because the design's header is
// `sticky` with a backdrop blur, which is only a backdrop blur because content
// scrolls under it. As an overlay the page can also scroll behind it, which is
// what makes the hairline under the header the only separation it needs.
func (r *Reader) headerOverlay(c *ui.Context, bar ui.TitleBar, left, right bool) {
	ui.Box(c).Absolute().Top(0).Left(0).Right(0).Height(headerHeight).Children(func() {
		r.header(c, bar, left, right)
	})
}

// pageView is the reading column, with the scroll under it.
//
// The page takes whatever the panels leave and centres its column in that room,
// so the prose stays in the middle of the space it has rather than sliding
// sideways as a panel opens. That is what makes folding a panel feel like the
// page widening rather than the page moving.
func (r *Reader) readingColumnView(c *ui.Context, bar ui.TitleBar, top float32) {
	pal := pageColors(c)

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		box := ui.Box(c).Fill()
		// A file dropped anywhere on the window opens here, which is what the
		// empty page promises. The welcome lights up under the drag itself.
		if dropped := box.DroppedFiles(); len(dropped) > 0 {
			for _, path := range dropped {
				r.OpenIn(path)
			}
		}
		box.Children(func() {
			if r.doc == nil {
				ui.Box(c).Fill().Background(pal.paperColor()).Children(func() {
					r.welcome(c, pal)
				})
			} else {
				// The page's first line clears the header, and the scroll's own
				// bar starts below it for the same reason. With Find open it
				// clears the bar's row too.
				inset := top
				if r.finding {
					inset += findBarHeight
				}
				if r.doc.Format == FormatPDF {
					ui.Scroll(c).Fill().TrackScroll(&r.contentScroll).
						Background(pal.paperColor()).
						ScrollbarInsets(inset, 0, 0).
						Children(func() { r.pdfContent(c, pal, inset) })
				} else {
					ui.Scroll(c).Fill().TrackScroll(&r.contentScroll).
						Background(pal.paperColor()).
						Padding(pageMargin+inset, pageMargin, pageBottomMargin+pageMargin, pageMargin).
						ScrollbarInsets(inset, 0, 0).
						Children(func() { r.pageContent(c, pal) })
				}

				// Find sits in its own row under the header, which the scroll has
				// already made room for: a bar laid over the text would hide the
				// first match it is counting.
				if r.finding {
					ui.Row(c).Absolute().Top(top).Left(0).Right(0).
						Height(findBarHeight).Padding(0, pageMargin/2).
						AlignItems(ui.Center).Children(func() { r.findBar(c) })
				}
			}
		})
	})
}

// pageContent is the reading column, inside the scroll area the page padded.
func (r *Reader) pageContent(c *ui.Context, pal pagePalette) {
	if r.doc.Err != "" {
		ui.Row(c).FillWidth().Children(func() {
			ui.Spacer(c)
			ui.Column(c).Width(r.typ.ColumnW).MaxWidthPercent(100).Children(func() {
				ui.Text(c, r.doc.Err).TextColor(c.Theme().Danger)
			})
			ui.Spacer(c)
		})
		return
	}
	ui.Row(c).FillWidth().Children(func() {
		ui.Spacer(c)
		// The column itself is the measurement anchor: a heading's offset is
		// its distance from this element's top. It must be the column and not a
		// zero-height row above it, because the heading measurement is guarded
		// on the anchor having a height — an empty row has none, so every
		// heading was silently left at offset 0 and the outline could not
		// restore a place.
		column := ui.Column(c).Width(r.typ.ColumnW).MaxWidthPercent(100).
			Bind(&r.column)
		column.Children(func() {
			r.lastColumn = column.Bounds()
			r.resetAnchors()
			r.resolveHeadings(c)
			opts := r.markdownOptions(c, pal)
	markdown.View(c, r.calloutSource(r.highlightedSource()), opts)
			r.endMark(c, pal)
		})
		ui.Spacer(c)
	})
}

// endMark closes the page. A document that simply stops at the foot of a scroll
// reads as though it were cut off; a centred mark says the text ended and the
// window is not still loading.
//
// It is three dots and not the words "· · ·": a screen reader reading a row of
// middle dots aloud is noise, so the mark is drawn and the row is named for what
// it means.
func (r *Reader) endMark(c *ui.Context, pal pagePalette) {
	if r.doc == nil || r.doc.Format == FormatPDF || r.doc.Err != "" {
		return
	}
	ui.Row(c).FillWidth().Center().Padding(r.typ.FontPx*1.8, 0, 0, 0).
		Label("End of document").Children(func() {
		ui.Row(c).Gap(6).Children(func() {
			for i := 0; i < 3; i++ {
				ui.Box(c).Size(4, 4).Radius(2).
					Background(pal.mutedColor().Alpha(0.55))
			}
		})
	})
}

// shortcuts binds the commands the menu bar also offers, so they work while the
// pointer is anywhere in the window.
func (r *Reader) shortcuts(c *ui.Context) {
	cmd := ui.Cmd
	switch {
	case c.Shortcut(cmd, ui.KeyO):
		r.Open()
	case c.Shortcut(cmd, ui.KeyF):
		r.finding = !r.finding
	case c.Shortcut(cmd, ui.KeyComma):
		r.settings = !r.settings
	case c.Shortcut(cmd, ui.KeyB):
		r.showLibrary = !r.libraryShown(c)
		r.narrowLibrary = !r.libraryShown(c)
		r.refreshMenu()
	case c.Shortcut(cmd, ui.KeyEqual):
		r.stepType(1.05)
	case c.Shortcut(cmd, ui.KeyMinus):
		r.stepType(1 / 1.05)
	case c.Shortcut(cmd, ui.Key0):
		r.manual = 0
		if r.display != nil {
			r.display.FontPx = 0
		}
	}
	// Zen is a reading mode, so it is entered and left by one key that the
	// reader can reach without looking: Escape leaves it, as it leaves any
	// other mode.
	if r.zen && c.Shortcut(0, ui.KeyEscape) {
		r.zen = false
	}
	if r.settings && c.Shortcut(0, ui.KeyEscape) {
		r.settings = false
	}
}

// refreshMenu rebuilds the menu bar so that the View menu's panel items say what
// they will do. A test drives the window without an app, so there is nothing to
// rebuild there.
func (r *Reader) refreshMenu() {
	if r.window == nil {
		return
	}
	r.InstallMenu()
}

// The frame's measurements, exported for the app that creates the window.
const (
	// TitleBarHeight is the row above everything.
	TitleBarHeight = headerHeight
	// SidebarWidth is the workspace column's width.
	SidebarWidth = leftPanelWidth
)

// The old names, kept because the tests and the position code still measure
// against them: the row that used to hold the window controls is the header now,
// and the column that used to be a draggable split is a fixed panel.
const (
	titleBarHeight = headerHeight
	titleLeft      = 92
	sidebarWidth   = leftPanelWidth
	outlineWidth   = rightPanelWidth
)

// clampPanel bounds a saved panel width to what the app can lay out. The panels
// are fixed now, so a layout saved from the draggable split is folded back into
// the one width the design uses rather than reproduced.
func clampPanel(w float32) float32 {
	switch {
	case w <= 0:
		return leftPanelWidth
	case w < 168:
		return 168
	case w > 420:
		return 420
	}
	return w
}
