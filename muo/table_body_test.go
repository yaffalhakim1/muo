package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// Every cell of the sample's table must stay on one line: the renderer sizes a
// column from its widest cell, and a cell that wraps means the cap ate into the
// text. The header test only checks the header row; the body rows carry the hex
// codes and ratios, which are the long ones.
func TestTableBodyCellsDoNotBreakWords(t *testing.T) {
	src := "| Surface | Background | Ink | Ratio |\n| --- | --- | --- | --- |\n" +
		"| Light | #faf8f5 | #1f1d1b | 13:1 |\n| Dark | #1a1a1c | #cdc8c2 | 9:1 |\n"

	// The width the page is actually read at: the live window on a 27" is
	// wider than this, but the column is capped there too, so the test is the
	// column's width that matters, not the window's.
	for _, width := range []int{760, 900, 1100, 1440} {
		r := New()
		r.doc = &Document{Format: FormatMarkdown, Source: src, Title: "t.md"}
		tt := ui.NewTester(r.FrameView, width, 700)
		tt.Frame()

		for _, want := range []string{"#faf8f5", "#1f1d1b", "13:1", "#cdc8c2"} {
			rect, ok := tt.Find(want)
			if !ok {
				t.Errorf("width %d: the cell %q is not on the page; the column wrapped it apart", width, want)
				continue
			}
			// One line of body text is around 25 DIP; two would be near 50.
			if rect.H > 34 {
				t.Errorf("width %d: the cell %q is %.1f DIP tall, so it wrapped", width, want, rect.H)
			}
		}
	}
}
