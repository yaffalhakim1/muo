package muo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// The window is driven headlessly: the tester builds frames and reads what is
// on them, so what is checked here is what the window actually shows.

func newTestWindow(t *testing.T, r *Reader, w, h int) *ui.Tester {
	t.Helper()
	return ui.NewTester(r.FrameView, w, h)
}

func TestViewShowsEmptyState(t *testing.T) {
	r := New()
	tt := newTestWindow(t, r, 900, 700)
	if !tt.HasText("Muo") {
		t.Fatal("the empty state does not name the reader")
	}
	if !tt.HasText("Open a File") {
		t.Fatal("the empty state offers no way in")
	}
}

func TestViewShowsMarkdown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.md")
	source := "# Reading\n\nA paragraph long enough to be a paragraph, with **bold** words in it.\n\n- one\n- two\n"
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New()
	r.OpenIn(path)
	tt := newTestWindow(t, r, 900, 700)
	for _, want := range []string{"Reading", "paragraph", "bold", "one", "two"} {
		if !tt.HasText(want) {
			t.Errorf("the page does not show %q", want)
		}
	}
}

func TestViewShowsPdfPageCounter(t *testing.T) {
	path := writeTemp(t, "doc.pdf", simplePDF("BT /F1 12 Tf 72 720 Td (Hello, reader) Tj ET"))
	r := New()
	r.OpenIn(path)
	tt := newTestWindow(t, r, 900, 700)
	if !tt.HasText("1 of 1") {
		t.Fatal("the page counter is missing")
	}
}

func TestViewExplainsUnreadablePdf(t *testing.T) {
	path := writeTemp(t, "broken.pdf", []byte("%PDF-1.7\nbroken"))
	r := New()
	r.OpenIn(path)
	tt := newTestWindow(t, r, 900, 700)
	if !tt.HasText("couldn't be read") {
		t.Fatal("an unreadable file should say so")
	}
}

func TestSettingsSurfaceOpensWithTheModel(t *testing.T) {
	r := New()
	r.OpenIn(writeTemp(t, "doc.md", []byte("# Title\n")))
	r.settings = true
	tt := newTestWindow(t, r, 900, 700)
	if !tt.HasText("Reading") {
		t.Fatal("the settings surface did not open")
	}
	if !tt.HasText("Reading") {
		t.Fatal("the settings surface has no categories")
	}
}

func TestTypeReadoutFormat(t *testing.T) {
	// A window wide enough for the full measure holds 72 characters; a narrow
	// one gives way in characters, which is the point of capping by count.
	wide := Model(Display{Distance: 65}, 1600, 109, "", 0)
	if wide.ChPerLine < 71 || wide.ChPerLine > 73 {
		t.Fatalf("measure = %.1f, want 72", wide.ChPerLine)
	}
	if wide.Comfort != ComfortGood {
		t.Fatalf("comfort = %v, want good", wide.Comfort)
	}
	narrow := Model(Display{Distance: 65}, 900, 109, "", 0)
	if narrow.ChPerLine >= 72 {
		t.Fatalf("a 900 pt window should not claim 72 characters, got %.1f", narrow.ChPerLine)
	}
	if narrow.ColumnW > 900 {
		t.Fatalf("the column overflows the window: %.0f", narrow.ColumnW)
	}
}



