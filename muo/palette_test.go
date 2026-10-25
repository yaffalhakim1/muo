package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The page's two colors are the whole reading surface; swapping them turns a
// light page into ink-on-ink. This pins the direction in both appearances.
func TestPageColorsFollowTheAppearance(t *testing.T) {
	cases := []struct {
		name      string
		dark      bool
		wantPaper rgbColor
		wantInk   rgbColor
	}{
		{"light", false, rgbColor{0xfa, 0xf8, 0xf5}, rgbColor{0x1f, 0x1d, 0x1b}},
		{"dark", true, rgbColor{0x1a, 0x1a, 0x1c}, rgbColor{0xcd, 0xc8, 0xc2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paper, ink, _, _, _ := readingSurface(tc.dark)
			if paper != tc.wantPaper {
				t.Fatalf("paper = #%02x%02x%02x, want #%02x%02x%02x",
					paper.R, paper.G, paper.B, tc.wantPaper.R, tc.wantPaper.G, tc.wantPaper.B)
			}
			if ink != tc.wantInk {
				t.Fatalf("ink = #%02x%02x%02x, want #%02x%02x%02x",
					ink.R, ink.G, ink.B, tc.wantInk.R, tc.wantInk.G, tc.wantInk.B)
			}
			// Whatever the appearance, the paper is the surface and the ink is
			// readable against it: in light mode the paper is the lighter, in
			// dark mode the ink is.
			if tc.dark && luminance(paper) >= luminance(ink) {
				t.Fatalf("dark page: paper %.2f should be darker than ink %.2f",
					luminance(paper), luminance(ink))
			}
			if !tc.dark && luminance(paper) <= luminance(ink) {
				t.Fatalf("light page: paper %.2f should be lighter than ink %.2f",
					luminance(paper), luminance(ink))
			}
		})
	}
}

// pageColors must hand the palette back in the same direction the frame's own
// theme agrees with, so the page never inverts against the chrome.
func TestPageColorsAgreeWithTheTheme(t *testing.T) {
	var got pagePalette
	var themeDark bool
	tt := ui.NewTester(func(c *ui.Context) {
		themeDark = c.Theme().Dark
		got = pageColors(c)
		ui.Text(c, "x")
	}, 400, 300)
	tt.Frame()

	if got.dark != themeDark {
		t.Fatalf("the page chose %v but the window theme is %v", got.dark, themeDark)
	}
	if themeDark {
		if got.paper == (rgbColor{0xfa, 0xf8, 0xf5}) {
			t.Fatal("a dark window got the light paper")
		}
		return
	}
	if got.paper != (rgbColor{0xfa, 0xf8, 0xf5}) {
		t.Fatalf("a light window got paper #%02x%02x%02x", got.paper.R, got.paper.G, got.paper.B)
	}
}
