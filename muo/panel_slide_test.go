package muo

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// A slide must not reflow the page.
//
// This is the difference between an animation and a stutter. If the reading
// column's width followed the panels' animated reveal, the measure would be
// re-solved every frame and every paragraph in the document would re-break every
// frame — a document's worth of text layout between each pair of frames. On a
// 60 Hz display that reads as a slightly heavy animation; on a 240 Hz one there
// are four times as many reflows for the same gesture and it reads as jank.
//
// So the column is solved from where the panels will be, not where they are, and
// what moves during the slide is the panels' clipping edge. This test holds that
// by sampling the column mid-flight: it must already be the width it will end at.
func TestASlideDoesNotReflowThePage(t *testing.T) {
	useStateDir(t, t.TempDir())
	r := New()
	r.OpenIn(writeTemp(t, "one.md", []byte("# Title\n\ntext\n\n## Section\n\nmore\n")))

	// A window narrow enough that the column is limited by the room the panels
	// leave rather than by the character count: at a wide window the measure caps
	// the column and folding a panel changes nothing, which would make this test
	// pass without testing anything.
	tt := ui.NewTester(r.FrameView, 900, 700)
	tt.Frame()
	tt.Frame()
	withPanels := r.typ.ColumnW
	if withPanels <= 0 {
		t.Fatal("the page has no column, so this proves nothing")
	}

	// Fold both panels and sample while the slide is under way.
	r.showLibrary = false
	r.showOutline = false
	tt.Frame()
	tt.Advance(60 * time.Millisecond)

	if f := r.libReveal; f >= 0.999 {
		t.Fatalf("the panel has already arrived (reveal %.2f), so nothing was sampled mid-slide", f)
	}
	if f := r.outReveal; f >= 0.999 {
		t.Fatalf("the outline has already arrived (reveal %.2f), so nothing was sampled mid-slide", f)
	}

	midFlight := r.typ.ColumnW
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	settled := r.typ.ColumnW

	if midFlight != settled {
		t.Errorf("the column was %.1f mid-slide and %.1f at rest: the page reflowed every frame",
			midFlight, settled)
	}
	if settled <= withPanels {
		t.Errorf("the column did not widen when the panels folded: %.1f then %.1f",
			withPanels, settled)
	}
}

// The slide is still a slide: the panels' reveal travels, rather than jumping.
func TestThePanelsStillSlide(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")

	start := r.libReveal
	r.showLibrary = false
	tt.Frame()
	tt.Advance(60 * time.Millisecond)

	mid := r.libReveal
	if mid >= start {
		t.Fatalf("the panel did not begin to close: %.2f then %.2f", start, mid)
	}
	if mid <= 0 {
		t.Fatal("the panel vanished in one step instead of sliding")
	}

	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	if r.libReveal > 0.002 {
		t.Errorf("the panel is still %.3f open after the slide should have ended", r.libReveal)
	}
}
