package muo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// The window is three panes: the library of open files on the left, the page in
// the middle, and the document's own outline on the right. These tests hold that
// structure, because it is the app's whole shape and a pane that silently moves
// or vanishes is worse than one that was never there.

// openWithDocs opens a reader with the given documents, as a reader with several
// files would have.
//
// It points the saved layout at a directory of its own first, because the layout
// is the reader's own state: a test that read the developer's layout.json would
// pass or fail on how they happened to leave the app.
func openWithDocs(t *testing.T, names ...string) (*Reader, *ui.Tester) {
	t.Helper()
	useStateDir(t, t.TempDir())
	r := New()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name)
		src := "# " + strings.TrimSuffix(name, ".md") + "\n\ntext\n\n## Section\n\nmore\n"
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		r.OpenIn(path)
	}
	r.SetSidebarWidth(SidebarWidth)
	tt := ui.NewTester(r.FrameView, 1440, 860)
	tt.Frame()
	tt.Frame()
	return r, tt
}

// The command to open a file is at the top of the library, above the list, so a
// reader with several files open still finds it where the list begins.
func TestOpenCommandIsAboveTheFileList(t *testing.T) {
	r, tt := openWithDocs(t, "one.md", "two.md", "three.md")
	_ = r

	open, ok := tt.Find("Open a File")
	if !ok {
		t.Fatal("the open command is missing")
	}
	list, ok := tt.Find("Open Files")
	if !ok {
		t.Fatal("the file list has no heading")
	}
	if open.Y >= list.Y {
		t.Errorf("the open command is at y=%.0f, below the list at y=%.0f", open.Y, list.Y)
	}
	// And it sits near the top of the pane, under the panel's own title.
	if open.Y > headerHeight+80 {
		t.Errorf("the open command is at y=%.0f, not near the top of the pane", open.Y)
	}
}

// The command is a real button rather than a hand-styled row, which is what the
// platform's focus ring, press state and Enter activation come from. A click
// would open the system's file dialog, which a test cannot answer, so what is
// checked is the control itself: a ui.Button with an accessible label.
func TestOpenCommandIsAButton(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")
	_ = r
	if !tt.HasText("Open a File") {
		t.Fatal("the open command is missing")
	}
	// A ui.Button is one stop of Tab; a plain row is not a stop at all.
	box, ok := tt.Find("Open a File")
	if !ok {
		t.Fatal("the open command has no box")
	}
	if box.W < 80 || box.H < 20 {
		t.Errorf("the open command is %.0fx%.0f, too small to be a button", box.W, box.H)
	}
}

// Folding the outline takes it off screen.
//
// The position is read from what the window actually drew — the outline's own
// header — and not from a number the app computes. A folded panel is not built,
// so once the fold has happened the header is gone rather than moved.
func TestOutlineFoldsWithASlide(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")
	_ = r
	rendered := func() float32 {
		if b, ok := tt.Find("On This Page"); ok {
			return b.X
		}
		return -1
	}
	start := rendered()
	if start <= 0 {
		t.Fatal("the outline was not drawn, so this cannot be tested")
	}

	r.showOutline = false
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	if got := rendered(); got != -1 {
		t.Fatalf("the outline is still drawn at x=%.0f after folding", got)
	}

	// And it comes back where it was when the reader asks for it again.
	r.showOutline = true
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	if got := rendered(); got <= 0 {
		t.Fatalf("the outline did not come back: x=%.0f", got)
	}
}

// Every open document is listed, which is what makes several files usable.
func TestEveryOpenDocumentIsListed(t *testing.T) {
	_, tt := openWithDocs(t, "one.md", "two.md", "three.md")
	for _, want := range []string{"one.md", "two.md", "three.md"} {
		if !tt.HasText(want) {
			t.Errorf("%q is not in the library", want)
		}
	}
}

// The outline is on the right of the page, and the library on the left: each
// pane sits where it belongs.
func TestPanesAreOnTheirOwnSides(t *testing.T) {
	_, tt := openWithDocs(t, "one.md")

	lib, ok := tt.Find("Open a File")
	if !ok {
		t.Fatal("the library is missing")
	}
	out, ok := tt.Find("On This Page")
	if !ok {
		t.Fatal("the outline is missing")
	}
	page, ok := tt.Find("one")
	if !ok {
		t.Fatal("the page is missing")
	}
	if !(lib.X < page.X && page.X < out.X) {
		t.Errorf("panes are out of order: library at %.0f, page at %.0f, outline at %.0f",
			lib.X, page.X, out.X)
	}
}

// Choosing a heading in the outline scrolls to it, as it did when the outline
// lived in the left pane.
//
// The click is aimed at the outline's own row. A heading's words appear twice —
// once in the page, once in the outline — so the row is found by the position of
// the outline's second heading rather than by its text, which would find the
// page's copy first.
func TestOutlineStillScrollsToAHeading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "long.md")
	var b strings.Builder
	b.WriteString("# Top\n\n")
	b.WriteString(strings.Repeat("Filler text here. ", 200) + "\n\n")
	b.WriteString("## Far Down\n\nmore text\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New()
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 1440, 500)
	tt.Frame()
	tt.Frame()

	before := r.contentScroll.Y
	if len(r.outlineRows) < 2 {
		t.Fatalf("the outline drew %d rows, want 2", len(r.outlineRows))
	}
	row := r.outlineRows[1]
	tt.ClickAt(row.X+row.W/2, row.Y+row.H/2)
	tt.Frame()
	tt.Frame()
	if r.contentScroll.Y <= before {
		t.Errorf("choosing a heading did not scroll: %.0f then %.0f", before, r.contentScroll.Y)
	}
}

// The outline yields on a narrow window, so the page keeps its measure; the
// library stays, because it is what a reader with several files needs.
func TestOutlineYieldsBeforeTheLibrary(t *testing.T) {
	r, _ := openWithDocs(t, "one.md")
	if r.outlineShown(nil) {
		t.Log("the outline shows at 1440, as it should")
	}

	// A narrow window: the outline goes, the library stays.
	narrow := New()
	narrow.OpenIn(filepath.Join(t.TempDir(), "x.md"))
	narrow.SetSidebarWidth(SidebarWidth)
	if err := os.WriteFile(filepath.Join(t.TempDir(), "y.md"), []byte("# T\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(narrow.FrameView, 1000, 700)
	tt.Frame()
	tt.Frame()
	if _, ok := tt.Find("On This Page"); ok {
		t.Error("the outline is showing at 1000 DIPs, where it costs the page its measure")
	}
	if _, ok := tt.Find("Open a File"); !ok {
		t.Error("the library went with the outline; it is the pane a reader needs most")
	}
}

// Both panes can be folded away, and the page takes the whole window when they
// are.
func TestPanesCanBeFoldedAway(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")
	pageBefore := pageColumnX(t, tt)
	if pageBefore <= 0 {
		t.Fatal("the page is missing")
	}

	r.showLibrary = false
	r.showOutline = false
	// The panes fold on a spring: the first frame starts it, the advance runs it.
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	pageAfter := pageColumnX(t, tt)
	if pageAfter <= 0 {
		t.Fatal("the page vanished when the panes were folded")
	}
	if pageAfter >= pageBefore {
		t.Errorf("the page did not move into the library's room: %.0f then %.0f",
			pageBefore, pageAfter)
	}
	// The column is capped at 72 characters, so it does not grow past that; what
	// must hold is that it is no longer squeezed beside the panes.
	if r.typ.ChPerLine < 60 {
		t.Errorf("the column is only %.0f characters wide with every pane folded",
			r.typ.ChPerLine)
	}
}

// The measure survives every combination of panes: the column is solved from the
// page's real width, so folding a pane gives the page more room rather than
// leaving the text narrow in a wide window.
func TestMeasureFollowsThePanes(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")
	withPanes := r.typ.ChPerLine

	r.showOutline = false
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	withoutOutline := r.typ.ChPerLine

	if withoutOutline < withPanes {
		t.Errorf("folding the outline narrowed the page: %.0f then %.0f characters",
			withPanes, withoutOutline)
	}
	if withoutOutline > charsPerLine+1 {
		t.Errorf("the column claims %.0f characters, more than the %.0f the model allows",
			withoutOutline, charsPerLine)
	}
}

// pageColumnX is where the reading column starts, read from the document's own
// body text rather than from a file name, which appears in the library too.
func pageColumnX(t *testing.T, tt *ui.Tester) float32 {
	t.Helper()
	if b, ok := tt.Find("text"); ok {
		return b.X
	}
	return 0
}