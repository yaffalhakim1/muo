package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// Where the page's own paragraph actually lands, and how wide it is. This uses
// the tester's geometry rather than pixels, so it cannot be fooled by the
// scroll bar or the title row.
func TestParagraphWidthInPage(t *testing.T) {
	const para = "A reader is a machine for one job: getting words from a screen into a head with as little friction as possible, and nothing else besides that."
	src := "# Heading\n\n" + para + "\n"
	r := New()
	r.OpenIn(writeTemp(t, "p.md", []byte(src)))
	r.SetSidebarWidth(sidebarWidth)

	const winW, winH = 884, 700
	tt := ui.NewTester(r.FrameView, winW, winH)
	tt.Frame()
	tt.Frame()

	rect, ok := tt.Find(para)
	if !ok {
		t.Fatal("the paragraph was not found")
	}
	t.Logf("pane starts at %d, is %d wide", sidebarWidth, winW-sidebarWidth)
	t.Logf("paragraph: x=%.0f w=%.0f right=%.0f", rect.X, rect.W, rect.X+rect.W)
	t.Logf("model ColumnW=%.0f (%.1fch)", r.typ.ColumnW, r.typ.ChPerLine)

	if rect.W > r.typ.ColumnW+12 {
		t.Fatalf("the paragraph is %.0f wide but the column is %.0f", rect.W, r.typ.ColumnW)
	}
}

// Nothing the page draws may reach the window's right edge: the measure is a
// promise, and the scroll bar needs its room.
func TestPageStaysInsideWindow(t *testing.T) {
	const para = "A reader is a machine for one job: getting words from a screen into a head with as little friction as possible, and nothing else besides that."
	r := New()
	r.OpenIn(writeTemp(t, "edge.md", []byte("# H\n\n"+para+"\n")))
	r.SetSidebarWidth(sidebarWidth)

	const winW, winH = 884, 700
	tt := ui.NewTester(r.FrameView, winW, winH)
	tt.Frame()
	tt.Frame()

	rect, ok := tt.Find(para)
	if !ok {
		t.Fatal("the paragraph was not found")
	}
	if right := rect.X + rect.W; right > float32(winW)-16 {
		t.Fatalf("the paragraph ends at %.0f in a %d window", right, winW)
	}
}
