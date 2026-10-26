package muo

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

// Callouts.
//
// GitHub's alert syntax is a block quote whose first line is a marker:
//
//	> [!NOTE]
//	> Something the reader should know.
//
// The parser knows nothing about them. It sees a quote, and the renderer draws
// the marker as text, so `[!NOTE]` appeared on the page as though it were part of
// the sentence. This file reads the marker and gives the block a tint, a label
// and an icon; the words inside are still the renderer's, drawn by the ordinary
// quote path.
//
// The five kinds are GitHub's own. Anything else, including a quote that merely
// starts with a bracket, is left as an ordinary quote.

// calloutKind is one of the five alerts and how it looks.
type calloutKind struct {
	// name is the marker as written, without the brackets: NOTE, TIP, and so on.
	name string
	// label is what the page shows in place of the marker. It is the marker's own
	// word, in sentence case: inventing a different name for a thing the document
	// already named would be a second vocabulary for the same idea.
	label string
	icon  *ui.SVG
	// tone picks the colour role, so the palette stays the one place colours are
	// written and a callout follows the appearance with everything else.
	tone string
}

var callouts = []calloutKind{
	{name: "NOTE", label: "Note", icon: iconInfo, tone: "accent"},
	{name: "TIP", label: "Tip", icon: iconBulb, tone: "good"},
	{name: "IMPORTANT", label: "Important", icon: iconFlag, tone: "accent"},
	{name: "WARNING", label: "Warning", icon: iconWarning, tone: "near"},
	{name: "CAUTION", label: "Caution", icon: iconStop, tone: "bad"},
}

// calloutOf reports whether a line opens a callout, and which kind.
//
// It reads one line: a quote whose first line is "[!NOTE]" and nothing else is a
// callout. A quote that opens with a sentence is a quote, even if it mentions a
// bracket later.
func calloutOf(line string) (calloutKind, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, ">") {
		return calloutKind{}, false
	}
	trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
	if len(trimmed) < 4 || !strings.HasPrefix(trimmed, "[!") || !strings.HasSuffix(trimmed, "]") {
		return calloutKind{}, false
	}
	name := strings.ToUpper(strings.TrimSpace(trimmed[2 : len(trimmed)-1]))
	for _, k := range callouts {
		if k.name == name {
			return k, true
		}
	}
	return calloutKind{}, false
}

// calloutSource rewrites the document's callouts before it is drawn.
//
// A callout is a block quote whose first line is a marker, as [!NOTE]. The
// renderer has no idea what that is, so it drew the marker as part of the
// sentence: the page showed "[!NOTE] Useful information" instead of a note.
//
// The marker lines are removed here, which leaves ordinary quotes the renderer
// draws properly, and each kind is recorded by the block's index so the block
// hook can style the quote it is about to draw. The indexes come from parsing the
// rewritten source, which is the same document the renderer will parse, so they
// line up exactly.
func (r *Reader) calloutSource(src string) string {
	r.callouts = nil
	if !strings.Contains(src, "[!") {
		return src
	}
	var kinds []calloutKind
	lines := strings.Split(src, "\n")
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		kind, ok := calloutOf(lines[i])
		if !ok {
			out = append(out, lines[i])
			continue
		}
		// The marker line goes; the quote's own lines stay, so the renderer
		// draws the body as the quote it is.
		kinds = append(kinds, kind)
		rest := i + 1
		for rest < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[rest]), ">") {
			out = append(out, lines[rest])
			rest++
		}
		i = rest - 1
	}
	clean := strings.Join(out, "\n")

	// Parse what the renderer will parse, and give the Nth quote the Nth kind.
	if len(kinds) > 0 {
		doc := parser.Parse(clean)
		n := 0
		for i := range doc.Children(0) {
			if doc.Block(i).Kind != parser.Quote {
				continue
			}
			if n < len(kinds) {
				if r.callouts == nil {
					r.callouts = map[int]calloutKind{}
				}
				r.callouts[i] = kinds[n]
			}
			n++
		}
	}
	return clean
}

// calloutAt reports whether the block at index i was a callout, and which kind.
func (r *Reader) calloutAt(i int) (calloutKind, bool) {
	k, ok := r.callouts[i]
	return k, ok
}
// calloutBlock draws a callout in place of the quote the renderer would draw.
//
// It takes the block rather than styling the renderer's wrapper, because the
// wrapper is built before the quote is added to it: styling it from the hook put
// the tint on the box and the words outside it. Taking the block means the label
// and the words are built here, in one element, which is what makes a callout one
// object rather than a label floating above a quote.
func (r *Reader) calloutBlock(c *ui.Context, pal pagePalette, d *parser.Document, i int, kind calloutKind) {
	t := c.Theme()
	tint, ink := calloutTone(pal, kind.tone)

	ui.Column(c).FillWidth().Radius(12).Background(tint).
		// The leading rule is the quote bar, made to carry meaning rather than
		// only to mark a quotation.
		BorderWidth(0, 0, 0, 3).BorderColor(ink).
		Padding(t.Space(3), t.Space(3)).Gap(t.Space(1.5)).
		Children(func() {
			ui.Row(c).Gap(t.Space(1.5)).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, kind.icon).Size(14, 14).TextColor(ink)
				ui.Text(c, kind.label).FontSize(typeSubhead).FontWeight(600).
					LetterSpacing(0.3).TextColor(ink)
			})
			// The words keep the page's own type, so a callout is a marked
			// passage rather than a different kind of document. They go through
			// the renderer, which is what gives them their paragraphs, lists and
			// links.
			r.quoteBody(c, d, i)
		})
}

// quoteBody draws a quote's inner blocks, with the quote's own marks removed.
//
// The renderer has no exported way to draw one block's children, so the body is
// rebuilt from the source the quote holds and handed back to the renderer. The
// hook is cleared while it runs, or a quote inside a callout would re-enter here.
func (r *Reader) quoteBody(c *ui.Context, d *parser.Document, i int) {
	lines := strings.Split(strings.TrimSpace(d.Lines(i)), "\n")
	for k, line := range lines {
		t := strings.TrimSpace(line)
		lines[k] = strings.TrimPrefix(strings.TrimPrefix(t, ">"), " ")
	}
	opts := &r.md
	opts.Theme = r.pageTheme(c, pageColors(c))
	saved := opts.Block
	opts.Block = nil
	markdown.View(c, strings.Join(lines, "\n"), opts)
	opts.Block = saved
}

// calloutTone resolves a callout's tint and ink from the page palette.
//
// The tint is the hue at about a tenth of its strength over the paper, so the
// words inside keep the contrast they had; the ink is the hue itself, which the
// palette has already checked against both surfaces.
func calloutTone(pal pagePalette, tone string) (tint, ink ui.Color) {
	p := palette(pal.dark)
	var hue ui.Color
	switch tone {
	case "good":
		hue = p.good
	case "near":
		hue = p.near
	case "bad":
		hue = p.bad
	default:
		hue = p.accent
	}
	return hue.Mix(pal.paperColor(), 0.9), hue
}
