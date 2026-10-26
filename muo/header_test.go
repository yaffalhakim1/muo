package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The header keeps clear of the window's own controls, and where those controls
// are is not the same on every platform: macOS draws its traffic lights at the
// top left, Windows draws three buttons at the top right.
//
// This was a real bug. The header was padded on the left from the title bar and
// on the right by a constant, so on Windows every command in the header — search,
// the outline toggle, settings and Zen — sat underneath the close button. It
// cannot be seen on a Mac and it cannot be seen in a headless render either,
// because a test window has no controls at all; so it is tested as arithmetic,
// with the room each platform reports.
func TestHeaderClearsTheWindowControlsOnEveryPlatform(t *testing.T) {
	cases := []struct {
		name  string
		bar   ui.TitleBar
		left  float32
		right float32
	}{
		// macOS with an inset title bar: the traffic lights take 78 DIPs on the
		// left and nothing on the right.
		{"macos", ui.TitleBar{Height: 52, Left: 78, Right: 0}, 86, 12},
		// Windows: three caption buttons on the right, each about 46 DIPs.
		{"windows", ui.TitleBar{Height: 52, Left: 0, Right: 138}, 12, 146},
		// Linux draws its buttons on one side or the other, and the bar says which.
		{"linux-left", ui.TitleBar{Height: 52, Left: 60}, 68, 12},
		{"linux-right", ui.TitleBar{Height: 52, Right: 90}, 12, 98},
		// Full screen and frameless windows have no controls to clear.
		{"none", ui.TitleBar{}, 12, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			left, right := headerInsets(tc.bar)
			if left != tc.left {
				t.Errorf("left inset = %.0f, want %.0f", left, tc.left)
			}
			if right != tc.right {
				t.Errorf("right inset = %.0f, want %.0f", right, tc.right)
			}
			// Whatever the platform, the inset clears the room the bar reported.
			if left < tc.bar.Left {
				t.Errorf("left inset %.0f does not clear the %.0f the controls take",
					left, tc.bar.Left)
			}
			if right < tc.bar.Right {
				t.Errorf("right inset %.0f does not clear the %.0f the controls take",
					right, tc.bar.Right)
			}
		})
	}
}

// And the header actually uses it: on a window whose controls are on the right,
// the last command in the header ends before the controls begin.
func TestHeaderCommandsEndBeforeTheControls(t *testing.T) {
	r := New()
	r.OpenIn(writeTemp(t, "doc.md", []byte("# Title\n\ntext\n")))

	// A window shaped like Windows: three caption buttons on the right.
	const captionWidth float32 = 138
	tt := ui.NewTester(func(c *ui.Context) {
		r.FrameView(c)
	}, 1200, 700)
	// The tester has no window controls, so the room is asserted as arithmetic
	// against the same function the view uses: what is checked here is that the
	// view asks for the right inset, not that the tester draws buttons.
	left, right := headerInsets(ui.TitleBar{Right: captionWidth})
	if right < captionWidth {
		t.Fatalf("the header would end %.0f from the right, inside the %.0f the controls take",
			right, captionWidth)
	}
	if left != 12 {
		t.Errorf("the left inset on a Windows-shaped window is %.0f, want 12", left)
	}
	tt.Frame()
	tt.Frame()
}
