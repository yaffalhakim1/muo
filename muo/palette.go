package muo

import "github.com/egoist/mygo/ui"

// The page's colors. Reading wants a lower contrast than maximum, but the
// desktop's own appearance and accessibility settings decide the direction:
// the light or dark page follows the system, and Increase Contrast moves the
// page toward the system's label colors rather than away from them
// (dedark-mode.md › Best practices: "Ensure that your app looks good in both
// appearance modes"; › Dark Mode colors: "Aim for sufficient color contrast in
// all appearances").

// pageColors resolves the reading surface for this frame. The appearance is the
// desktop's, read from the frame's preferences rather than from the theme: the
// theme's appearance is resolved once per window, so a page that switched with
// the system at sunset instead of at launch. Increase Contrast then tightens the
// pair toward the system's own label colors.
func pageColors(c *ui.Context) pagePalette {
	prefs := c.Preferences()
	dark := appearanceDark(c)
	paper, ink, muted, rule, selection := readingSurface(dark)
	if prefs.HighContrast {
		// Move the page toward the system's own label colors, which the desktop
		// has already tuned for legibility in this appearance.
		ink = mixColor(ink, colorOfTheme(c.Theme().Text), 0.6)
		muted = mixColor(muted, colorOfTheme(c.Theme().TextMuted), 0.4)
	}
	return pagePalette{dark: dark, ink: ink, paper: paper, muted: muted, rule: rule, selection: selection}
}

// rgbColor is a plain triple the page mixes and the UI layer turns into a theme
// color. It is deliberately not ui.Color: the page does arithmetic on its own
// values, and a theme color carries wide-gamut state that mixing would discard.
type rgbColor struct{ R, G, B uint8 }

// uiRGB makes a theme color out of a triple.
func uiRGB(c rgbColor) ui.Color { return ui.RGB(c.R, c.G, c.B) }

// pagePalette is the resolved reading surface for one frame.
type pagePalette struct {
	dark              bool
	ink, paper, muted rgbColor
	rule, selection   rgbColor
}

func (p pagePalette) inkColor() ui.Color    { return uiRGB(p.ink) }
func (p pagePalette) paperColor() ui.Color  { return uiRGB(p.paper) }
func (p pagePalette) mutedColor() ui.Color  { return uiRGB(p.muted) }
func (p pagePalette) ruleColor() ui.Color   { return uiRGB(p.rule) }
func (p pagePalette) selectionColor() ui.Color { return uiRGB(p.selection) }

// colorOfTheme turns a theme color back into the reader's triple, so the page
// can mix toward the system's own choices.
func colorOfTheme(c ui.Color) rgbColor { return rgbColor{R: c.R, G: c.G, B: c.B} }

// mixColor moves a towards b by t, 0 to 1.
func mixColor(a, b rgbColor, t float64) rgbColor {
	if t <= 0 {
		return a
	}
	if t > 1 {
		t = 1
	}
	mix := func(x, y uint8) uint8 {
		return uint8(float64(x)*(1-t) + float64(y)*t + 0.5)
	}
	return rgbColor{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B)}
}

