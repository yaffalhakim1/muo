package muo

import (
	"path/filepath"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The two panels beside the page.
//
// They dock, they do not float. A floating drawer over the page is a web app's
// way of doing this, because a web page has no layout of its own to give away; a
// window does. Docked columns keep the page's measure honest — the column is
// solved from the width the page actually has — and they let both panels be read
// at once, which is what a reader comparing a file's outline with the files open
// is doing.
//
// A panel that is folded is not built at all. Building it and hiding it would
// cost a frame's work forever and would leave its rows in the window's
// accessibility tree, where a screen reader could reach what nobody can see.

// workspace is the left panel: the open files, and the way to open another.
//
// The reveal is how much of it is on screen, from 0 to 1: a fold in progress
// leaves a column that is narrower than its width, and the panel is not built at
// all once the reveal has reached nothing, so a hidden panel costs no work and
// leaves nothing in the accessibility tree.
func (r *Reader) workspace(c *ui.Context, reveal float32) {
	if reveal <= 0.002 {
		return
	}
	t := c.Theme()

	// The panel keeps its full width while it slides; only its clipping edge
	// moves. Shrinking the panel's own width would re-wrap every file name inside
	// it on every frame of the slide, which is the same mistake as letting the
	// page reflow, one order of magnitude smaller and still visible at 240 Hz.
	//
	// So the outer box is the sliding edge and the inner column is the panel at
	// rest. What the reader sees is a pane going under a fixed edge, which is what
	// a sliding panel is.
	ui.Box(c).Width(leftPanelWidth * reveal).Shrink(0).FillHeight().Clip().
		Children(func() {
			r.workspaceBody(c, t)
		})
}

// workspaceBody is the panel's content at its full width, cut by the box around
// it as it slides.
func (r *Reader) workspaceBody(c *ui.Context, t *ui.Theme) {
	// The panel starts under the header, not at the top of the window: the header
	// is an overlay, so a panel that began at 0 would have its title behind it.
	// The gap below the header is the panel's own breathing room, and it is wider
	// than the page's, because a title followed immediately by a button reads as
	// one cramped block.
	pane := ui.Column(c).Width(leftPanelWidth).Shrink(0).FillHeight().
		Padding(headerHeight+t.Space(5), t.Space(4), t.Space(4), t.Space(4)).
		Gap(t.Space(5))
	// The pane is opaque, whatever the window's material is.
	//
	// It used to be the background at half strength whenever the window had
	// vibrancy, on the theory that the material would show through and read as
	// one continuous surface. It does not: MyGo's vibrancy is a material on the
	// window, not per-pane, so a translucent pane lets the page scrolling
	// beneath it show through — legible words at the sidebar's left edge, and
	// grayscale text where the page itself gets ClearType. The chrome is a
	// surface of its own, and the palette's `chrome` is already the shade that
	// says so.
	pane.Background(palette(appearanceDark(c)).chrome)
	pane.BorderWidth(0, 1, 0, 0).BorderColor(t.Border.Alpha(0.8))

	pane.Children(func() {
		r.panelTitle(c, "Workspace", "Hide Workspace Panel", r.toggleLibrary)

		ui.Column(c).Gap(t.Space(4)).Children(func() {
			r.openFileCommand(c)
			if len(r.documents) > 0 {
				r.openFilesList(c)
			}
			// The files the app opened lately sit under the open ones, so a reader
			// who has this window on one file can still reach the others without a
			// dialog. The empty page carries the same list, with the times; here it
			// is the shorter form, because the panel is narrow and the page is not.
			r.recentFilesList(c)
		})

		ui.Spacer(c)
		r.workspaceFooter(c)
	})
}

// panelTitle is a panel's header: its name, and the command that folds it away.
//
// It is set in the chrome's small caps with wide tracking, which is what makes it
// read as a label over a group rather than as the first row of the group. The
// rule below it and the room under it are what separate it from what follows: a
// title with a button immediately beneath reads as one block, and the two are not
// the same kind of thing.
func (r *Reader) panelTitle(c *ui.Context, title, collapseLabel string, collapse func(*ui.Context)) {
	t := c.Theme()
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1)).
		Padding(0, 0, t.Space(3), 0).
		BorderWidth(0, 0, 1, 0).BorderColor(t.Border.Alpha(0.7)).
		Children(func() {
			// A panel's name is a label over a group, so it is set quieter than
			// the rows beneath it — in the chrome's own face, sentence case,
			// and without the wide tracking and small caps it used to wear.
			//
			// Mono capitals in a tracked pair read as a magazine's folio rather
			// than as an app's pane: it is a treatment for page furniture, and
			// this is a sidebar. The hierarchy is carried by colour and size,
			// which the rows below can share, so the panel reads as one list
			// with a name on it rather than as a printed page with furniture.
			ui.Text(c, title).FontSize(typeSubhead).FontWeight(600).
				TextColor(t.TextMuted).
				SingleLine().Grow(1).Label(title)
			if b := iconButton(c, iconLeft, collapseLabel, 22, 22); b.Clicked() {
				collapse(c)
			}
		})
}

// sectionLabel is a small heading over a group of rows: "Open Files", "Recent".
// It is the same treatment as a panel title, one step quieter, so the panel reads
// as a hierarchy of labels rather than as three competing headings.
func sectionLabel(c *ui.Context, label string) {
	t := c.Theme()
	ui.Text(c, label).FontSize(typeSubhead).FontWeight(500).
		TextColor(t.TextMuted).
		SingleLine().Label(label).Padding(0, t.Space(0.5))
}

// openFileCommand is the way into the app. It is a real button, not a styled
// row, so it gets the platform's hit area, focus ring, press state and Enter
// activation; and it carries its shortcut, because the menu bar shows the same
// mark and the two should agree without a sentence explaining either.
func (r *Reader) openFileCommand(c *ui.Context) {
	t := c.Theme()
	b := ui.ButtonBase(c).FillWidth().Radius(6).
		Padding(t.Space(1.5), t.Space(2)).
		Border(1, t.Border.Alpha(0.7)).
		Background(t.SurfaceHover.Alpha(0.7)).
		Label("Open a File").Tooltip("Open a File")
	if b.Hovered() {
		b.Background(t.SurfacePressed)
	}
	b.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1.5)).Children(func() {
			ui.Icon(c, iconPlus).Size(14, 14).TextColor(t.TextMuted)
			ui.Text(c, "Open a File").FontSize(typeCallout).FontWeight(500).
				TextColor(t.Text).SingleLine().Grow(1)
			ui.Text(c, "⌘O").Font("monospace").FontSize(typeFootnote).
				TextColor(t.TextMuted).SingleLine()
		})
	})
	if b.Clicked() {
		r.Open()
	}
}

// openFilesList is the files this window has open, with the active one marked.
//
// The active row is marked by its face and by the word, not by its colour alone:
// a reader who cannot tell the tint apart still reads "active".
func (r *Reader) openFilesList(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Gap(t.Space(1)).Children(func() {
		sectionLabel(c, "Open Files")
		for i, d := range r.documents {
			i, d := i, d
			active := i == r.active
			row := ui.ButtonBase(c).FillWidth().Radius(6).
				Padding(t.Space(1.5), t.Space(2)).
				Label(d.Title).Tooltip(d.Path)
			switch {
			case active:
				row.Background(t.SurfaceHover).Border(1, t.Border.Alpha(0.6))
			case row.Hovered():
				row.Background(t.SurfaceHover.Alpha(0.6))
			}
			row.Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1.5)).Children(func() {
					// The active file is marked by its row's face, which is the
					// same signal every list in this app uses. It used to carry
					// a coloured dot and the word "active" as well — three marks
					// for one fact, in a panel that holds a single file.
					ui.Text(c, d.Title).FontSize(typeCallout).
						TextColor(t.Text).SingleLine().Grow(1)
				})
			})
			if row.Clicked() && !active {
				r.selectDocument(i)
			}
		}
	})
}

// recentFilesList is the files the app opened lately, which is more use than a
// list of shortcuts a reader can read in the menu bar.
//
// Files already open in this window are left out: they are in the list above, and
// a file that appears twice in one panel is a panel that has not decided what it
// is listing.
func (r *Reader) recentFilesList(c *ui.Context) {
	t := c.Theme()
	open := map[string]bool{}
	for _, d := range r.documents {
		open[d.Path] = true
	}
	var paths []string
	for _, p := range mygo.App.RecentDocuments() {
		if !open[p] {
			paths = append(paths, p)
		}
		if len(paths) == 3 {
			break
		}
	}
	if len(paths) == 0 {
		return
	}
	ui.Column(c).Gap(t.Space(1)).Children(func() {
		sectionLabel(c, "Recent Documents")
		for _, path := range paths {
			path := path
			row := ui.ButtonBase(c).FillWidth().Radius(6).Justify(ui.Start).
				Padding(t.Space(1.5), t.Space(1.5)).
				Label(filepath.Base(path)).Tooltip(path)
			if row.Hovered() {
				row.Background(t.SurfaceHover.Alpha(0.6))
			}
			row.Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1.5)).Children(func() {
					mark := iconDocument
					if classify(path) == FormatPDF {
						mark = iconPDF
					}
					ui.Icon(c, mark).Size(13, 13).TextColor(t.TextMuted)
					ui.Text(c, filepath.Base(path)).FontSize(typeCallout).
						TextColor(t.TextMuted).SingleLine().Grow(1)
				})
			})
			if row.Clicked() {
				r.OpenIn(path)
			}
		}
	})
}

// workspaceFooter says where the open file lives, and in what encoding it was
// read: the two facts a reader checks when a document looks wrong.
func (r *Reader) workspaceFooter(c *ui.Context) {
	t := c.Theme()
	dir := ""
	if r.doc != nil {
		dir = filepath.Dir(r.doc.Path)
	}
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1)).
		Padding(t.Space(2), 0, 0, 0).
		BorderWidth(1, 0, 0, 0).BorderColor(t.Border.Alpha(0.7)).
		Children(func() {
			ui.Text(c, shortenPath(dir)).Font("monospace").FontSize(typeFootnote).
				TextColor(t.TextMuted).SingleLine().Grow(1)
			// The encoding is only worth saying when there is a file to say it
			// about; a lone "utf-8" over an empty panel is a fact about nothing.
			if r.doc != nil {
				ui.Text(c, "utf-8").Font("monospace").FontSize(typeFootnote).
					TextColor(t.TextMuted).SingleLine()
			}
		})
}

// shortenPath puts a path in the home directory back to the ~ the shell shows,
// which is shorter and reads as "yours".
func shortenPath(dir string) string {
	if dir == "" {
		return ""
	}
	home, err := mygo.App.Path(mygo.PathHome)
	if err != nil || home == "" {
		return dir
	}
	rel, err := filepath.Rel(home, dir)
	if err != nil || rel == ".." || filepath.IsAbs(rel) || len(rel) >= len(dir) {
		return dir
	}
	if rel == "." {
		return "~"
	}
	return "~/" + filepath.ToSlash(rel)
}

// outlinePane is the right panel: the document's own headings, with a track down
// the left that turns a list of words into a position.
func (r *Reader) outlinePane(c *ui.Context, reveal float32) {
	if reveal <= 0.002 {
		return
	}
	t := c.Theme()

	// The outline slides the same way the workspace does: a fixed edge over a
	// panel at its full width, so the headings do not re-wrap on the way out.
	ui.Box(c).Width(rightPanelWidth * reveal).Shrink(0).FillHeight().Clip().
		Children(func() {
			r.outlineBody(c, t)
		})
}

// outlineBody is the outline at its full width, cut by the box around it as it
// slides.
func (r *Reader) outlineBody(c *ui.Context, t *ui.Theme) {
	// The outline starts under the header, like the workspace, and keeps the same
	// breathing room below it.
	pane := ui.Column(c).Width(rightPanelWidth).Shrink(0).FillHeight().
		Padding(headerHeight+t.Space(5), t.Space(4), t.Space(4), t.Space(4)).
		Gap(t.Space(4))
	// Opaque for the same reason as the workspace: a translucent pane lets the
	// page through and costs the chrome its ClearType.
	pane.Background(palette(appearanceDark(c)).chrome)
	pane.BorderWidth(0, 0, 0, 1).BorderColor(t.Border.Alpha(0.8))

	pane.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1)).
			Padding(0, 0, t.Space(3), 0).
			BorderWidth(0, 0, 1, 0).BorderColor(t.Border.Alpha(0.7)).
			Children(func() {
				ui.Text(c, "On This Page").FontSize(typeSubhead).FontWeight(600).
					TextColor(t.TextMuted).
					SingleLine().Grow(1).Label("On This Page")
				if b := iconButton(c, iconRight, "Hide Outline Panel", 22, 22); b.Clicked() {
					r.toggleOutline(c)
				}
			})

		r.outlineList(c)

		ui.Spacer(c)
		r.outlineFooter(c)
	})
}

// outlineList is the headings, indented by depth and marked at the reader's
// place.
func (r *Reader) outlineList(c *ui.Context) {
	t := c.Theme()
	current := r.currentSection()
	r.outlineRows = r.outlineRows[:0]

	ui.Column(c).FillWidth().Gap(t.Space(0.5)).Padding(0, 0, 0, t.Space(2)).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border.Alpha(0.5)).
		Children(func() {
			for i, h := range r.doc.headings {
				i, h := i, h
				depth := h.Level - 1
				if depth < 0 {
					depth = 0
				}
				if depth > 3 {
					depth = 3
				}
				active := i == current
				row := ui.ButtonBase(c).FillWidth().Radius(4).Justify(ui.Start).
					Padding(t.Space(0.5), t.Space(1)).
					PaddingStart(float32(depth)*10 + t.Space(1)).
					Label(h.Text).Tooltip(h.Text)
				switch {
				case active:
					row.Background(t.SurfaceHover.Alpha(0.8))
				case row.Hovered():
					row.Background(t.SurfaceHover.Alpha(0.5))
				}
				row.Children(func() {
					ink := t.TextMuted
					if active {
						ink = t.Text
						// The reader's place is marked by a dot on the track as
						// well as by the row's face: a tint alone is a signal a
						// reader who cannot separate hues would miss.
						ui.Box(c).Size(5, 5).Radius(3).Shrink(0).
							Margin(0, t.Space(1), 0, 0).
							Background(t.Text.Alpha(0.75))
					}
					ui.Text(c, h.Text).FontSize(typeCallout).
						FontWeight(weightFor(active)).LineHeight(1.4).
						TextColor(ink).MaxLines(2).Grow(1)
				})
				// A heading's words appear twice on screen — in the page and
				// here — so a test that clicks by text finds the page's and
				// proves nothing about the outline. The row's own box is kept
				// for that.
				r.outlineRows = append(r.outlineRows, row.LayoutBounds())
				if row.Clicked() {
					r.jumpTo(h)
				}
			}
		})
}

// weightFor is the weight an outline row takes: the reader's place is the one
// row set in a heavier weight, which is a second signal beside the tint.
func weightFor(active bool) int {
	if active {
		return 500
	}
	return 400
}

// outlineFooter reports the reader's place in sections rather than as a
// percentage: "3 of 12" says something a reader can hold, where "42%" does not.
func (r *Reader) outlineFooter(c *ui.Context) {
	t := c.Theme()
	if r.doc == nil || len(r.doc.headings) == 0 {
		return
	}
	at := r.currentSection() + 1
	total := len(r.doc.headings)

	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(1)).
		Padding(t.Space(2), 0, 0, 0).
		BorderWidth(1, 0, 0, 0).BorderColor(t.Border.Alpha(0.7)).
		Children(func() {
			// "3 of 12" without the word "Progress" beside it: the row says
			// where the reader is, and the pill at the foot already reports the
			// same progress as a percentage. One number each, named by where it
			// sits rather than by a label it does not need.
			ui.Text(c, strconv.Itoa(at)+" of "+strconv.Itoa(total)+" sections").
				FontSize(typeFootnote).
				FontFeatures("tnum").TextColor(t.TextMuted).SingleLine()
		})
}

// currentSection is the heading the reader is in: the last one to have passed
// under the header.
//
// The line it measures against is the header's bottom edge, not the scroll
// offset, because the header floats over the page: a heading scrolled to y=2000
// is already under the bar and the reader is reading what follows it, so the
// section they are in is the one after. Measuring against the raw offset leaves
// the outline one heading behind for the whole of every section.
func (r *Reader) currentSection() int {
	if r.doc == nil || len(r.doc.headings) == 0 {
		return 0
	}
	line := r.contentScroll.Y + float32(headerHeight)
	found := 0
	for i, h := range r.doc.headings {
		if h.Offset <= line {
			found = i
		}
	}
	return found
}
