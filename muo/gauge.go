package muo

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// The readout is the one piece of chrome that says something the window could
// not say any other way: the apparent size of the type on this screen at this
// distance, and whether it landed where reading is comfortable. It was a
// coloured string; it is a gauge now, because a position in a band is what the
// model actually computes, and the number is only that position written down.
//
// This is the app's signature and it stays singular: no second gauge, no
// second number, nothing else in the window measures anything.

const (
	// The span the gauge shows: the reading band with room either side, so a
	// size that has left the band still has somewhere to sit.
	gaugeFrom = 0.18
	gaugeTo   = 0.46

	gaugeWidth  = 58
	gaugeHeight = 4
	gaugeTick   = 11 // the marker's height, taller than the track it sits on
)

// measureGauge draws the band and the marker. The track is the whole span, the
// brighter run is the band that reads well, and the marker is where this size
// fell — so a glance says "in the band" without reading the number.
func measureGauge(c *ui.Context, typ Typography) ui.Element {
	t := c.Theme()
	box := ui.Box(c).Width(gaugeWidth).Height(gaugeHeight).Shrink(0)
	at := gaugeAt(typ.Angle)
	lo, hi := gaugeAt(smallAngle), gaugeAt(largeAngle)
	box.Draw(func(p *ui.Painter, r ui.Rect) {
		p.Fill(r, t.Text.Alpha(0.14), r.H/2)
		p.Fill(ui.Rect{X: r.X + r.W*lo, Y: r.Y, W: r.W * (hi - lo), H: r.H},
			t.Text.Alpha(0.34), r.H/2)
		p.Fill(ui.Rect{
			X: r.X + r.W*at - 1.5,
			Y: r.Y - (gaugeTick-gaugeHeight)/2,
			W: 3,
			H: gaugeTick,
		}, comfortColor(t, typ.Comfort), 1.5)
	})
	return box
}

// gaugeAt places an angle on the gauge, clamped to its ends.
func gaugeAt(angle float64) float32 {
	v := float32((angle - gaugeFrom) / (gaugeTo - gaugeFrom))
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// typeReadout is the gauge and its two numbers: the apparent capital height,
// graded, and the measure. Tabular figures, because the numbers change a step
// at a time as the reader sizes the type and must not shuffle while they do.
//
// The row is named for assistive technology rather than left as a run of
// glyphs: "0.32 degrees, dot, 69 ch" is what the text alone would be read as,
// which says nothing to someone who cannot see the gauge beside it.
func typeReadout(c *ui.Context, typ Typography) {
	t := c.Theme()
	grade := "in the comfortable band"
	switch typ.Comfort {
	case ComfortSmall:
		grade = "smaller than is comfortable"
	case ComfortLarge:
		grade = "larger than is comfortable"
	}
	row := ui.Row(c).Gap(t.Space(1.5)).AlignItems(ui.Center).Shrink(0).
		Label("Type size").
		Description(fmt.Sprintf(
			"Apparent character height %.2f degrees, %s; %.0f characters per line.",
			typ.Angle, grade, typ.ChPerLine))
	row.Children(func() {
		measureGauge(c, typ)
		ui.Text(c, fmt.Sprintf("%.2f°", typ.Angle)).
			FontSize(typeSubhead).FontFeatures("tnum").
			TextColor(comfortColor(t, typ.Comfort)).SingleLine()
		ui.Text(c, "·").FontSize(typeSubhead).
			TextColor(t.TextMuted.Alpha(0.55)).SingleLine()
		ui.Text(c, fmt.Sprintf("%.0fch", typ.ChPerLine)).
			FontSize(typeSubhead).FontFeatures("tnum").
			TextColor(t.TextMuted).SingleLine()
	})
}
