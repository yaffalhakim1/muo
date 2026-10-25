package muo

import (
	"image/png"
	"os"
	"testing"

	"github.com/egoist/mygo/ui"
)

// A narrow window is where the table used to break: the column cap and the
// window fight, and a word wraps mid-word instead of the table scrolling.
func TestTableFitsANarrowWindow(t *testing.T) {
	src := "| Surface | Background | Ink | Ratio |\n| --- | --- | --- | --- |\n" +
		"| Light | #faf8f5 | #1f1d1b | 13:1 |\n| Dark | #1a1a1c | #cdc8c2 | 9:1 |\n"

	r := New()
	r.doc = &Document{Format: FormatMarkdown, Source: src, Title: "t.md"}
	tt := ui.NewTester(r.FrameView, 560, 500)
	tt.Frame()

	for _, want := range []string{"Surface", "Background", "Ink", "Ratio"} {
		if _, ok := tt.Find(want); !ok {
			t.Errorf("%q wrapped at a 560 pt window", want)
		}
	}
	f, _ := os.Create("Y:/Learn/Go/muo/capture-narrow.png")
	defer f.Close()
	png.Encode(f, tt.Image())
}

