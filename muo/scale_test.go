package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The chrome must be built on one spacing unit and one control height, or it
// reads as assembled rather than designed. This checks the numbers rather than
// the look.
func TestChromeUsesOneScale(t *testing.T) {
	src := "# Heading\n\ntext\n"
	r := New()
	r.OpenIn(writeTemp(t, "c.md", []byte(src)))
	r.SetSidebarWidth(sidebarWidth)

	tt := ui.NewTester(r.FrameView, 884, 700)
	tt.Frame()
	tt.Frame()

	// The title bar's controls and the sidebar's rows come from the theme's
	// own metrics; nothing here should be a number I invented.
	var theme *ui.Theme
	probe := ui.NewTester(func(c *ui.Context) {
		theme = c.Theme()
		ui.Text(c, "x")
	}, 100, 100)
	probe.Frame()
	t.Logf("theme: FontSize=%.1f Spacing=%.1f Radius=%.1f",
		theme.FontSize, theme.Spacing, theme.Radius)

	// The library toggle sits inside the title bar, not clipped by it.
	rect, ok := tt.Find("Hide Library")
	if !ok {
		t.Fatal("the library toggle has no accessible label")
	}
	t.Logf("toggle: x=%.0f y=%.0f w=%.0f h=%.0f", rect.X, rect.Y, rect.W, rect.H)
	if rect.X < 0 || rect.X+rect.W > sidebarWidth {
		t.Fatalf("the toggle at x=%.0f..%.0f leaves the sidebar's %.0f",
			rect.X, rect.X+rect.W, float32(sidebarWidth))
	}
	if rect.Y < 0 || rect.Y+rect.H > titleBarHeight {
		t.Fatalf("the toggle at y=%.0f..%.0f leaves the %.0f title bar",
			rect.Y, rect.Y+rect.H, float32(titleBarHeight))
	}
}

