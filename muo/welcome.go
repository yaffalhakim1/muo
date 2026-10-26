package muo

import (
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The empty window is the app's frontispiece: the one place it can say what it
// is for before there is anything to read.
//
// It is a centred page, not a form: a wordmark, one italic sentence, one action,
// and the files the reader was last in. Everything is centred and everything is
// set in the page's own serif, because the first thing a reader sees should
// already be the page they came for.
//
// The window keeps its header, its panels and its footer here, so the empty state
// is a page with nothing in it rather than a different screen. A reader who opens
// a file sees the same window with the page filled in.

// welcomeMeasure is the width the invitation is set to: narrow enough that the
// tagline breaks into two lines at a comfortable measure, as the design sets it.
const welcomeMeasure = 460

func (r *Reader) welcome(c *ui.Context, pal pagePalette) {
	t := c.Theme()
	root := ui.Column(c).Fill().Center().Padding(pageMargin)
	if root.FileDragOver() {
		// The page itself says it will take the file, rather than a dashed
		// target appearing somewhere inside it.
		root.Background(t.SurfaceHover.Alpha(0.5))
	}
	root.Children(func() {
		ui.Column(c).Width(welcomeMeasure).MaxWidthPercent(100).
			AlignItems(ui.Center).Children(func() {
			// The wordmark is set large, in the page's serif, at the weight a
			// title takes: the name of the thing, before any explanation of it.
			ui.Text(c, "Muo").Font(pageFamily).FontSize(52).FontWeight(400).
				LetterSpacing(-1.2).TextColor(pal.inkColor()).
				TextAlign(ui.Center)

			// The tagline is the design's own sentence, in italic: it is a
			// claim about the app rather than a label, and the italic is what
			// separates a claim from an instruction.
			ui.Text(c, "A page sized to the screen you are on, and to how far you sit from it.").
				Font(pageFamily).FontSize(19).Italic().LineHeight(1.55).
				TextColor(pal.mutedColor()).MaxWidth(welcomeMeasure - 60).
				TextAlign(ui.Center).Margin(0, 0, t.Space(6), 0)

			r.openCommand(c)
			r.welcomeRecent(c, pal)
		})
	})
}

// openCommand is the one action on the empty page: a filled control carrying its
// shortcut inside it.
//
// The fill is the page's ink on its paper, not an accent colour. The design has
// no accent at all — its loudest surface is the text colour itself — and a blue
// button on a page of warm cream is the one thing that would look borrowed from
// another app.
func (r *Reader) openCommand(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).AlignItems(ui.Center).Gap(t.Space(3)).Children(func() {
		b := ui.Button(c, "").Radius(8).
			Padding(t.Space(2), t.Space(3)).
			Background(t.Inverse).BorderWidth(0).
			Label("Open a File").Tooltip("Open a File")
		if b.Hovered() {
			b.Background(t.Inverse.Mix(t.Background, 0.18))
		}
		b.Children(func() {
			ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Open a File").FontSize(typeBody+1).FontWeight(500).
					TextColor(t.InverseText).SingleLine()
				// The shortcut sits inside the button, on a slightly darker
				// well, so it reads as the key to this control rather than as a
				// second thing beside it.
				ui.Row(c).Padding(1, t.Space(1)).Radius(4).
					Background(t.InverseText.Alpha(0.16)).Children(func() {
					ui.Text(c, "⌘O").Font("monospace").FontSize(typeFootnote).
						TextColor(t.InverseText.Alpha(0.9)).SingleLine()
				})
			})
		})
		if b.Clicked() {
			r.Open()
		}

		ui.Text(c, "or drop any Markdown file anywhere in this window").
			FontSize(typeCallout).TextColor(t.TextMuted).TextAlign(ui.Center)
	})
}

// welcomeRecent is the recent list under the empty page's rule: the files the
// app opened lately, newest first.
//
// It is the one part of the empty state that does something the menu bar cannot:
// choosing a row opens it here, with no dialog in the way. It is deliberately not
// the same list the workspace panel shows — that one is the files this window has
// open, which is empty here, so the panel's recent list is the only copy.
func (r *Reader) welcomeRecent(c *ui.Context, pal pagePalette) {
	t := c.Theme()
	paths := mygo.App.RecentDocuments()
	if len(paths) > 3 {
		paths = paths[:3]
	}
	if len(paths) == 0 {
		return
	}

	ui.Box(c).Height(1).FillWidth().Background(t.Border.Alpha(0.5)).
		Margin(t.Space(6), 0, t.Space(4), 0).MaxWidth(420)

	ui.Column(c).Width(420).MaxWidthPercent(100).Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Padding(0, t.Space(1), t.Space(2), t.Space(1)).
			Children(func() {
			ui.Text(c, "Recent Documents").Font("monospace").
				FontSize(typeSubhead).LetterSpacing(0.6).
				TextColor(t.TextMuted).SingleLine().Grow(1)
			if clear := ui.Button(c, "Clear").FontSize(typeSubhead).
				Padding(0).Background(ui.Transparent).BorderWidth(0).
				TextColor(t.TextMuted); clear.Clicked() {
				_ = mygo.App.ClearRecentDocuments()
			}
		})

		for _, path := range paths {
			path := path
			row := ui.ButtonBase(c).FillWidth().Radius(6).Justify(ui.Start).
				Padding(t.Space(2), t.Space(1.5)).
				Label(filepath.Base(path)).Tooltip(path)
			if row.Hovered() {
				row.Background(t.SurfaceHover.Alpha(0.7))
			}
			row.Children(func() {
				ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(t.Space(2)).Children(func() {
					ui.Icon(c, iconDocument).Size(15, 15).TextColor(t.TextMuted)
					ui.Column(c).Grow(1).MinWidth(0).Children(func() {
						ui.Text(c, filepath.Base(path)).Font("monospace").
							FontSize(typeCallout).TextColor(pal.inkColor()).SingleLine()
						ui.Text(c, shortenPath(filepath.Dir(path))).Font("monospace").
							FontSize(typeFootnote).TextColor(t.TextMuted).SingleLine()
					})
					if when, ok := recentOpenedAt(path); ok {
						ui.Text(c, agoLabel(when)).Font("monospace").
							FontSize(typeFootnote).TextColor(t.TextMuted).SingleLine()
					}
				})
			})
			if row.Clicked() {
				r.OpenIn(path)
			}
		}
	})
}

// keycap draws a shortcut as the key it is pressed with. It is the same mark the
// menus show, so the button beside it and the menu bar agree without a sentence
// explaining either.
func keycap(c *ui.Context, label string) {
	t := c.Theme()
	ui.Row(c).Shrink(0).Padding(3, 6).Radius(5).
		Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, label).FontSize(typeFootnote).TextColor(t.TextMuted)
	})
}
