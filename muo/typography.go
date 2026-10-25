package muo

import (
	"math"
	"strings"

	"github.com/egoist/mygo/ui"
)

// Measuring. The type model needs one number from the text engine — the
// advance of a character at a size — and it must come from the same shaper the
// reader draws with, not from a table of font metrics that may be a different
// font than the system's.

// Advance measures one character's width in DIPs for a family and size.
func Advance(family string, size float32) float32 {
	if size < 1 {
		return size * 0.5
	}
	w, _ := ui.ShapeText("0123456789", ui.Font{Family: family, Size: size}, 1e6).Size()
	if w <= 0.1 {
		return size * 0.5
	}
	return w / 10
}

// proportionalFactor is how much wider an average prose character is than a
// digit. Digits are the narrowest common glyph in most reading faces; prose in
// Latin scripts sits about ten percent above them. Measured across Inter and
// the system UI faces at reading sizes.
const proportionalFactor = 1.10

// isMonospace reports whether a family is fixed pitch, by measuring two strings
// of equal length made of narrow and wide glyphs.
func isMonospace(family string, size float32) bool {
	if family == "" || strings.Contains(family, "monospace") {
		return false
	}
	a, _ := ui.ShapeText("iiiiiiii", ui.Font{Family: family, Size: size}, 1e6).Size()
	b, _ := ui.ShapeText("MMMMMMMM", ui.Font{Family: family, Size: size}, 1e6).Size()
	return a > 0 && b > 0 && math.Abs(float64(a-b)) < 0.5
}

// Model turns a display's calibration into the reader's type size and column.
//
// Two measured rules do the work:
//
//   - the character height stays at the reading angle (0.21°), whatever the
//     screen, so a 27" monitor at 65 cm and a 16" laptop at 42 cm get different
//     pixel sizes for the same apparent comfort;
//   - the column stays at 72 characters, so the measure never changes — only
//     the physical size of the page does.
func Model(d Display, viewportPx, ppi float64, family string, manual float32) Typography {
	distance := d.Distance
	if distance <= 0 {
		distance = 45
	}
	if ppi <= 0 {
		ppi = 96
	}
	size := comfortSize(ppi, distance)
	if manual > 0 {
		size = manual
	}
	angle := ComfortAngle(size, ppi, distance)

	t := Typography{
		FontPx:        size,
		Line:          readingLineHeight(size),
		Angle:         angle,
		Comfort:       grade(angle),
		Distance:      distance,
		EffectivePPI:  ppi,
		PreferredSize: comfortSize(ppi, distance),
	}
	t.Ch = Advance(family, size)
	if !isMonospace(family, size) {
		t.Ch *= proportionalFactor
	}
	if t.Ch <= 0 {
		t.Ch = size * 0.5
	}
	t.ColumnW, t.ChPerLine = column(float64(size), float64(t.Ch), viewportPx)
	return t
}
// capHeightRatio is the fraction of the em that a reading face's capitals
// occupy. What the eye recognises as a character is that height, not the em
// box, so it is what the angle describes.
const capHeightRatio = 0.70

// ComfortAngle reports the apparent character height, in degrees, for a type
// size at a density and a viewing distance.
func ComfortAngle(fontPx float32, ppi, distanceCm float64) float64 {
	if fontPx <= 0 || ppi <= 0 || distanceCm <= 0 {
		return 0
	}
	// The capitals' height in centimetres, from pixels at the display's density.
	heightCm := float64(fontPx*capHeightRatio) / ppi * 2.54
	return angleOf(heightCm, distanceCm)
}

// angleOf is the angle a height subtends at a distance, both in centimetres.
func angleOf(heightCm, distanceCm float64) float64 {
	if heightCm <= 0 || distanceCm <= 0 {
		return 0
	}
	return 2 * math.Atan(heightCm/2/distanceCm) * rad2deg
}

const rad2deg = 57.29577951308232

// comfortSize solves the angle back to a type size in pixels: the inverse of
// ComfortAngle, so that the two can never disagree.
func comfortSize(ppi, distanceCm float64) float32 {
	heightCm := 2 * math.Tan(targetAngle/2/rad2deg) * distanceCm
	return float32(heightCm / capHeightRatio / 2.54 * ppi)
}

// grade buckets the apparent size into the reading band.
func grade(angle float64) Comfort {
	switch {
	case angle < smallAngle:
		return ComfortSmall
	case angle > largeAngle:
		return ComfortLarge
	}
	return ComfortGood
}

// column picks the widest 72-character column that fits the viewport with a
// margin to spare. When the window is too narrow for 72 characters the column
// is capped by the space, never by a pixel width: a pixel cap would change the
// character count whenever the type size changed, and so the measure on one
// machine only.
//
// The margin is 0.78 of the viewport, which leaves room for the page's own
// insets. On a dense laptop panel the type is large enough that 72 characters
// need more than that margin, and the measure gave way to 59.7 characters —
// which is the failure this function exists to prevent. So the margin is what
// gives way instead, down to the full viewport, and the measure holds.
func column(fontPx, ch, viewportPx float64) (width, chars float32) {
	ideal := ch * charsPerLine
	if ideal <= viewportPx*0.78 {
		return float32(ideal), float32(charsPerLine)
	}
	// The column wants more than the comfortable margin. It takes the room up to
	// the viewport's own width before it gives up a character, because the
	// measure is the promise and the margin is the preference.
	if ideal <= viewportPx {
		return float32(ideal), float32(charsPerLine)
	}
	// The window genuinely cannot hold 72 characters at this size: the count
	// gives way, and it is reported honestly so the readout says so.
	return float32(viewportPx), float32(viewportPx / ch)
}
