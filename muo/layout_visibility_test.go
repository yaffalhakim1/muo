package muo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// The window comes back as the reader left it, checked end to end: write a layout
// with the library folded, build a reader from it, and look at what the window
// actually draws.
func TestFoldedLibraryStaysFolded(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	// A layout as the app would have written it: the reader folded the library.
	folded := `{"sidebarWidth": 300, "outlineWidth": 220, "library": false, "outline": true}`
	if err := os.WriteFile(filepath.Join(dir, layoutFile), []byte(folded), 0o644); err != nil {
		t.Fatal(err)
	}
	layoutMu.Lock()
	layoutLoaded, layoutCache = false, layoutState{}
	layoutMu.Unlock()

	r := New()
	if r.showLibrary {
		t.Fatal("the library is open, though the saved layout has it folded")
	}
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("# Title\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 1440, 800)
	tt.Frame()
	tt.Frame()

	if _, ok := tt.Find("Open a File"); ok {
		t.Error("the library's open command is on screen though the pane is folded")
	}
	if _, ok := tt.Find("On This Page"); !ok {
		t.Error("the outline is missing though the saved layout has it open")
	}
}

// And the other way: a reader who folded the outline but kept the library.
func TestFoldedOutlineStaysFolded(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	folded := `{"sidebarWidth": 236, "outlineWidth": 220, "library": true, "outline": false}`
	if err := os.WriteFile(filepath.Join(dir, layoutFile), []byte(folded), 0o644); err != nil {
		t.Fatal(err)
	}
	layoutMu.Lock()
	layoutLoaded, layoutCache = false, layoutState{}
	layoutMu.Unlock()

	r := New()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("# Title\n\nbody\n\n## Section\n\nmore\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 1440, 800)
	tt.Frame()
	tt.Frame()

	if _, ok := tt.Find("Open a File"); !ok {
		t.Error("the library is missing though the saved layout has it open")
	}
	if _, ok := tt.Find("On This Page"); ok {
		t.Error("the outline is on screen though the saved layout has it folded")
	}
}

// The saved layout is the reader's own state, so a test never reads the
// developer's file: this checks that a fresh test directory starts from the
// default rather than from whatever window is open on the machine.
func TestFreshLayoutStartsOpen(t *testing.T) {
	useStateDir(t, t.TempDir())
	r := New()
	if !r.showLibrary || !r.showOutline {
		t.Errorf("a fresh layout did not start with both panes open: library=%v outline=%v",
			r.showLibrary, r.showOutline)
	}
	var got layoutState
	if err := json.Unmarshal([]byte(`{"sidebarWidth": 250}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.Library != nil || got.Outline != nil {
		t.Error("a layout without visibility should read as not stated, not as closed")
	}
}
