package muo

import (
	"github.com/egoist/mygo/ui"
)

// The window's footer: a quiet strip at the foot of the page, saying what the
// window is doing and what the keys are.
//
// It is the one band that is not the page, and it earns its room only because it
// is the place a reader looks when they want to know something about the app
// rather than the document: that the reader is ready, where the shortcuts are,
// and — when a document is open — how far through it they are. It is drawn in
// the chrome's mono at the smallest size the window uses, so it reads as a
// caption rather than as content.
//
// It is hidden in Zen: a reading mode that leaves a status strip at the foot of
// the page is not a reading mode.

func (r *Reader) footer(c *ui.Context) {
	t := c.Theme()

	row := ui.Row(c).FillWidth().Height(40).Shrink(0).
		Padding(0, t.Space(5)).AlignItems(ui.Center).Gap(t.Space(4))
	row.BorderWidth(1, 0, 0, 0).BorderColor(t.Border.Alpha(0.5))
	// The footer is a surface for the same reason the header is: it is in the
	// layout, so a material under it would be frosting the window's own paper.
	row.Background(t.Surface)

	row.Children(func() {
		// Left: the engine's own state. It is a dot and a sentence, because the
		// one thing a reader wants from a status strip is "is anything wrong",
		// and a green dot answers that before the sentence does.
		ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Shrink(0).Children(func() {
			ui.Box(c).Size(6, 6).Radius(3).Shrink(0).
				Background(t.Success.Alpha(0.7))
			ui.Text(c, "Typography engine ready").Font("monospace").
				FontSize(typeFootnote).TextColor(t.TextMuted).SingleLine()
		})

		ui.Spacer(c)

		// Centre: three dots, the editorial mark that says the window is a page.
		// It is drawn only on the empty page, where there is no page footer of its
		// own to close; a document draws its own end mark instead.
		if r.doc == nil {
			ui.Row(c).Gap(t.Space(1)).AlignItems(ui.Center).Shrink(0).
				Label("Muo").Children(func() {
				for i := 0; i < 3; i++ {
					ui.Box(c).Size(4, 4).Radius(2).Background(t.TextMuted.Alpha(0.35))
				}
			})
			ui.Spacer(c)
		}

		// Right: the shortcut on the empty page; nothing on a document, because
		// the pill at the foot of the page already carries the reader's progress
		// and two copies of one number is a window that does not trust its own
		// display.
		if r.doc != nil {
			return
		}
		ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).Shrink(0).Children(func() {
			ui.Text(c, "Press").Font("monospace").FontSize(typeFootnote).
				TextColor(t.TextMuted).SingleLine()
			ui.Row(c).Padding(1, 5).Radius(4).
				Border(1, t.Border.Alpha(0.8)).Children(func() {
				ui.Text(c, "?").Font("monospace").FontSize(typeFootnote).
					TextColor(t.TextMuted).SingleLine()
			})
			ui.Text(c, "for shortcuts").Font("monospace").FontSize(typeFootnote).
				TextColor(t.TextMuted).SingleLine()
		})
	})
}
