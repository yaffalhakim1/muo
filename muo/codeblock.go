package muo

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"

	"example.com/muo/muo/code"
)

// Code blocks can be formatted in place. The renderer re-reads the document's
// source every frame, so formatting means rewriting that block's text in the
// source: the renderer then highlights and lays out the formatted code exactly
// as it lays out any other, and Copy copies what is shown.
//
// The block carries a header of its own, but a quiet one: the language in small
// caps on the left, the two commands on the right, and no rule between them and
// the code. A header with a border above a bordered block is a second box
// inside the first, which is heavier than the thing it introduces.

// codeBlockView wraps the renderer's default block, adding a Format action
// beside its language label and Copy button.
func (r *Reader) codeBlockView(c *ui.Context, b *markdown.CodeBlockView) {
	if b.Open {
		// A block still streaming has no closing fence, so there is nothing
		// stable to format yet.
		markdown.DefaultCodeBlock(c, b)
		return
	}
	t := c.Theme()
	th := r.md.Theme

	box := ui.Column(c).AlignSelf(ui.Stretch).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(t.Space(0.5)).
			Padding(t.Space(1.5), t.Space(2), t.Space(1), t.Space(3)).
			Children(func() {
				// A fence the author did not tag gets no label at all. Writing
				// "CODE" over a block of code says nothing the block does not
				// already say, and it puts a placeholder where every other
				// block has a fact.
				if b.Language != "" {
					ui.Text(c, strings.ToUpper(b.Language)).
						FontSize(typeFootnote).FontWeight(600).LetterSpacing(0.5).
						TextColor(t.TextMuted).Grow(1).SingleLine()
				} else {
					ui.Spacer(c)
				}
				// The commands are quiet: they carry no face of their own, so
				// the block reads as one surface with a label rather than as a
				// box with a toolbar screwed to the top of it.
				//
				// Format is built only for a language the app can tidy. A
				// disabled button that can never be pressed is not a command,
				// and it is not a hint either: it is a thing on the page that
				// does nothing, on a block the reader did not ask about.
				if code.HasFormatter(b.Language) {
					format := ui.Button(c, "Format").FontSize(typeCallout).
						Padding(2, 8).Radius(5).
						Background(ui.Transparent).BorderWidth(0).
						TextColor(t.TextMuted)
					if format.Hovered() {
						format.Background(t.SurfaceHover)
					}
					if format.Clicked() {
						r.formatBlock(c, b)
					}
				}
				// The renderer's own Copy button, so the copy path stays its own.
				b.CopyButton(c)
			})
		b.Body(c)
	})
	box.Background(th.CodeBlockBackground).
		Border(1, th.CodeBlockBorder).
		Radius(th.CodeBlockRadius).Clip()
}

// formatBlock rewrites one fenced block in the document's source with the
// formatted code, keeping the fence and its language.
//
// It says what happened, because a button that appears to do nothing is
// indistinguishable from one that is broken. Markup and prose have no
// formatter, and a block already laid out has nothing to change; both say so
// rather than leaving the reader pressing a button and guessing.
func (r *Reader) formatBlock(c *ui.Context, b *markdown.CodeBlockView) {
	if r.doc == nil {
		return
	}
	if !code.HasFormatter(b.Language) {
		c.Toast("No formatter for this language")
		return
	}
	raw := b.Code()
	formatted, changed := code.Format(b.Language, raw)
	if !changed {
		c.Toast("Already formatted")
		return
	}
	src := r.doc.Source
	if !strings.Contains(src, raw) {
		// The block the reader sees is not in the source as written: the
		// document changed under it. Say so instead of rewriting the wrong
		// block, which is how a formatter eats someone's code.
		c.Toast("This block no longer matches the document")
		return
	}
	r.doc.Source = strings.Replace(src, raw, formatted, 1)
	c.Toast("Formatted")
}
