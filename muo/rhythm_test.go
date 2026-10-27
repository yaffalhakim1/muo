package muo

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"
)

// A heading must sit closer to the paragraph it introduces than to the one above
// it. That is what makes a page read as sections rather than as a list, and the
// renderer does not enforce it: it takes the *maximum* of two adjacent block
// margins, so a heading's Bottom is silently discarded whenever it is smaller
// than the block spacing. The old theme asked for 0.55em below an h2 and got
// 1.25em, which made every heading equidistant from the text on both sides.
//
// This measures the rendered gaps in pixels, so it fails on the outcome rather
// than on the numbers that produce it.
func TestHeadingSitsCloserToWhatFollowsIt(t *testing.T) {
	src := "# Title\n\nfirst paragraph\n\n## Section\n\nsecond paragraph\n"
	r := openString(t, "rhythm.md", src)
	tt := ui.NewTester(r.FrameView, 900, 900)
	tt.Frame()
	tt.Frame()

	// The column's rect is measured by the frame itself; the scan is bounded by
	// it so that chrome is never read as text.
	col := r.lastColumn
	gaps := textGaps(tt.Image(), col)
	if len(gaps) < 5 {
		t.Fatalf("expected at least five gaps between lines of text, got %d: %v", len(gaps), gaps)
	}

	// The page draws, in order: title, paragraph 1, heading, paragraph 2. The
	// heading is the third block, so the gap above it is gaps[2] and the gap
	// below it is gaps[3].
	above, below := gaps[2], gaps[3]
	t.Logf("gaps between lines: %v", gaps)
	t.Logf("air above the section heading: %d px; below it: %d px", above, below)

	if below >= above {
		t.Errorf("a heading is %d px from the text above it and %d px from the text below it; "+
			"the air below must be the smaller of the two", above, below)
	}
}

// textGaps returns the vertical gaps between consecutive lines of ink inside the
// reading column, in pixels.
func textGaps(img *image.RGBA, col ui.Rect) []int {
	b := img.Bounds()
	if col.W <= 0 {
		return nil
	}
	// Sample the paper in the top margin of the page, above the first line.
	paper := img.RGBAAt(int(col.X+col.W/2), int(headerHeight)+4)

	left, right := int(col.X), int(col.X+col.W)
	if left < 0 {
		left = 0
	}
	if right > b.Dx() {
		right = b.Dx()
	}
	if right <= left {
		return nil
	}

	// A row has ink if any pixel across the column differs from the paper.
	inked := make([]bool, b.Dy())
	for y := int(headerHeight) + pageMargin; y < b.Dy(); y++ {
		for x := left; x < right; x++ {
			if !closeColor(img.RGBAAt(x, y), paper, 12) {
				inked[y] = true
				break
			}
		}
	}

	var gaps []int
	lastInk := -1
	for y, on := range inked {
		if !on {
			continue
		}
		if lastInk >= 0 && y-lastInk > 1 {
			gaps = append(gaps, y-lastInk-1)
		}
		lastInk = y
	}
	return gaps
}