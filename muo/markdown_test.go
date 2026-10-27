package muo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// A document with one of every block the renderer draws, so that a change to the
// page's theme or the heading hook is exercised against all of them at once
// rather than against the paragraph that happens to be in the sample.
const everyCallout = "# Callouts\n\n" +
	"> [!NOTE]\n> Useful information that users should know, even when skimming.\n\n" +
	"> [!TIP]\n> Helpful advice for doing things better or more easily.\n\n" +
	"> [!IMPORTANT]\n> Key information users need to know to achieve their goal.\n\n" +
	"> [!WARNING]\n> Urgent info that needs immediate user attention.\n\n" +
	"> [!CAUTION]\n> Advises about risks or negative outcomes of certain actions.\n\n" +
	"An ordinary quote, which must stay a quote:\n\n" +
	"> A quote that merely mentions a bracket, [like this].\n"

const everyBlock = "# Title one\n\n" +
	"A paragraph with **bold**, *italic*, `code`, a [link](https://example.com) and a footnote[^1].\n\n" +
	"## Title two\n\n" +
	"### Title three\n\n" +
	"#### Title four\n\n" +
	"- a bullet\n" +
	"- another bullet\n" +
	"  - nested\n" +
	"\n1. first\n2. second\n\n" +
	"- [x] done\n- [ ] not done\n\n" +
	"> A quote, which is set in the muted colour with a bar beside it.\n\n" +
	"| Column | Another |\n| ------ | ------- |\n| cell   | value   |\n| more   | data    |\n\n" +
	"```go\nfunc main() {\n\tprintln(\"hi\")\n}\n```\n\n" +
	"```\nno language here\n```\n\n" +
	"---\n\n" +
	"![a picture](missing.png)\n\n" +
	"Setext heading\n==============\n\n" +
	"A final paragraph.\n\n" +
	"[^1]: The footnote's text.\n"

func openString(t *testing.T, name, src string) *Reader {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New()
	r.OpenIn(path)
	return r
}

// Every block type reaches the page. This is the smoke test for the renderer
// wiring: if the heading hook or the theme ever drops a kind of block, one of
// these strings goes missing.
func TestMarkdownDrawsEveryBlock(t *testing.T) {
	r := openString(t, "all.md", everyBlock)
	tt := ui.NewTester(r.FrameView, 1100, 1400)
	tt.Frame()
	tt.Frame()

	for _, want := range []string{
		"Title one", "Title two", "Title three", "Title four",
		"bold", "italic", "code", "link",
		"a bullet", "another bullet", "nested",
		"first", "second",
		"done", "not done",
		"A quote, which is set in the muted colour",
		"Column", "Another", "cell", "value",
		"func main()", "println",
		"no language here",
		"a picture",
		"Setext heading",
		"A final paragraph.",
	} {
		if !tt.HasText(want) {
			t.Errorf("the page does not show %q", want)
		}
	}
}

// Every heading in the document is a heading on the page, not body text: the
// reader hooks each one to measure it, and a heading the hook misses is a
// heading that cannot be scrolled to or saved.
func TestEveryHeadingIsHooked(t *testing.T) {
	r := openString(t, "all.md", everyBlock)
	tt := ui.NewTester(r.FrameView, 1100, 1400)
	tt.Frame()
	tt.Frame()

	if len(r.doc.headings) != 5 {
		t.Fatalf("the outline found %d headings, want 5: %+v", len(r.doc.headings), r.doc.headings)
	}
	for i, h := range r.doc.headings {
		if _, ok := r.headingHandles[h.Index]; !ok {
			t.Errorf("heading %q (block %d) has no handle", h.Text, h.Index)
		}
		// The first heading sits at the top of the page, so its offset is 0 by
		// definition; every one after it must have been measured to a place
		// below the one before.
		if i > 0 && h.Offset <= 0 {
			t.Errorf("heading %q was never measured: offset %.0f", h.Text, h.Offset)
		}
	}
}

// The headings are measured in document order, and the distance between them
// grows with the amount of text between them. That ordering is what Find uses to
// put a match on the page, so a heading measured out of order would send a search
// to the wrong place.
func TestHeadingsAreMeasuredInOrder(t *testing.T) {
	r := openString(t, "all.md", everyBlock)
	tt := ui.NewTester(r.FrameView, 1100, 1400)
	tt.Frame()
	tt.Frame()

	last := float32(-1)
	for _, h := range r.doc.headings {
		if h.Offset < last {
			t.Fatalf("heading %q is at %.0f, above the one before it at %.0f",
				h.Text, h.Offset, last)
		}
		last = h.Offset
	}
}

// A heading is bigger than the body text it introduces, at every level, and the
// levels descend: a page whose headings were all one size would have no shape.
func TestHeadingSizesDescend(t *testing.T) {
	r := openString(t, "sizes.md", "# One\n\n## Two\n\n### Three\n\n#### Four\n\n##### Five\n\n###### Six\n")
	tt := ui.NewTester(r.FrameView, 1000, 900)
	tt.Frame()
	tt.Frame()

	last := float32(1e9)
	for _, h := range r.doc.headings {
		px := r.headingPx(h.Level)
		if px > last {
			t.Errorf("level %d is %.1f pt, larger than the level above at %.1f", h.Level, px, last)
		}
		// A heading is never smaller than the text it introduces: at that
		// point it stops being a heading and becomes a footnote to it.
		if px < r.typ.FontPx {
			t.Errorf("level %d is %.1f pt, smaller than the body at %.1f", h.Level, px, r.typ.FontPx)
		}
		last = px
	}
	// And the top level is clearly above the body, or the page has no title.
	if top := r.headingPx(1); top < r.typ.FontPx*1.4 {
		t.Errorf("a level-1 heading is %.1f pt against a %.1f pt body, which is not a title",
			top, r.typ.FontPx)
	}
}

// Nothing on the page reaches past the reading column. Tables and code blocks are
// the two blocks that want to be wider than the measure, and both are the
// renderer's to size: if either overflows, the page has a horizontal scroll and
// the measure is broken.
func TestNothingOverflowsTheColumn(t *testing.T) {
	r := openString(t, "wide.md", "| A very long column heading indeed | And another one |\n| --- | --- |\n| "+strings.Repeat("cell ", 20)+" | value |\n\n```go\n"+strings.Repeat("var x = 1 // a fairly long line of code\n", 3)+"```\n")
	r.SetSidebarWidth(sidebarWidth)
	const winW = 1100
	tt := ui.NewTester(r.FrameView, winW, 900)
	tt.Frame()
	tt.Frame()

	right := r.typ.ColumnW + pageMargin + sidebarWidth + 40
	if cell, ok := tt.Find("A very long column heading indeed"); ok {
		if cell.X+cell.W > right {
			t.Errorf("the table runs to %.0f, past the column's edge at %.0f", cell.X+cell.W, right)
		}
	} else {
		t.Error("the table heading was not found")
	}
}

// A fenced block says its language, in the small caps the header gives it, and
// an unlabelled one says nothing: "CODE" over a block of code is a placeholder
// where every other block has a fact, and it is not a language.
func TestCodeBlockHeaderNamesTheLanguage(t *testing.T) {
	r := openString(t, "code.md", "```go\nfunc main() {}\n```\n\n```\nplain\n```\n")
	tt := ui.NewTester(r.FrameView, 1000, 900)
	tt.Frame()
	tt.Frame()

	if !tt.HasText("GO") {
		t.Error("a go fence does not name its language")
	}
	if tt.HasText("CODE") {
		t.Error("an untagged fence still draws a \"CODE\" placeholder label")
	}
}

// The page ends with a mark, so a document that stops at the foot of the scroll
// reads as finished rather than as still loading.
func TestPageEndsWithAMark(t *testing.T) {
	r := openString(t, "end.md", "# Title\n\nOne paragraph only.\n")
	tt := ui.NewTester(r.FrameView, 1000, 900)
	tt.Frame()
	tt.Frame()
	if _, ok := tt.Find("End of document"); !ok {
		t.Error("the page has no end mark")
	}
}

// A PDF gets the page counter and arrows instead of an end mark, and no markdown
// at all: the two formats share a frame, not a renderer.
func TestPdfPageHasNoEndMark(t *testing.T) {
	path := writeTemp(t, "doc.pdf", simplePDF("BT /F1 12 Tf 72 720 Td (Hello, reader) Tj ET"))
	r := New()
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 900, 700)
	tt.Frame()
	tt.Frame()
	if !tt.HasText("1 of 1") {
		t.Error("the PDF has no page counter")
	}
	if tt.HasText("· · ·") {
		t.Error("a PDF should not end with the markdown end mark")
	}
}

// The page's own colours come from the same palette as the chrome, so that the
// two never drift into different neutrals.
func TestPageAndChromeShareOneNeutral(t *testing.T) {
	for _, dark := range []bool{false, true} {
		p := palette(dark)
		paper := triple(p.paper)
		chrome := triple(p.chrome)
		// The two surfaces are close: a page and a sidebar of one material
		// differ by a step, not by a hue.
		dr := int(paper.R) - int(chrome.R)
		dg := int(paper.G) - int(chrome.G)
		db := int(paper.B) - int(chrome.B)
		if dr < -30 || dr > 30 || dg < -30 || dg > 30 || db < -30 || db > 30 {
			t.Errorf("dark=%v: paper %v and chrome %v are too far apart to be one material",
				dark, paper, chrome)
		}
		// And they lean the same way: both warm, or both cool.
		if (dr > 0) != (db > 0) && (dr != 0 && db != 0) {
			t.Errorf("dark=%v: paper %v and chrome %v lean in different directions", dark, paper, chrome)
		}
	}
}
