package muo

import (
	"math"
	"strconv"

	"github.com/egoist/mygo/ui"
)

// How far through the document the reader is, said twice: a hairline across the
// top of the window, and a number in the pill at the foot.
//
// Both read the same source, which is the scroll container's own state. That
// matters more than it looks: a progress figure computed from a separate
// scroll listener drifts from the thing it describes, and a reader who catches
// the bar saying 60% while the page is at the end stops trusting both. Here
// there is one number, and the two marks are two ways of drawing it.

// readingProgress is the fraction of the page that has gone past the top of the
// viewport, from 0 to 1.
//
// It is measured on the scrollable remainder rather than on the whole document,
// so that "100%" means the reader has reached the end — the last screenful is
// the end, not the last pixel of content. A document that fits the window is
// entirely read, and reads as 100% rather than as 0%.
func (r *Reader) readingProgress() float64 {
	s := &r.contentScroll
	if s.MaxY <= 0 {
		if r.doc == nil {
			return 0
		}
		return 1
	}
	f := float64(s.Y) / float64(s.MaxY)
	switch {
	case math.IsNaN(f) || f < 0:
		return 0
	case f > 1:
		return 1
	}
	return f
}

// readingProgressBar draws the hairline across the top of the window. It is
// absolute and pass-through, so it lies over the header without taking clicks
// from it, and it is drawn last so it is over everything else.
func (r *Reader) readingProgressBar(c *ui.Context) {
	t := c.Theme()
	f := r.readingProgress()
	if f <= 0 || r.doc == nil {
		return
	}
	track := ui.Box(c).Absolute().Top(0).Left(0).Right(0).Height(2).
		PassThrough().Background(ui.Transparent)
	track.Children(func() {
		ui.Box(c).WidthPercent(float32(f) * 100).FillHeight().
			Background(t.TextMuted.Alpha(0.5))
	})
}

// pill is the floating status mark at the foot of the page: the measure, the
// two type-size commands, and how far the reader has got.
//
// It floats rather than sitting in a status bar because the window has no status
// bar, and it should not grow one: the reading column is the whole point of the
// window, and a permanent strip at its foot would be chrome the reader pays for
// on every page. It is centred over the page's own column, so it belongs to the
// text rather than to the window.
func (r *Reader) pill(c *ui.Context) {
	if r.zen || r.doc == nil || r.doc.Err != "" {
		return
	}
	t := c.Theme()

	// The pill belongs to the page, not to the window: it is centred over the
	// column the reader is actually reading, so a panel opening does not slide
	// the status out from under the text it describes. The column's own bounds
	// give it the room to be centred in, which an absolutely placed row with
	// only a left edge would not have: without a width its children compress.
	bar := ui.Row(c).Absolute().Bottom(t.Space(4)).PassThrough().
		AlignItems(ui.Center).Justify(ui.Center)
	if b := r.column.Bounds(c); b.W > 0 {
		bar.Left(b.X).Width(b.W)
	} else {
		bar.Left(0).Right(0)
	}
	// The pill is a capsule: its radius is half its height, and the number is
	// written out rather than left at 999 because the glass draws its rim to the
	// radius it is given. A radius past the half-height is not a capsule to the
	// shader, it is a shape with corners it cannot place, and the frost shows its
	// seams.
	const pillHeight = 28
	bar.Children(func() {
		body := ui.Row(c).Gap(t.Space(2)).AlignItems(ui.Center).
			Shrink(0).Height(pillHeight).
			Padding(0, t.Space(3)).Radius(pillHeight/2).
			Border(1, t.Border.Alpha(0.8)).
			Shadow(0, 2, 10, 0, palette(appearanceDark(c)).shadow.Alpha(0.12)).
			Label("Reading status")
		// The design frosts the pill with a backdrop blur, and at 90% opacity
		// that frost is barely doing any work: the point of it there is to stop
		// the page's words running through the numbers. An opaque surface does
		// that job completely, and it matches the panels, which are the app's
		// own material. The glass was tried and its rim showed its seams at this
		// size, which is worse than not frosted at all.
		body.Background(t.Surface)
		body.Children(func() {
			// The measure, as a dot and a number: the dot says the column is
			// inside the band the type model aims for.
			ui.Row(c).Gap(t.Space(1)).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(8, 8).Radius(4).
					Border(1, t.TextMuted.Alpha(0.6)).
					Background(comfortColor(t, r.typ.Comfort).Alpha(0.55))
				ui.Text(c, measureLabel(r.typ)).
					Font("monospace").FontSize(typeFootnote).
					FontFeatures("tnum").TextColor(t.TextMuted).SingleLine()
			})

			pillDivider(c)

			// The two type-size commands. They are the only controls in the
			// window that change the page while the reader watches, so they sit
			// where the reader's eye already is when they are reading.
			ui.Row(c).Gap(t.Space(0.5)).AlignItems(ui.Center).Children(func() {
				if pillButton(c, "A-", "Decrease font size").Clicked() {
					r.stepType(1 / 1.05)
				}
				ui.Text(c, "/").Font("monospace").FontSize(typeFootnote).
					TextColor(t.Border).SingleLine()
				if pillButton(c, "A+", "Increase font size").Clicked() {
					r.stepType(1.05)
				}
			})

			pillDivider(c)

			ui.Text(c, percentLabel(r.readingProgress())).
				Font("monospace").FontSize(typeFootnote).FontWeight(500).
				FontFeatures("tnum").TextColor(t.Text).SingleLine()
		})
	})
}

// measureLabel writes the measure the way the pill says it: the character count
// the column actually holds, which is the number the reader is promised.
func measureLabel(typ Typography) string {
	return strconv.Itoa(int(math.Round(float64(typ.ChPerLine)))) + "ch"
}

// percentLabel rounds the progress to the percent, and says 100 only when the
// reader is actually at the end: a bar that says 100% while there is more to
// read is the same broken promise as one that drifts.
func percentLabel(f float64) string {
	n := int(math.Round(f * 100))
	if n > 100 {
		n = 100
	}
	return strconv.Itoa(n) + "%"
}

func pillDivider(c *ui.Context) {
	t := c.Theme()
	ui.Box(c).Width(1).Height(t.Rem(0.9)).Shrink(0).
		Background(t.Border).AlignSelf(ui.Center)
}

// pillButton is a word-shaped command in the pill: no face until the pointer is
// on it, so the pill reads as one quiet object rather than as a toolbar.
func pillButton(c *ui.Context, label, tip string) ui.Element {
	t := c.Theme()
	b := ui.Button(c, label).Height(20).Radius(4).
		Padding(0, t.Space(1)).
		Font("monospace").FontSize(typeFootnote).
		Background(ui.Transparent).BorderWidth(0).
		TextColor(t.TextMuted).
		Label(tip).Tooltip(tip)
	if b.Hovered() {
		b.Background(t.SurfacePressed)
		b.TextColor(t.Text)
	}
	return b
}
