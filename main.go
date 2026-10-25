package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"example.com/muo/muo"
)

func main() {
	// A file named on the command line is opened first; anything opened later
	// with the app joins it as another document.
	initial := ""
	if len(os.Args) > 1 {
		if path, err := filepath.Abs(os.Args[1]); err == nil {
			initial = path
		}
	}

	r := muo.New()

	// The window follows the desktop's appearance, not the app's
	// (dark-mode.md › Best practices).
	mygo.Theme.SetSource(mygo.ThemeSystem)
	muo.SetAppearanceSource(mygo.Theme.SetSource)

	// On macOS the app keeps its menu bar with no window, so its commands stay
	// reachable; elsewhere it ends with its last window.
	if runtime.GOOS == "darwin" {
		mygo.App.OnWindowAllClosed(func() {})
	}

	mygo.App.WhenReady(func() {
		// The window is sized in the display's own points, so it fits the work
		// area it opens on instead of running past the screen's edge.
		d := mygo.Screen.DisplayNearestPoint(mygo.Screen.CursorScreenPoint())
		scale := d.ScaleFactor
		if scale <= 0 {
			scale = 1
		}
		workW := float32(d.WorkArea.Width) / float32(scale)
		workH := float32(d.WorkArea.Height) / float32(scale)

		width := workW * 0.72
		if width < 780 {
			width = 780
		}
		if width > workW-32 {
			width = workW - 32
		}
		height := workH - 100
		if height < 560 {
			height = 560
		}

		r.SetSidebarWidth(muo.SidebarWidth)

		win := mygo.NewWindow(mygo.WindowOptions{
			Title:          "Muo",
			Width:          int(width),
			Height:         int(height),
			MinWidth:       620,
			MinHeight:      460,
			StateKey:       "reader",
			TitleBarStyle:  mygo.TitleBarHiddenInset,
			TitleBarHeight: muo.TitleBarHeight,
			// The material shows wherever the view paints no background: the
			// sidebar is translucent and only the page is opaque.
			Vibrancy: mygo.VibrancySidebar,
			Content:  ui.View(r.FrameView),
		})
		// A window remembers its bounds, which is right until it is reopened on
		// a display smaller than the one it was last on: the saved size is then
		// wider than the screen and the right-hand end of the title row — the
		// readout and the two commands — is off the edge with no way to reach
		// it. The saved size is kept, so the reader's window comes back as they
		// left it, but it is brought inside the display it opens on.
		//
		// It is done as the window is first shown, not at creation: the saved
		// bounds are applied on the way to the screen, so a clamp made before
		// that is the clamp of the size the window is about to be given.
		win.OnShow(func() { fitToDisplay(win) })
		r.Attach(win)
		win.OnClose(func(e *mygo.CloseEvent) {
			// Where the reader stopped, and the window's own layout, are
			// written before the window goes, so a crash in the close path
			// cannot cost them either.
			if err := r.Save(); err != nil {
				fmt.Fprintln(os.Stderr, "muo: saving your place:", err)
			}
			if err := r.SaveLayout(); err != nil {
				fmt.Fprintln(os.Stderr, "muo: saving the layout:", err)
			}
		})
		win.OnClosed(func() { mygo.App.Quit() })
		r.InstallMenu()
		mygo.App.OnOpenFile(r.OpenIn)
		if initial != "" {
			r.OpenIn(initial)
		}
	})
	if err := mygo.App.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "muo:", err)
	}
}

// fitToDisplay brings a window inside the work area of the display it opened
// on, keeping its size where the display allows it.
//
// Restoring a saved size is right, and restoring a size larger than the display
// is not: a window 14 points wider than the screen puts the far end of its
// title row out of reach, and the reader has no way to know the commands are
// there. So the size is kept, clamped to what the display can show, and the
// window is moved back inside the work area if the restore left it hanging off
// an edge.
func fitToDisplay(win *mygo.Window) {
	r := win.Bounds()
	if r.Width <= 0 || r.Height <= 0 {
		return
	}
	d := mygo.Screen.DisplayMatching(r)
	work := d.WorkArea
	if work.Width <= 0 || work.Height <= 0 {
		return
	}
	next := r
	if next.Width > work.Width {
		next.Width = work.Width
	}
	if next.Height > work.Height {
		next.Height = work.Height
	}
	if next.X+next.Width > work.X+work.Width {
		next.X = work.X + work.Width - next.Width
	}
	if next.Y+next.Height > work.Y+work.Height {
		next.Y = work.Y + work.Height - next.Height
	}
	if next.X < work.X {
		next.X = work.X
	}
	if next.Y < work.Y {
		next.Y = work.Y
	}
	if next != r {
		win.SetBounds(next)
	}
}
