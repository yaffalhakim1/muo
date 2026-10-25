package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// Is a Width on a Column inside a padded Row honoured? The column's text must
// wrap inside the width it was given, so its ink never reaches the padding.
func TestWidthInPaddedRow(t *testing.T) {
	const want = 518
	const text = "The quick brown fox jumps over the lazy dog and keeps on running past the edge of any narrow column you give it, so that the line must wrap somewhere."
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Scroll(c).Fill().Children(func() {
			ui.Row(c).FillWidth().Padding(28, 32, 72, 32).Children(func() {
				ui.Spacer(c)
				ui.Column(c).Width(want).MaxWidthPercent(100).Children(func() {
					ui.Text(c, text)
				})
				ui.Spacer(c)
			})
		})
	}, 884, 700)
	tt.Frame()
	tt.Frame()

	// The text's own box, from the tester's own geometry rather than pixels.
	rect, ok := tt.Find(text)
	if !ok {
		t.Fatal("the text was not found")
	}
	t.Logf("text box: x=%.0f w=%.0f right=%.0f (column given %d)", rect.X, rect.W, rect.X+rect.W, want)
	if rect.W > want+12 {
		t.Fatalf("the text is %.0f wide inside a %d column: Width is ignored", rect.W, want)
	}
}

