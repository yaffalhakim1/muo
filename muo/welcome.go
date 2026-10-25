package muo

import (
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The empty window is the app's frontispiece: the one place it can say what it
// is for before there is anything to read.
//
// It is set at the reading column's own measure and centred in the page, so the
// first thing a reader sees is already a page rather than a form. What it says
// is one sentence, one action, and — if the reader has been here before — the
// files they were last in, which is more use than a list of shortcuts they can
// read in the menu bar.

// welcomeMeasure is the width the invitation is set to: the reading column's
// own, so the empty state and the page are the same shape.
const welcomeMeasure = 400

func (r *Reader) welcome(c *ui.Context, pal pagePalette) {
	t := c.Theme()
	root := ui.Column(c).Fill().Center().Padding(pageMargin)
	if root.FileDragOver() {
		// The page itself says it will take the file, rather than a dashed
		// target appearing somewhere inside it.
		root.Background(t.Accent.Alpha(0.05))
	}
	root.Children(func() {
		ui.Column(c).Width(welcomeMeasure).MaxWidthPercent(100).
			Gap(t.Space(3)).Children(func() {
			ui.Text(c, "Muo").FontSize(typeTitle1).FontWeight(600).
				LetterSpacing(-0.4).TextColor(pal.inkColor())
			ui.Text(c, "A page sized to the screen you are on, and to how far you sit from it.").
				FontSize(typeBody+1).LineHeight(1.55).
				TextColor(pal.mutedColor()).MaxWidth(welcomeMeasure-40)
			ui.Row(c).Gap(t.Space(3)).AlignItems(ui.Center).
				Padding(t.Space(1), 0, 0, 0).Children(func() {
				if ui.PrimaryButton(c, "Open a File").Clicked() {
					r.Open()
				}
				keycap(c, "⌘O")
			})
			ui.Text(c, "or drop a file anywhere in this window").
				FontSize(typeCallout).TextColor(pal.mutedColor())
			r.recentDocuments(c, pal)
		})
	})
}

// recentDocuments lists the files the app opened lately, newest first. It is
// the one part of the empty state that does something the menu bar cannot:
// choosing a row opens it here, with no dialog in the way.
func (r *Reader) recentDocuments(c *ui.Context, pal pagePalette) {
	t := c.Theme()
	paths := mygo.App.RecentDocuments()
	if len(paths) > 4 {
		paths = paths[:4]
	}
	if len(paths) == 0 {
		return
	}
	ui.Box(c).Height(1).FillWidth().Background(t.Border).
		Margin(t.Space(3), 0, 0, 0)
	ui.Text(c, "Recent").FontSize(typeSubhead).FontWeight(600).
		LetterSpacing(0.3).TextColor(pal.mutedColor()).
		Padding(0, t.Space(1)).Margin(t.Space(1), 0, 0, 0)
	for _, path := range paths {
		path := path
		row := ui.ButtonBase(c).FillWidth().Radius(6).
			Padding(t.Space(1.5), t.Space(2)).
			Label(filepath.Base(path)).Tooltip(path)
		if row.Hovered() {
			row.Background(t.SurfaceHover)
		}
		row.Children(func() {
			icon := iconDocument
			if classify(path) == FormatPDF {
				icon = iconPDF
			}
			ui.Icon(c, icon).Size(14, 14).TextColor(pal.mutedColor())
			ui.Text(c, filepath.Base(path)).FontSize(typeBody).
				TextColor(pal.inkColor()).SingleLine().Grow(1)
			ui.Text(c, filepath.Dir(path)).FontSize(typeFootnote).
				TextColor(pal.mutedColor()).SingleLine().MaxWidth(150).
				TextAlign(ui.End)
		})
		if row.Clicked() {
			r.OpenIn(path)
		}
	}
}

// keycap draws a shortcut as the key it is pressed with. It is the same mark
// the menus show, so the button beside it and the menu bar agree without a
// sentence explaining either.
func keycap(c *ui.Context, label string) {
	t := c.Theme()
	ui.Row(c).Shrink(0).Padding(3, 6).Radius(5).
		Background(t.Surface).Border(1, t.Border).Children(func() {
		ui.Text(c, label).FontSize(typeFootnote).TextColor(t.TextMuted)
	})
}
