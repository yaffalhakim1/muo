package muo

import (
	"image"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Icon centering is checked by ink, not by intent: the tester draws a frame and
// these read where the non-background pixels actually land, so an icon that is
// off by a pixel or two inside its button is a failure rather than a hunch.
//
// It is worth checking because the failure is invisible in code and obvious on
// screen: a stroke drawn on a half-pixel grid, or a viewBox whose shapes do not
// sit on its centre, reads as a smudge that is "almost right" and never
// resolves.

// inkBox returns the bounding box of pixels that differ from the frame's own
// background, inside a region.
func inkBox(img *image.RGBA, r ui.Rect) (image.Rectangle, bool) {
	bg := img.RGBAAt(int(r.X)+1, int(r.Y)+1)
	minX, minY := 1<<30, 1<<30
	maxX, maxY := -1, -1
	for y := int(r.Y); y < int(r.Y+r.H); y++ {
		for x := int(r.X); x < int(r.X+r.W); x++ {
			if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
				continue
			}
			c := img.RGBAAt(x, y)
			if !closeColor(c, bg, 6) {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < 0 {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}

func closeColor(a, b color.RGBA, tolerance int) bool {
	d := func(x, y uint8) int {
		if x > y {
			return int(x) - int(y)
		}
		return int(y) - int(x)
	}
	return d(a.R, b.R) <= tolerance && d(a.G, b.G) <= tolerance && d(a.B, b.B) <= tolerance
}

// TestSidebarToggleIconIsCentered checks the icon the app draws inside a fixed-size\r\n// button: the library toggle in the title bar.
func TestLibraryToggleIconIsCentered(t *testing.T) {
	r := New()
	r.OpenIn(writeTemp(t, "s.md", []byte("# T\n\nText.\n")))
	tt := ui.NewTester(r.FrameView, 1100, 760)
	tt.Frame()
	tt.Frame()

	button, ok := tt.Find("Hide Library")
	if !ok {
		t.Fatal("the library toggle was not found")
	}
	box, ok := inkBox(tt.Image(), button)
	if !ok {
		t.Fatal("the library toggle drew nothing")
	}
	// The button's own box, in device pixels: the tester renders at 1x.
	offX := float32(box.Min.X+box.Max.X)/2 - (button.X + button.W/2)
	offY := float32(box.Min.Y+box.Max.Y)/2 - (button.Y + button.H/2)
	t.Logf("toggle %.0fx%.0f at (%.0f,%.0f); ink %v; offset (%.1f, %.1f)",
		button.W, button.H, button.X, button.Y, box, offX, offY)
	if offX > 1.5 || offX < -1.5 {
		t.Errorf("the icon is %.1f DIPs off centre horizontally", offX)
	}
	if offY > 1.5 || offY < -1.5 {
		t.Errorf("the icon is %.1f DIPs off centre vertically", offY)
	}
}

// TestTitleRowIconsAreCentered checks the two borderless commands beside the
// readout: they are drawn with no face of their own, so a stray border would
// show up as ink at the very edge of the button.
//
// The sidebar is hidden for this, so that the title bar's Reading Settings
// button is the only element with that name: the sidebar's own item carries
// the same label, and the same command in two places should read the same.
func TestTitleRowIconsAreCentered(t *testing.T) {
	r := New()
	r.OpenIn(writeTemp(t, "s.md", []byte("# T\n\nText.\n")))
	r.showOutline = false
	tt := ui.NewTester(r.FrameView, 1100, 760)
	tt.Frame()
	tt.Frame()

	for _, label := range []string{"Find in Document", "Reading Settings"} {
		button, ok := tt.Find(label)
		if !ok {
			t.Fatalf("%q was not found", label)
		}
		if button.Y > titleBarHeight {
			t.Fatalf("%q was found outside the title bar, at y=%.0f", label, button.Y)
		}
		box, ok := inkBox(tt.Image(), button)
		if !ok {
			t.Fatalf("%q drew nothing", label)
		}
		offX := float32(box.Min.X+box.Max.X)/2 - (button.X + button.W/2)
		offY := float32(box.Min.Y+box.Max.Y)/2 - (button.Y + button.H/2)
		t.Logf("%-20s %.0fx%.0f; ink %dx%d; offset (%.1f, %.1f)",
			label, button.W, button.H, box.Dx(), box.Dy(), offX, offY)
		if offX > 1.5 || offX < -1.5 {
			t.Errorf("%q is %.1f DIPs off centre horizontally", label, offX)
		}
		if offY > 1.5 || offY < -1.5 {
			t.Errorf("%q is %.1f DIPs off centre vertically", label, offY)
		}
	}
}
