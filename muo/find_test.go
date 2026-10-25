package muo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Find, for both formats. A Markdown page and a PDF are searched by different
// machinery — source offsets against rendered blocks, and page text against page
// positions — so each gets the same smoke test.

func settle(tt *ui.Tester) {
	tt.Advance(700 * time.Millisecond)
	tt.Frame()
}

// ---------------------------------------------------------------------------
// Markdown

// A long page, so that matches are off screen and scrolling has somewhere to go.
func longMarkdown(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Alpha\n\n")
	b.WriteString(strings.Repeat("Filler without the word. ", 40) + "\n\n")
	b.WriteString("## Beta\n\nneedle one here.\n\n")
	b.WriteString(strings.Repeat("More filler text. ", 60) + "\n\n")
	b.WriteString("## Gamma\n\nneedle two here.\n")
	path := filepath.Join(t.TempDir(), "long.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// A long page with no headings at all: this is the case the old
// heading-interpolation could not place, because it had no anchors.
func longMarkdownNoHeadings(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	b.WriteString(strings.Repeat("Filler without the word. ", 40) + "\n\n")
	b.WriteString("needle one here.\n\n")
	b.WriteString(strings.Repeat("More filler text. ", 60) + "\n\n")
	b.WriteString("needle two here.\n")
	path := filepath.Join(t.TempDir(), "flat.md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func openMarkdown(t *testing.T, path string) (*Reader, *ui.Tester) {
	t.Helper()
	r := New()
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 700, 400)
	tt.Frame()
	tt.Frame()
	return r, tt
}

// Find scrolls to a match that is off screen, on a page with headings.
func TestFindScrollsToOffscreenMatchMarkdown(t *testing.T) {
	r, tt := openMarkdown(t, longMarkdown(t))
	if r.contentScroll.MaxY <= 0 {
		t.Fatalf("the page does not scroll, so this cannot be tested (maxY=%.0f)", r.contentScroll.MaxY)
	}
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	if r.contentScroll.Y <= 0 {
		t.Fatalf("the page did not move to the first match (Y=%.0f)", r.contentScroll.Y)
	}
	first := r.contentScroll.Y

	// The second match is further down.
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if r.match != 1 {
		t.Fatalf("Next left the match at %d, want 1", r.match)
	}
	if r.contentScroll.Y <= first {
		t.Errorf("the page did not move to the second match: %.0f then %.0f", first, r.contentScroll.Y)
	}
}

// And on a page with no headings, which used to land at a guess.
func TestFindScrollsWithoutHeadings(t *testing.T) {
	r, tt := openMarkdown(t, longMarkdownNoHeadings(t))
	if r.contentScroll.MaxY <= 0 {
		t.Fatalf("the page does not scroll (maxY=%.0f)", r.contentScroll.MaxY)
	}
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	if r.contentScroll.Y <= 0 {
		t.Fatalf("the page did not move to a match on a page with no headings (Y=%.0f)", r.contentScroll.Y)
	}
	first := r.contentScroll.Y
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if r.contentScroll.Y <= first {
		t.Errorf("the page did not move to the second match: %.0f then %.0f", first, r.contentScroll.Y)
	}
}

// The Previous button goes back, and the page follows it.
func TestFindPreviousButtonMarkdown(t *testing.T) {
	r, tt := openMarkdown(t, longMarkdown(t))
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	second := r.contentScroll.Y
	if err := tt.Click("Previous"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if r.match != 0 {
		t.Fatalf("Previous left the match at %d, want 0", r.match)
	}
	if r.contentScroll.Y >= second {
		t.Errorf("the page did not move back: %.0f then %.0f", second, r.contentScroll.Y)
	}
}

// Every match is marked, including one off screen.
func TestFindMarksOffscreenMatchesMarkdown(t *testing.T) {
	r, tt := openMarkdown(t, longMarkdown(t))
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	if n := strings.Count(r.highlightedSource(), markOpen); n != 2 {
		t.Errorf("%d matches marked, want 2", n)
	}
}

// ---------------------------------------------------------------------------
// PDF

// A PDF with text on it: the fixture writes Helvetica at the positions given.
func pdfWithText(t *testing.T, lines ...string) string {
	t.Helper()
	var b strings.Builder
	for i, line := range lines {
		// Each line is a text object at its own baseline, 24 points apart.
		y := 720 - i*24
		b.WriteString("BT /F1 12 Tf 72 ")
		b.WriteString(itoa(y))
		b.WriteString(" Td (")
		b.WriteString(line)
		b.WriteString(") Tj ET\n")
	}
	return writeTemp(t, "text.pdf", simplePDF(b.String()))
}

func openPDF(t *testing.T, path string) (*Reader, *ui.Tester) {
	t.Helper()
	r := New()
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 900, 700)
	tt.Frame()
	tt.Frame()
	return r, tt
}

// A PDF is searched by its text, which the interpreter reads off the page.
func TestFindCountsPDFMatches(t *testing.T) {
	path := pdfWithText(t,
		"a line with no match",
		"the needle is here",
		"another line",
		"the needle again",
	)
	r, tt := openPDF(t, path)
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	if n := len(r.matches()); n != 2 {
		t.Fatalf("found %d matches in the PDF, want 2", n)
	}
	if !tt.HasText("1 of 2") {
		t.Errorf("the bar does not say 1 of 2; texts: %q", tt.Texts())
	}
}

// A PDF with no text has no matches, and says so.
func TestFindPDFWithNoText(t *testing.T) {
	path := writeTemp(t, "blank.pdf", simplePDF("0 0 m 100 100 l S"))
	r, tt := openPDF(t, path)
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	if n := len(r.matches()); n != 0 {
		t.Fatalf("found %d matches in a page with no text", n)
	}
	if !tt.HasText("No matches") {
		t.Errorf("the bar does not say No matches; texts: %q", tt.Texts())
	}
}

// The Next button walks a PDF's matches, and the page follows each one.
func TestFindNextButtonPDF(t *testing.T) {
	var b strings.Builder
	b.WriteString("BT /F1 12 Tf 72 760 Td (needle near the top) Tj ET\n")
	b.WriteString("BT /F1 12 Tf 72 120 Td (needle near the bottom) Tj ET\n")
	path := writeTemp(t, "two.pdf", simplePDF(b.String()))

	r := New()
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 520, 300)
	tt.Frame()
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	if n := len(r.matches()); n != 2 {
		t.Fatalf("found %d matches, want 2", n)
	}
	if r.contentScroll.MaxY <= 0 {
		t.Fatalf("the PDF page does not scroll (maxY=%.0f)", r.contentScroll.MaxY)
	}
	if r.match != 0 {
		t.Fatalf("the first match should be current, got %d", r.match)
	}

	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if r.match != 1 {
		t.Fatalf("Next left the match at %d, want 1", r.match)
	}
	if r.contentScroll.Y <= 0 {
		t.Errorf("the page did not scroll to the lower match (Y=%.0f, target=%.0f)",
			r.contentScroll.Y, r.contentScrollAt)
	}
}
// Previous goes back to the earlier match.
func TestFindPreviousButtonPDF(t *testing.T) {
	var b strings.Builder
	b.WriteString("BT /F1 12 Tf 72 760 Td (needle near the top) Tj ET\n")
	b.WriteString("BT /F1 12 Tf 72 120 Td (needle near the bottom) Tj ET\n")
	path := writeTemp(t, "two.pdf", simplePDF(b.String()))

	r, tt := openPDF(t, path)
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	r.match = 1
	tt.Frame()
	tt.Frame()
	low := r.contentScroll.Y

	if err := tt.Click("Previous"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if r.match != 0 {
		t.Fatalf("Previous left the match at %d, want 0", r.match)
	}
	if r.contentScroll.Y >= low {
		t.Errorf("the page did not move back up: %.0f then %.0f", low, r.contentScroll.Y)
	}
}

// Enter and the buttons agree: both walk the same list.
func TestFindEnterPDF(t *testing.T) {
	var b strings.Builder
	b.WriteString("BT /F1 12 Tf 72 760 Td (needle one) Tj ET\n")
	b.WriteString("BT /F1 12 Tf 72 120 Td (needle two) Tj ET\n")
	path := writeTemp(t, "enter.pdf", simplePDF(b.String()))

	r, tt := openPDF(t, path)
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if r.match != 1 {
		t.Errorf("Enter left the match at %d, want 1", r.match)
	}
	// Round the end, back to the first.
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if r.match != 0 {
		t.Errorf("Enter at the last match left it at %d, want 0", r.match)
	}
}

// A match on another page turns to that page.
func TestFindPDFTurnsThePage(t *testing.T) {
	// A two-page file: the first page has no match, the second does.
	path := writeTemp(t, "two-pages.pdf", twoPagePDF(
		"BT /F1 12 Tf 72 720 Td (nothing here) Tj ET",
		"BT /F1 12 Tf 72 720 Td (the needle is on this page) Tj ET",
	))
	r, tt := openPDF(t, path)
	if r.file == nil || r.file.PageCount() != 2 {
		t.Fatalf("the fixture is not two pages: %v", r.file)
	}
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("needle")
	tt.Frame()
	tt.Frame()
	if n := len(r.matches()); n != 1 {
		t.Fatalf("found %d matches, want 1", n)
	}
	if r.viewing != 1 {
		t.Errorf("the view is on page %d, want the second page", r.viewing)
	}
}

// twoPagePDF builds a two-page file whose pages have the given content.
func twoPagePDF(first, second string) []byte {
	w := newPDFWriter()
	w.obj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	w.obj(2, "<< /Type /Pages /Kids [3 0 R 5 0 R] /Count 2 >>")
	w.obj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 7 0 R >> >> >>")
	w.stream(4, "", []byte(first))
	w.obj(5, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 6 0 R /Resources << /Font << /F1 7 0 R >> >> >>")
	w.stream(6, "", []byte(second))
	w.obj(7, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /FirstChar 32 /LastChar 126 /Widths ["+helveticaWidths()+"] >>")
	return w.finish(1)
}
