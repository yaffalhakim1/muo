package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// A table's header must not break words. The renderer sizes a column from its
// widest cell, capped at TableMaxColumn ems, and adds the cell padding outside
// that cap; when the cap is too tight the padding eats into the text and a word
// wraps mid-word ("Background" into "Backgroun/d").
func TestTableHeaderCellsDoNotBreakWords(t *testing.T) {
	src := "| Surface | Background | Ink | Ratio |\n| --- | --- | --- | --- |\n" +
		"| Light | #faf8f5 | #1f1d1b | 13:1 |\n| Dark | #1a1a1c | #cdc8c2 | 9:1 |\n"

	r := New()
	r.doc = &Document{Format: FormatMarkdown, Source: src, Title: "t.md"}
	tt := ui.NewTester(r.FrameView, 900, 600)
	tt.Frame()

	for _, want := range []string{"Surface", "Background", "Ink", "Ratio"} {
		rect, ok := tt.Find(want)
		if !ok {
			t.Fatalf("the header cell %q is not on the page; the column wrapped it apart", want)
		}
		t.Logf("%-12s box: x=%.1f w=%.1f h=%.1f", want, rect.X, rect.W, rect.H)
	}
}

