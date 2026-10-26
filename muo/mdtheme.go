package muo

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/markdown"
)

// The Markdown page's look. The reader sets its own type and page colors, then
// builds the code theme on top of them: the editor's default token colors are
// tuned for the editor's background, so they cannot simply be reused on the
// page's recessed block.
func (r *Reader) pageTheme(c *ui.Context, pal pagePalette) *markdown.Theme {
	// Start from the renderer's defaults for the window theme, which carry
	// every value the reader does not set.
	base := c.Theme()
	md := markdown.DefaultTheme(base)

	md.FontSize = r.typ.FontPx
	md.LineHeight = r.typ.Line
	// The page is set in the reader's serif, and the measure was solved in the
	// same family: a column computed in one face and drawn in another is a
	// measure that is wrong by however much the two differ.
	md.Font = pageFamily
	md.Text = pal.inkColor()
	md.Strong = uiRGB(triple(palette(pal.dark).ink))
	md.Muted = pal.mutedColor()
	// Links take the accent as a *text* colour, which in the dark is the lighter
	// of the two: the deeper one that fills the primary button would sit at
	// 2.6:1 on the page, which is not a link, it is a rumour.
	md.Link = uiRGB(triple(palette(pal.dark).accent))
	md.Rule = pal.ruleColor()
	// The quote is a quiet passage, not a box: a hairline bar in the rule's own
	// colour and the words in the muted ink, italic in the serif's own voice.
	md.QuoteBar = uiRGB(mixColor(pal.rule, pal.ink, 0.12))
	md.QuoteBarWidth = 2
	md.QuoteText = pal.mutedColor()
	md.TableBorder = pal.ruleColor()
	// Prose wants more air between paragraphs than the renderer's default: the
	// design sets 1.25rem of space between blocks on a 1.05rem body, which is a
	// full em and a quarter. At a 1.8 line height anything tighter leaves
	// paragraphs reading as one run of text rather than as separate thoughts.
	md.BlockSpacing = 1.25
	md.ItemSpacing = 0.35

	// Heading colors must be set, not left to the renderer: a zero heading
	// color takes the window theme's text color, which is the app's ink and not
	// the page's. On a light page with a dark theme that renders near-white
	// headings on paper.
	//
	// The sizes here are the ones headingPx gives the headings the reader draws
	// itself; they are kept in step so that the renderer's own margin maths, which
	// reads Top and Bottom from this table, spaces the page the way the sizes
	// imply. The top step is the widest: a heading has to be clearly a different
	// size from the body, or it is only bold text with air around it.
	// The sizes and weights here are the ones the reader draws its headings
	// with, from the same tables: the renderer's own margin maths reads Top and
	// Bottom from this list, so the two must agree or the page spaces headings by
	// a size it does not draw them at.
	weights := headingWeights
	tops := [6]float32{1.9, 1.55, 1.35, 1.15, 1.0, 0.9}
	bottoms := [6]float32{0.75, 0.55, 0.45, 0.36, 0.28, 0.24}
	sizes := headingScale
	for i := range md.Headings {
		// Headings take the full-strength ink; the prose under them takes the
		// softer body ink. That is the design's own split, and it is the right
		// one: a heading is read in a glance and can spend contrast, a paragraph
		// cannot.
		md.Headings[i].Color = uiRGB(triple(palette(pal.dark).ink))
		md.Headings[i].Weight = weights[i]
		md.Headings[i].Size = sizes[i]
		md.Headings[i].Top = tops[i]
		md.Headings[i].Bottom = bottoms[i]
		// No rule under the first two levels: on a page of prose a hairline
		// under a heading is a second divider where the space already divides.
		md.Headings[i].Rule = false
	}
	// The deepest level is told apart by colour rather than by size, since it is
	// already at the body's size: a slightly quieter ink says "this is a
	// sub-heading" without making it smaller than what it introduces.
	md.Headings[5].Color = pal.mutedColor()

	// Code sits slightly recessed from the page, never inverted: a dark panel in
	// a light page or a light one in a dark page reads as a different surface
	// pasted in. The recess is a mix of the page's own two colors, so it follows
	// both appearances and Increase Contrast.
	block := mixColor(pal.paper, pal.ink, 0.06)
	md.CodeBackground = uiRGB(mixColor(pal.paper, pal.ink, 0.09))
	md.Code = pal.inkColor()
	md.CodeBlockBackground = uiRGB(block)
	md.CodeBlockBorder = uiRGB(mixColor(pal.paper, pal.ink, 0.02))
	md.CodeBlockRadius = 10
	md.CodeBlockPadding = 0.95
	md.Checkbox = uiRGB(triple(palette(pal.dark).accent))
	md.Mark = ui.RGBA(250, 204, 21, 0.35)

	// Tables are the page's own table: a rule between rows, a tinted header, no
	// wide side scroll. The renderer caps each column at TableMaxColumn ems and
	// adds the cell padding outside that cap, so the cap is widened by the
	// padding to stop a column from being measured narrower than its text
	// (which wrapped "Background" into "Backgroun/d").
	md.TableHeader = uiRGB(mixColor(pal.paper, pal.ink, 0.045))
	md.TableStripe = uiRGB(mixColor(pal.paper, pal.ink, 0.018))
	md.TableBorder = uiRGB(mixColor(pal.paper, pal.ink, 0.13))
	// Cells get room to breathe: at the renderer's 0.6 em across, a two-column
	// table of short words reads as a cramped box rather than as a table.
	//
	// The across value is a full em and not 0.85, and the extra tenth is not
	// cosmetic. The renderer sizes a column to its widest cell's measured width
	// plus twice this padding plus one DIP, so the padding is the whole margin a
	// word has before it wraps. At 0.85 em a heading as short as "Ratio" broke
	// after five letters, because its measured width landed a fraction of a DIP
	// inside the column.
	md.TableCellPadding = [2]float32{1.0, 0.6}
	md.TableMaxColumn = 26

	md.CodeColors = codeTheme(base, pal, block)
	return md
}

// codeTheme colors the tokens of a fenced code block.
//
// The palette keeps a code convention readers already know, over the page's own
// recessed surface: warm keywords, cooler types and functions, a warm string,
// a grey comment. Every color is chosen against `block`, not against the page,
// and every one clears 4.5:1 there, so highlighting never costs legibility.
func codeTheme(base *ui.Theme, pal pagePalette, block rgbColor) *editor.Theme {
	ct := editor.DefaultTheme(base)
	ct.Background = uiRGB(block)
	ct.Text = pal.inkColor()
	ct.LineNumber = pal.mutedColor()
	ct.Selection = base.Accent.Alpha(0.22)
	ct.InactiveSelection = base.Accent.Alpha(0.12)

	pick := func(light, night string) ui.Color {
		if pal.dark {
			return ui.Hex(night)
		}
		return ui.Hex(light)
	}

	// One role, one color. The pairs are the same hue at two lightnesses, so the
	// block reads as one palette in either appearance rather than as two.
	keyword := pick("#9d2235", "#f28ba8")  // control flow, declarations
	typ := pick("#0f6b4f", "#6fd0a8")      // type names, classes
	function := pick("#1c5fa8", "#7db8f0") // calls, definitions
	str := pick("#7a4a0f", "#e3b06a")      // strings, chars
	num := pick("#8a5a00", "#f0c060")      // numbers, constants
	comment := pick("#6b7069", "#8f958e")  // comments, doc comments
	meta := pick("#6b4a8a", "#c9a6e8")     // decorators, preprocessor
	tag := pick("#1c5fa8", "#7db8f0")      // markup tags
	invalid := pick("#b3261e", "#ff8a80")  // errors

	set := func(kind editor.TokenKind, color ui.Color) {
		style := ct.Tokens[kind]
		style.Color = color
		ct.Tokens[kind] = style
	}
	set(editor.TokenKeyword, keyword)
	set(editor.TokenType, typ)
	set(editor.TokenFunction, function)
	set(editor.TokenString, str)
	set(editor.TokenEscape, num)
	set(editor.TokenRegexp, str)
	set(editor.TokenNumber, num)
	set(editor.TokenConstant, num)
	set(editor.TokenComment, comment)
	set(editor.TokenMeta, meta)
	set(editor.TokenTag, tag)
	set(editor.TokenAttribute, typ)
	set(editor.TokenLink, function)
	set(editor.TokenHeading, keyword)
	set(editor.TokenStrong, pal.inkColor())
	set(editor.TokenEmphasis, pal.inkColor())
	set(editor.TokenVariable, pal.inkColor())
	set(editor.TokenProperty, pal.inkColor())
	set(editor.TokenOperator, pal.mutedColor())
	set(editor.TokenPunctuation, pal.mutedColor())
	set(editor.TokenInserted, typ)
	set(editor.TokenDeleted, keyword)
	set(editor.TokenInvalid, invalid)

	// Two weight changes carry what color alone should not: comments read as
	// notes, and inserted and deleted lines in a diff stand apart.
	commentStyle := ct.Tokens[editor.TokenComment]
	commentStyle.Italic = true
	ct.Tokens[editor.TokenComment] = commentStyle
	emph := ct.Tokens[editor.TokenEmphasis]
	emph.Italic = true
	ct.Tokens[editor.TokenEmphasis] = emph
	strong := ct.Tokens[editor.TokenStrong]
	strong.Bold = true
	ct.Tokens[editor.TokenStrong] = strong

	return ct
}
