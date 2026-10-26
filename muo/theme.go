package muo

import (
	"github.com/egoist/mygo/ui"
)

// The window's colours come from one place: the page's own paper.
//
// A reader is a sheet of paper in a window, so the chrome has to be the same
// paper in a different light. It was not, before: the page was warm cream and
// the sidebar was the toolkit's cool grey, and the two neutrals meeting at a
// hairline read as two materials glued together rather than as one object.
// That is the single defect this palette exists to remove.
//
// One colour is left over for meaning. The accent is an ink blue, because a
// reader's interactive colour should be the colour of a pen annotation rather
// than a brand: it marks links, the chosen row, the primary action and the
// focus ring, and it appears nowhere else. The three grades the type model
// gives — in the band, too small, too large — are the only other colour in the
// app, and they appear only in the readout that reports them.

// inkPalette is the app's whole colour set for one appearance.
type inkPalette struct {
	paper     ui.Color // the page
	chrome    ui.Color // the sidebar and every control's face
	raised    ui.Color // a control under the pointer
	pressed   ui.Color // a control being pressed
	border    ui.Color // hairlines and outlines
	ink       ui.Color // text on the paper
	// body is the page's prose ink, which is a shade softer than the chrome's.
	// A heading is short and wants the full contrast; a paragraph runs for
	// hundreds of words, and the halation of maximum contrast is a documented
	// fatigue source over a long session. The chrome keeps `ink`, because a
	// label at 11 points has no such run to spend contrast on.
	body      ui.Color
	muted     ui.Color // secondary text, still at 4.5:1 on the paper
	accent    ui.Color // links, selection, the focus ring
	accentInk ui.Color // text on a filled accent surface
	// accentFill is the accent under a filled control. It is deeper than the
	// accent itself in the dark, where a pale blue block would be the brightest
	// thing on the screen and the text on it would fight the paper around it.
	accentFill ui.Color
	good       ui.Color // the type landed in the reading band
	near       ui.Color // outside it, but close
	bad        ui.Color // outside it, and not close
	shadow     ui.Color // the tint every shadow is made of
}

// palette returns the colours for an appearance. The dark set is not the light
// one inverted: warm charcoal and warm grey ink, because pure black under white
// text blooms at the edges, which is exactly what a reader must not do.
//
// Every text colour here clears 4.5:1 against both the paper and the chrome.
func palette(dark bool) inkPalette {
	if dark {
		return inkPalette{
			paper:      ui.Hex("#1a1a1c"),
			chrome:     ui.Hex("#151517"),
			raised:     ui.Hex("#242428"),
			pressed:    ui.Hex("#2e2e33"),
			border:     ui.Hex("#34343a"),
			ink:        ui.Hex("#e8e3db"),
			body:       ui.Hex("#d6d0c7"),
			muted:      ui.Hex("#a29a90"),
			accent:     ui.Hex("#7aa3cc"),
			accentInk:  ui.Hex("#f2f6fb"),
			accentFill: ui.Hex("#35618f"),
			good:       ui.Hex("#63bd89"),
			near:       ui.Hex("#d9a441"),
			bad:        ui.Hex("#e0776a"),
			shadow:     ui.RGBA(0, 0, 0, 0.55),
		}
	}
	return inkPalette{
		paper:      ui.Hex("#faf8f4"),
		chrome:     ui.Hex("#f5f1ea"),
		raised:     ui.Hex("#ebe4d8"),
		pressed:    ui.Hex("#e2dacb"),
		border:     ui.Hex("#e3ddd3"),
		ink:        ui.Hex("#1c1a17"),
		body:       ui.Hex("#2d2a26"),
		muted:      ui.Hex("#746e65"),
		accent:     ui.Hex("#3a5a7d"),
		accentInk:  ui.Hex("#ffffff"),
		accentFill: ui.Hex("#3a5a7d"),
		good:       ui.Hex("#2f7d4f"),
		near:       ui.Hex("#8a5d0e"),
		bad:        ui.Hex("#b3402f"),
		shadow:     ui.RGBA(31, 29, 27, 0.16),
	}
}

// The chrome's type scale, taken from the sizes macOS gives its own text
// styles. Naming them keeps the window to one hierarchy instead of the six
// nearly-equal sizes it had, where a section header, a title and a readout all
// sat within two points of each other and so nothing read as subordinate.
const (
	typeTitle1   = 22 // the wordmark on the empty page
	typeTitle2   = 17 // a pane's own name
	typeTitle3   = 15 // the document's name in the title row
	typeBody     = 13 // sidebar rows, labels, body text in the chrome
	typeCallout  = 12 // a value beside a slider
	typeSubhead  = 11 // section headers, the type readout
	typeFootnote = 10 // the smallest label the window shows
)

// ShellTheme is MyGo's own theme with the app's palette in place of the
// toolkit's. Everything else — the spacing unit, the control metrics, the
// light and dark bases — stays the platform's, which is what makes the
// controls the right size without being told.
//
// Setting a theme is the documented way to adjust one (docs/ui/styling.md):
// copy what the frame already resolved, change what you mean to, and set it.
func ShellTheme(c *ui.Context) *ui.Theme {
	t := *c.Theme()
	dark := appearanceDark(c)
	p := palette(dark)

	t.Dark = dark
	t.Background = p.paper
	t.Surface = p.chrome
	t.SurfaceHover = p.raised
	t.SurfacePressed = p.pressed
	t.Border = p.border
	t.Text = p.ink
	t.TextMuted = p.muted
	t.Accent = p.accentFill
	t.AccentHover = p.accentFill.Mix(p.ink, 0.12)
	t.AccentPressed = p.accentFill.Mix(p.ink, 0.24)
	t.AccentText = p.accentInk
	t.Danger = p.bad
	t.Warning = p.near
	t.Success = p.good
	t.Selection = p.accent.Alpha(0.22)
	t.Focus = p.accent.Alpha(0.6)
	t.Inverse = p.ink
	t.InverseText = p.paper
	t.Scrollbar = p.ink.Alpha(0.2)
	t.Radius = 6
	return &t
}

// comfortColor is the tint the type readout carries: the model's grade of the
// apparent size, so the one number the app computes says whether it landed
// well. It is never the only carrier — the readout also puts the size on a
// gauge, so the grade reads as a position as well as a colour.
func comfortColor(t *ui.Theme, k Comfort) ui.Color {
	switch k {
	case ComfortSmall:
		return t.Danger
	case ComfortLarge:
		return t.Warning
	}
	return t.Success
}
