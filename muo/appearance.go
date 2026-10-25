package muo

import (
	"sync/atomic"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// appearanceMode is the reader's choice; it is read every frame and written
// from the settings surface, so it is atomic.
var appearanceMode atomic.Uint32

// Appearance. The page follows the desktop by default, which is what a native
// app should do; the override here is for the reader who wants a fixed light
// page by day or a dark one at night regardless of the system, and it is
// labelled as such rather than presented as the normal way to run the app
// (dark-mode.md › Best practices).
// AppearanceMode is how the page chooses its light or dark surface.
type AppearanceMode uint8

const (
	// AppearanceSystem follows the desktop.
	AppearanceSystem AppearanceMode = iota
	// AppearanceLight keeps a light page.
	AppearanceLight
	// AppearanceDark keeps a dark page.
	AppearanceDark
)

// setAppearance records the choice and, when a window is running, applies it to
// the window chrome too, so the title bar and the page never disagree.
func setAppearance(mode AppearanceMode) {
	appearanceMode.Store(uint32(mode))
	source := mygo.ThemeSystem
	switch mode {
	case AppearanceLight:
		source = mygo.ThemeLight
	case AppearanceDark:
		source = mygo.ThemeDark
	}
	// The window chrome follows the choice, but only once the app is running:
	// in a headless test there is no window to re-theme.
	if appearanceAppliesSource != nil {
		appearanceAppliesSource(source)
	}
}

// appearanceAppliesSource is set by the app at startup, so that the setting can
// be exercised without a window.
var appearanceAppliesSource func(source mygo.ThemeSource)

// SetAppearanceSource installs the hook that applies an appearance choice to
// the window chrome. The app calls it once it is running; a test leaves it nil.
func SetAppearanceSource(fn func(source mygo.ThemeSource)) {
	appearanceAppliesSource = fn
}

// appearanceDark reports whether the page is dark this frame, from the chosen
// mode or, when following the system, from the window's appearance.
func appearanceDark(c *ui.Context) bool {
	switch AppearanceMode(appearanceMode.Load()) {
	case AppearanceLight:
		return false
	case AppearanceDark:
		return true
	}
	return c.Theme().Dark
}


// appearanceNames are the modes as the settings control shows them, in the
// order of the AppearanceMode constants. They are one word each because the
// control that shows them is a segmented one, where a sentence would not fit
// and would not need to: the description under it says what "System" means.
var appearanceNames = []string{"System", "Light", "Dark"}

