package muo

import (
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

// The window's frame: a sidebar of paper, a page of paper, and one title row
// over both.
//
// The pieces are worth naming, because getting them wrong is what makes a
// window look assembled rather than designed:
//
//   - the sidebar is a real ui.Sidebar with a selected id, not buttons styled to
//     look like rows, so it gets the platform's row heights, hover and keyboard;
//   - the title bar is inset, so the window controls sit over the sidebar and the
//     content pane draws its own title row instead of a second toolbar;
//   - the material shows wherever the view paints nothing, so the sidebar is
//     translucent and only the page is opaque;
//   - the sidebar and the page are the same warm neutral, so the window reads as
//     one sheet of paper rather than two surfaces meeting at a seam.

const (
	// titleBarHeight is the inset title bar's height: the room the window
	// controls take, and the row that carries the document's name.
	titleBarHeight = 52
	// sidebarWidth is the width the navigation column opens at. It is wide
	// enough for a heading of the length this kind of document actually has, so
	// the outline does not truncate every entry to its first four words.
	sidebarWidth = 236
	// sidebarMin and sidebarMax bound the width the reader can drag it to. The
	// floor is where a source list stops being one — a row of icons with no room
	// for its words — and the ceiling is where it starts taking the page's share
	// of a normal window.
	sidebarMin = 168
	sidebarMax = 420
	// outlineWidth is the right pane's width, and its bounds. It is narrower than
	// the library: it holds headings, which are read at a glance, not file names,
	// which are read before opening one.
	outlineWidth = 220
	outlineMin   = 150
	outlineMax   = 360
	// pageBottomMargin is the room below the last line of a page: enough that
	// the text never runs under the window's bottom edge, which is where a
	// reader's eye stops trusting the scroll.
	pageBottomMargin = 64
	// titleLeft is where a pane's title starts: clear of the window controls,
	// which the inset title bar draws over the content.
	titleLeft = 92
	// findBarHeight is the row Find takes under the title bar.
	findBarHeight = 44
)

// clampSidebar keeps the reader's width inside the range the app can lay out.
// It is applied before the split is built, so a drag past either end stops
// there rather than producing a frame the page cannot survive.
func clampSidebar(w float32) float32 {
	switch {
	case w <= 0:
		return sidebarWidth
	case w < sidebarMin:
		return sidebarMin
	case w > sidebarMax:
		return sidebarMax
	}
	return w
}

// clampOutline is clampSidebar for the right pane.
func clampOutline(w float32) float32 {
	switch {
	case w <= 0:
		return outlineWidth
	case w < outlineMin:
		return outlineMin
	case w > outlineMax:
		return outlineMax
	}
	return w
}

// sidebarSlide is the spring the sidebar travels on when it is shown or hidden.
// A spring rather than a curve because the command can be given again
// mid-flight: the width keeps its velocity and turns around where it is,
// instead of stopping dead and starting over. The desktop's Reduce Motion
// setting makes it arrive at once.
func sidebarSlide(c *ui.Context) ui.Animation {
	if c.Preferences().ReduceMotion {
		return ui.Animation{Duration: animationUnit}
	}
	return ui.Animation{Spring: ui.SpringOptions{
		Duration: 260 * animationUnit,
		Bounce:   0,
	}}
}

// frameView is the window, in three panes: the library on the left, the page in
// the middle, and the document's own outline on the right.
//
// The split is worth stating because it is the app's whole structure. The left
// pane answers "which file", and is the one thing a reader with several files
// open needs; it holds only documents, so a long list never pushes the page
// around. The right pane answers "where in this file", which is a question about
// the page and not about the window — so it sits beside the page it describes,
// and it can be folded away when the reader is just reading.
//
// Both panes are ui.Split's, so each divider gives the pointer a resize cursor,
// moves with the arrows once focused, and reports itself as a splitter.
func (r *Reader) frameView(c *ui.Context) {
	t := c.Theme()
	vibrant := c.Vibrancy()
	if vibrant {
		// The material shows where the view draws no background, so the root
		// paints none and the panes leave their own transparent.
		c.Root().Background(ui.Transparent)
	}

	bar := c.TitleBar()
	body := ui.Row(c).Fill().AlignItems(ui.Stretch)
	library := body.AnimateTo("library", boolToFloat(r.libraryShown(c)), sidebarSlide(c))
	outline := body.AnimateTo("outline", boolToFloat(r.outlineShown(c)), sidebarSlide(c))
	body.Children(func() {
		// A pane is built while it is on screen or on its way out, and not at all
		// once it has gone: a hidden pane that is still built is work every frame,
		// and it leaves its rows in the window's accessibility tree where a screen
		// reader can reach what nobody can see.
		if library >= 0.002 {
			width := clampSidebar(r.sidebarW) * library
			split := ui.Split(c, &width, func() {
				pane := ui.Column(c).Fill()
				if !vibrant {
					pane.Background(t.Surface)
				}
				pane.Children(func() { r.library(c, vibrant) })
			}, func() {
				r.center(c, bar, outline)
			}).Fill()
			// Only a drag or an arrow key changes the divider's size, and only then
			// is the reader's width updated: writing it every frame would make the
			// reveal animation overwrite the width the reader chose.
			if split.Changed() {
				r.sidebarW = clampSidebar(width)
			}
			return
		}
		r.center(c, bar, outline)
	})

	r.paneToggles(c, bar)
}

// center is the page, with the outline beside it when it is shown.
//
// The pane's width is animated directly, as a Row of two columns, and the outline
// is the second. A ui.Split was tried first and does not work here: it reads its
// *size against its own last-frame room, so feeding it a value that changes each
// frame fights its own layout and the pane snaps instead of sliding.
//
// A Row gives each column the width it asks for, so the outline's width is the
// reveal times the reader's chosen width, and the page takes the rest with Grow.
// That is what makes the fold a slide: every frame the outline is a little
// narrower and the page a little wider.
func (r *Reader) center(c *ui.Context, bar ui.TitleBar, outline float32) {
	if outline <= 0 {
		r.pane(c, bar, r.libraryShown(c))
		return
	}
	outlineW := clampOutline(r.outlineW) * outline
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		page := ui.Column(c).Grow(1).MinWidth(0)
		page.Children(func() { r.pane(c, bar, r.libraryShown(c)) })

		// The seam, then the pane: a hairline of the page's own rule colour, as
		// between the library and the page.
		ui.Box(c).Width(1).Shrink(0).Background(c.Theme().Border.Alpha(0.9))
		side := ui.Column(c).Width(outlineW).Shrink(0).FillHeight().Clip()
		if !c.Vibrancy() {
			side.Background(c.Theme().Surface)
		}
		side.Children(func() { r.outlinePane(c) })
	})
	_, w := c.Size()
	r.outlineX = float32(w) - outlineW
}

// pageFloor is the least width the page may be squeezed to: its reading column
// plus the room the column is centred in. Below this the pane would rather not be
// shown at all, which is what the outline's own threshold decides.
func pageFloor(columnW float32) float32 {
	if columnW <= 0 {
		columnW = 420
	}
	return columnW + pageMargin*2
}

// narrowWidth is the window width below which the panes yield to the page.
//
// A pane that keeps its column at a narrow width leaves the reading measure so
// short that the page stops being a page: at 720 DIPs the column fell to 38
// characters, well under the 45 the model holds as the floor. So below this the
// panes hide themselves, which is what a macOS window does when it is narrowed,
// and they come back when it is widened
// (sidebars.md › Desktop: "Consider automatically hiding and revealing a
// sidebar when its container window resizes").
const narrowWidth = 820

// outlineNarrowWidth is the window width below which the outline yields first:
// it is the pane a reader can most easily do without, since the page's own
// headings are still in front of them. It is higher than the library's threshold
// because the outline sits on top of the page's measure: at 1100 DIPs, showing
// all three panes left the column at 51 characters, under the 45 the model holds
// as its floor plus the room a reader needs to read comfortably.
const outlineNarrowWidth = 1200

// libraryShown is whether the left pane is on screen this frame: the reader's
// choice, unless the window is too narrow to afford it — and a reader who asks
// for it anyway at that width gets it, because a command that appears to do
// nothing is worse than a cramped page.
func (r *Reader) libraryShown(c *ui.Context) bool {
	if !r.showLibrary {
		return false
	}
	w, _ := c.Size()
	return float32(w) >= narrowWidth || r.narrowLibrary
}

// outlineShown is the same for the right pane, which is wider to afford: it sits
// beside the page, so it competes with the reading measure on both sides at once.
func (r *Reader) outlineShown(c *ui.Context) bool {
	if !r.showOutline || r.doc == nil || len(r.doc.headings) == 0 {
		return false
	}
	w, _ := c.Size()
	return float32(w) >= outlineNarrowWidth || r.narrowOutline
}

// library is the left pane: the button that opens a file, then the files that
// are open.
//
// The button is a real ui.Button, not a hand-styled row, so it gets the
// platform's hit area, focus ring, press state and keyboard activation for free.
// It carries a plus sign rather than a folder because it adds a file to the list
// below it — a folder would say "browse", which is only half of what happens.
func (r *Reader) library(c *ui.Context, vibrant bool) {
	t := c.Theme()
	// Room for the window controls, which sit over this pane.
	ui.Box(c).Height(titleBarHeight).Shrink(0).DragWindow()

	ui.Row(c).FillWidth().Padding(t.Space(2), t.Space(2)).Children(func() {
		open := ui.Button(c, "").FillWidth().Label("Open a File").
			Tooltip("Open a File").Radius(t.Radius)
		open.Children(func() {
			ui.Icon(c, iconPlus).Size(15, 15)
			ui.Text(c, "Open a File").SingleLine()
		})
		if open.Clicked() {
			r.Open()
		}
	})

	if len(r.documents) == 0 {
		return
	}

	list := ui.Sidebar(c, &r.sidebarChoice, func() {
		ui.SidebarSection(c, "Open Files", nil, func() {
			for _, d := range r.documents {
				icon := iconDocument
				if d.Format == FormatPDF {
					icon = iconPDF
				}
				ui.SidebarItem(c, "doc:"+d.Path, icon, d.Title)
			}
		})
	}).Grow(1).Label("Open Files")
	if vibrant {
		list.Background(ui.Transparent)
	}
	if list.Changed() {
		r.activate(r.sidebarChoice)
	}
}

// outlinePane is the right pane: the document's own headings.
//
// It is a source list of its own rather than a section of the library, because
// it answers a question about the page — where in this file — and belongs beside
// the page. Choosing a heading scrolls to it, exactly as it did when the outline
// lived on the left.
//
// Its title is the list's own section header, exactly as "Open Files" is in the
// library, and not a row in the title bar. The title bar is where the window
// controls and this pane's toggle are drawn; putting words there ran them under
// the buttons and made the header look broken rather than placed.
func (r *Reader) outlinePane(c *ui.Context) {
	t := c.Theme()
	// Room for the window controls, which sit over this pane.
	ui.Box(c).Height(titleBarHeight).Shrink(0).DragWindow()

	if r.doc == nil || len(r.doc.headings) == 0 {
		ui.Column(c).Padding(t.Space(3), t.Space(3)).Gap(t.Space(1)).Children(func() {
			ui.Text(c, "On This Page").FontSize(typeSubhead).FontWeight(600).
				LetterSpacing(0.3).TextColor(t.TextMuted)
			ui.Text(c, "This page has no headings.").
				FontSize(typeCallout).TextColor(t.TextMuted)
		})
		return
	}

	list := ui.Sidebar(c, &r.outlineChoice, func() {
		ui.SidebarSection(c, "On This Page", nil, func() {
			for _, h := range r.doc.headings {
				r.outlineItem(c, h)
			}
		})
	}).Grow(1).FillWidth().Label("On This Page")
	// The list's own box, kept for the tests that click its rows: a heading's
	// words appear twice on screen, in the page and here, so a click by text
	// finds the page's and proves nothing about the outline.
	r.outlineBounds = list.LayoutBounds()
	if c.Vibrancy() {
		list.Background(ui.Transparent)
	}
	// The outline keeps its own choice, apart from the library's: they are two
	// lists, and a choice made in one must not be read as a choice in the other.
	if list.Changed() {
		r.activate(r.outlineChoice)
	}
}

// outlineItem is one heading in the source list. A heading below the first is
// indented by its depth, which is what makes the list read as the document's
// shape rather than as a flat list of strings — the same job indentation does
// in every outline, and the reason a reader can find a section by its place.
//
// The indent is padding on the row itself, not a wrapper around it: a
// ui.SidebarItem only reports a choice while it is inside a ui.Sidebar, and a
// wrapper would put it in a Row of its own, outside the list.
func (r *Reader) outlineItem(c *ui.Context, h heading) {
	depth := h.Level - 1
	if depth < 0 {
		depth = 0
	}
	if depth > 3 {
		depth = 3
	}
	item := ui.SidebarItem(c, "h:"+h.Key, nil, h.Text)
	if depth > 0 {
		item.PaddingStart(float32(depth) * 12)
	}
}

// activate turns a chosen sidebar row into an action.
func (r *Reader) activate(choice string) {
	switch {
	case choice == "open":
		r.Open()
	case len(choice) > 4 && choice[:4] == "doc:":
		for i, d := range r.documents {
			if "doc:"+d.Path == choice {
				r.selectDocument(i)
				return
			}
		}
	case len(choice) > 2 && choice[:2] == "h:":
		for _, h := range r.doc.headings {
			if "h:"+h.Key == choice {
				r.jumpTo(h)
				return
			}
		}
	}
}

// pane is the page: the reading column, with the title bar floating over it and
// the content fading into the page's colour as it scrolls beneath. That is the
// macOS 26 scroll-edge pattern, and it is why there is no rule between the bar
// and the page: the fade is the separation.
func (r *Reader) pane(c *ui.Context, bar ui.TitleBar, library bool) {
	t := c.Theme()
	pal := pageColors(c)

	// The title starts clear of whatever is drawn over this pane's left edge: the
	// window controls and the library's toggle when the library is hidden, or
	// nothing but the pane's own margin when it is shown.
	left := float32(titleLeft)
	if !library {
		left = bar.Left + 8 + 32 + 20
	}
	// The right end clears the outline's toggle, which is drawn over this pane
	// when the outline is hidden or when this pane reaches the window's edge.
	right := bar.Right + 20
	if r.outlineShown(c) {
		right = 20
	} else {
		right = bar.Right + 8 + 32 + 12
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		page := ui.Box(c).Fill()
		// A file dropped anywhere on the window opens here, which is what the
		// empty page promises. The welcome lights up under the drag itself.
		if dropped := page.DroppedFiles(); len(dropped) > 0 {
			for _, path := range dropped {
				r.OpenIn(path)
			}
		}
		page.Children(func() {
			if r.doc == nil {
				ui.Box(c).Fill().Background(pal.paperColor()).Children(func() {
					r.welcome(c, pal)
				})
			} else {
				// The page scrolls the whole height, its first line clearing the
				// bar; what has scrolled up fades out under it. With Find open
				// the first line clears the bar and the bar's own row, so the
				// bar is never laid over the text being searched.
				top := float32(titleBarHeight)
				if r.finding {
					top += findBarHeight
				}
				// A PDF's page is a picture drawn in the same scroll container as
				// a page of text, so it scrolls — and Find's scroll lands —
				// exactly as text does.
				if r.doc.Format == FormatPDF {
					ui.Scroll(c).Fill().TrackScroll(&r.contentScroll).
						Background(pal.paperColor()).
						ScrollbarInsets(top, 0, 0).
						Children(func() { r.pdfContent(c, pal, top) })
				} else {
					ui.Scroll(c).Fill().TrackScroll(&r.contentScroll).
						Background(pal.paperColor()).
						Padding(top, pageMargin, pageBottomMargin+pageMargin, pageMargin).
						ScrollbarInsets(top, 0, 0).
						Children(func() { r.pageContent(c, pal) })
				}

				// The fade, painted after the content so it has something to fade.
				ui.Box(c).Absolute().Top(0).Left(0).Right(0).
					Height(top+14).PassThrough().
					Material(glass.ScrollEdge{Background: pal.paperColor()})

				// Find sits in the chrome's own row under the title, which the
				// page has already made room for: a bar laid over the text
				// would hide the first match it is counting.
				if r.finding {
					ui.Row(c).Absolute().Top(titleBarHeight).Left(0).Right(0).
						Height(findBarHeight).Padding(0, pageMargin/2).
						AlignItems(ui.Center).Children(func() { r.findBar(c) })
				}
			}

			// The bar itself, over the fade.
			ui.Row(c).Absolute().Top(0).Left(0).Right(0).Height(titleBarHeight).
				Padding(0, right, 0, left).
				AlignItems(ui.Center).DragWindow().Children(func() {
				name := "Muo"
				if r.doc != nil {
					name = r.doc.Title
				}
				ui.Text(c, name).FontSize(typeTitle3).FontWeight(600).
					LetterSpacing(-0.1).SingleLine().MinWidth(0).Shrink(1)
				ui.Spacer(c)
				ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Shrink(0).Children(func() {
					// The readout is status, not a control: it sits apart from
					// the buttons with a hairline between them, so a glance
					// separates what reports from what acts
					// (toolbars.md › Best practices: choose items deliberately).
					typeReadout(c, r.typ)
					ui.Box(c).Width(1).Height(t.Rem(0.9)).
						Background(t.Border).AlignSelf(ui.Center)
					r.paneButtons(c)
				})
			})
		})
	})
}

// paneButtons are the two commands the title row carries. Find is only offered
// while there is a page to find in.
//
// Both are borderless until the pointer is on them: a title row with two
// outlined buttons in it reads as a second toolbar stacked under the window's
// own, which is the thing an inset title bar exists to avoid.
func (r *Reader) paneButtons(c *ui.Context) {
	if r.doc != nil {
		if iconButton(c, iconSearch, "Find in Document", 30, 26).Clicked() {
			r.finding = !r.finding
		}
	}
	if iconButton(c, iconSettings, "Reading Settings", 30, 26).Clicked() {
		r.settings = !r.settings
	}
}

// pageContent is the reading column, inside the scroll area the pane padded.
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
	// A PDF is not text set in a column: it is a page drawn as large as the
	// window allows, and it scrolls in the pane's own scroll container, which the
	// branch above gives it.
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
			r.resetAnchors()
			r.resolveHeadings(c)
			opts := r.markdownOptions(c, pal)
			markdown.View(c, r.highlightedSource(), opts)
			r.endMark(c, pal)
		})
		ui.Spacer(c)
	})
}

// blockMarkers is no longer needed: the marker is built by the Block hook, in
// the same parent and immediately before the block it belongs to.

// endMark closes the page. A document that simply stops at the foot of a
// scroll reads as though it were cut off; a centred mark says the text ended
// and the window is not still loading.
func (r *Reader) endMark(c *ui.Context, pal pagePalette) {
	if r.doc == nil || r.doc.Format == FormatPDF || r.doc.Err != "" {
		return
	}
	ui.Row(c).FillWidth().Center().Padding(r.typ.FontPx*1.8, 0, 0, 0).Children(func() {
		ui.Text(c, "· · ·").FontSize(r.typ.FontPx * 1.05).
			TextColor(pal.mutedColor()).LetterSpacing(3)
	})
}

// sidebarToggle sits next to the window controls, whether the sidebar shows or
// not, as the vibrancy example puts it. It is a plain button with an icon and
// an accessible label, which is the documented way to draw an icon-only
// control (docs/ui/button.md).
func (r *Reader) paneToggles(c *ui.Context, bar ui.TitleBar) {
	t := c.Theme()
	// The library's toggle sits where it always has, beside the window controls:
	// the left pane is the one a reader toggles most, and the control for it
	// belongs where the eye already goes.
	libShown := r.libraryShown(c)
	libLabel := "Hide Library"
	if !libShown {
		libLabel = "Show Library"
	}
	b := ui.Button(c, "").Key("toggle-library").Absolute().Left(bar.Left+8).
		Top((titleBarHeight-28)/2).Size(32, 28).Radius(6).Padding(0).
		Background(ui.Transparent).BorderWidth(0).TextColor(t.TextMuted).
		Label(libLabel).Tooltip(libLabel)
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() { ui.Icon(c, iconSidebar).Size(16, 16) })
	if b.Clicked() {
		r.showLibrary = !libShown
		r.narrowLibrary = !libShown
		r.refreshMenu()
	}

	// The outline's toggle sits at the right end of the title row, on the same
	// side as the pane it opens: a control belongs on the side of the thing it
	// changes.
	if r.doc == nil || len(r.doc.headings) == 0 {
		return
	}
	outShown := r.outlineShown(c)
	outLabel := "Hide Page Outline"
	if !outShown {
		outLabel = "Show Page Outline"
	}
	o := ui.Button(c, "").Key("toggle-outline").
		Absolute().Top((titleBarHeight-28)/2).Right(bar.Right + 8).Size(32, 28).
		Radius(6).Padding(0).
		Background(ui.Transparent).BorderWidth(0).TextColor(t.TextMuted).
		Label(outLabel).Tooltip(outLabel)
	if o.Hovered() {
		o.Background(t.SurfaceHover)
	}
	o.Children(func() { ui.Icon(c, iconOutline).Size(16, 16) })
	if o.Clicked() {
		r.showOutline = !outShown
		r.narrowOutline = !outShown
		r.refreshMenu()
	}
}

// refreshMenu rebuilds the menu bar so that the View menu's sidebar item says
// what it will do. A test drives the window without an app, so there is nothing
// to rebuild there.
func (r *Reader) refreshMenu() {
	if r.window == nil {
		return
	}
	r.InstallMenu()
}

// The frame's measurements, exported for the app that creates the window.
const (
	// TitleBarHeight is the inset title bar's height.
	TitleBarHeight = titleBarHeight
	// SidebarWidth is the navigation column's width.
	SidebarWidth = sidebarWidth
)

// FrameView builds the whole window: sidebar, title bar and page.
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
	if r.settings && c.Shortcut(0, ui.KeyEscape) {
		r.settings = false
	}
}
