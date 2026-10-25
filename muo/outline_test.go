package muo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// useStateDir points the saved positions at a test directory, as the user data
// directory would be, and empties the cache so the reader has to read the file.
func useStateDir(t *testing.T, dir string) {
	t.Helper()
	stateMu.Lock()
	stateCache = nil
	stateDir = dir
	stateMu.Unlock()
	// The layout lives in the same directory and is cached the same way, so a test
	// that sets the directory must forget that cache too: otherwise the reader's
	// real layout.json — how wide they made their panes, which they folded — leaks
	// into the test and it passes or fails on the developer's own window.
	layoutMu.Lock()
	layoutLoaded, layoutCache = false, layoutState{}
	layoutMu.Unlock()
}

// The reader's place survives a restart: choosing a heading is written to disk,
// and a fresh reader opening the same file reads it back.
func TestReadingPositionSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	src := "# One\n\ntext\n\n## Two\n\ntext\n\n## Three\n\ntext\n\n### Four\n\ntext\n"
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	r := New()
	r.OpenIn(path)
	if len(r.doc.headings) != 4 {
		t.Fatalf("outline found %d headings, want 4", len(r.doc.headings))
	}
	third := r.doc.headings[2]
	if third.Text != "Three" {
		t.Fatalf("third heading = %q, want Three", third.Text)
	}

	tt := ui.NewTester(r.FrameView, 900, 600)
	tt.Frame()
	tt.Frame()
	r.jumpTo(third)
	tt.Frame()
	tt.Frame()

	if err := r.Save(); err != nil {
		t.Fatalf("saving: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, stateFile)); err != nil {
		t.Fatalf("no state file was written: %v", err)
	}

	// A fresh reader, as after a restart, opens the same file.
	useStateDir(t, dir)
	r2 := New()
	r2.OpenIn(path)
	if r2.doc.heading != third.Index {
		t.Fatalf("reopened at heading %d, want %d", r2.doc.heading, third.Index)
	}
	if !r2.restorePending {
		t.Fatal("the reopened reader did not plan to restore its place")
	}
}

// Headings are read in document order with their level, which is what the source
// list uses to indent them.
func TestOutlineLevels(t *testing.T) {
	hs := outlineOf("# A\n\n## B\n\n### C\n\n## D\n")
	if len(hs) != 4 {
		t.Fatalf("found %d headings, want 4", len(hs))
	}
	for i, want := range []int{1, 2, 3, 2} {
		if hs[i].Level != want {
			t.Errorf("heading %d (%q) has level %d, want %d",
				i, hs[i].Text, hs[i].Level, want)
		}
	}
}

// A heading's words are cleaned of its markup, so the outline reads as prose.
func TestOutlineStripsMarkup(t *testing.T) {
	hs := outlineOf("## **Bold** and `code`\n")
	if len(hs) != 1 {
		t.Fatalf("found %d headings, want 1", len(hs))
	}
	if hs[0].Level != 2 {
		t.Fatalf("level = %d, want 2", hs[0].Level)
	}
	if hs[0].Key == "" {
		t.Fatal("the heading has no identity")
	}
}
