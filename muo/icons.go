package muo

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Icons drawn the way MyGo's vibrancy example draws its own: a 16-point viewBox,
// a 1.3 stroke, round caps and joins, and currentColor so they inherit the text
// colour and the theme's tint. One visual language across the window, which is
// what makes a set of small marks read as one app rather than a collage.
func icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	iconSidebar  = icon(`<rect x="1.5" y="2.5" width="13" height="11" rx="2"/><path d="M6 2.5v11"/>`)
	iconOutline  = icon(`<rect x="1.5" y="2.5" width="13" height="11" rx="2"/><path d="M10 2.5v11"/>`)
	iconSearch   = icon(`<circle cx="7" cy="7" r="4.5"/><path d="M10.4 10.4 14.5 14.5"/>`)
	iconSettings = icon(`<path d="M6.6 1.9a1.4 1.4 0 0 1 2.8 0l.1.6a1 1 0 0 0 1.5.7l.5-.3a1.4 1.4 0 0 1 2 2l-.3.5a1 1 0 0 0 .7 1.5l.6.1a1.4 1.4 0 0 1 0 2.8l-.6.1a1 1 0 0 0-.7 1.5l.3.5a1.4 1.4 0 0 1-2 2l-.5-.3a1 1 0 0 0-1.5.7l-.1.6a1.4 1.4 0 0 1-2.8 0l-.1-.6a1 1 0 0 0-1.5-.7l-.5.3a1.4 1.4 0 0 1-2-2l.3-.5a1 1 0 0 0-.7-1.5l-.6-.1a1.4 1.4 0 0 1 0-2.8l.6-.1a1 1 0 0 0 .7-1.5l-.3-.5a1.4 1.4 0 0 1 2-2l.5.3a1 1 0 0 0 1.5-.7z"/><circle cx="8" cy="8" r="2.1"/>`)
	iconDocument = icon(`<path d="M9 1.5H4.5A1.5 1.5 0 0 0 3 3v10a1.5 1.5 0 0 0 1.5 1.5h7A1.5 1.5 0 0 0 13 13V5.5z"/><path d="M9 1.5v4h4"/>`)
	iconPDF      = icon(`<path d="M9 1.5H4.5A1.5 1.5 0 0 0 3 3v10a1.5 1.5 0 0 0 1.5 1.5h7A1.5 1.5 0 0 0 13 13V5.5z"/><path d="M9 1.5v4h4"/><path d="M5.5 11.5V8.8h1a.9.9 0 0 1 0 1.8h-1"/>`)
	iconOpen     = icon(`<path d="M1.5 12.5v-9A1 1 0 0 1 2.5 2.5h3l1.5 2h6a1 1 0 0 1 1 1v7a1 1 0 0 1-1 1h-10a1 1 0 0 1-1-1z"/>`)
	iconPlus     = icon(`<path d="M8 3v10M3 8h10"/>`)

	// Callout marks, in the same visual language as the rest: 16pt viewBox, a
	// 1.3 stroke, currentColor.
	iconInfo    = icon(`<circle cx="8" cy="8" r="6.2"/><path d="M8 7.2v4"/><path d="M8 4.9v.1"/>`)
	iconBulb    = icon(`<path d="M6.2 11.2a4.2 4.2 0 1 1 3.6 0v1.6a1 1 0 0 1-1 1H7.2a1 1 0 0 1-1-1z"/><path d="M6.6 8.6a1.6 1.6 0 0 1 2.8 0"/>`)
	iconFlag    = icon(`<path d="M3.6 14V2.4"/><path d="M3.6 3h7.2l-1 2.6 1 2.6H3.6"/>`)
	iconWarning = icon(`<path d="M8 2.6 14.2 13H1.8z"/><path d="M8 6.6v3"/><path d="M8 11.4v.1"/>`)
	iconStop    = icon(`<path d="M5.6 1.8h4.8l3.8 3.8v4.8l-3.8 3.8H5.6L1.8 10.4V5.6z"/><path d="M8 5v3.4"/><path d="M8 10.4v.1"/>`)
	iconType     = icon(`<path d="M2.5 4V2.5h11V4M8 2.5v11M5.5 13.5h5"/>`)
	iconLeft     = icon(`<path d="M10 3.5 5.5 8l4.5 4.5"/>`)
	iconRight    = icon(`<path d="M6 3.5 10.5 8 6 12.5"/>`)
	iconClose    = icon(`<path d="M4 4l8 8M12 4l-8 8"/>`)
	// iconZen is the expand mark: four corners turned outward. It says "more
	// page", which is what the mode does.
	iconZen = icon(`<path d="M9.5 2.5H13.5V6.5M6.5 13.5H2.5V9.5M13.5 2.5 9.5 6.5M2.5 13.5 6.5 9.5"/>`)
)

// animationUnit is the sidebar's slide duration.
const animationUnit = time.Millisecond

// iconButton is an icon-only command: no face of its own, a rounded hover
// behind it, and a label and tooltip for the pointer and for assistive
// technology.
//
// The padding is zeroed, and that is the whole point of the helper. A button
// carries 14 DIPs of padding either side for its label, so an icon in a 30-DIP
// button is laid out in the 2 DIPs that are left and sits hard against one
// edge. Setting the size alone does not remove the padding — it only makes the
// overflow worse — so every icon button in the window goes through here.
func iconButton(c *ui.Context, mark *ui.SVG, label string, w, h float32) ui.Element {
	t := c.Theme()
	b := ui.Button(c, "").Size(w, h).Radius(6).Padding(0).
		Background(ui.Transparent).BorderWidth(0).
		TextColor(t.TextMuted).
		Label(label).Tooltip(label)
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() { ui.Icon(c, mark).Size(16, 16) })
	return b
}

// boolToFloat is the target of an animation that moves between 0 and 1.
func boolToFloat(b bool) float32 {
	if b {
		return 1
	}
	return 0
}
