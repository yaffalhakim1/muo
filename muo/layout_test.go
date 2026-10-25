package muo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The window comes back as the reader left it. This is the bug it fixes: the
// widths were kept but the panes' visibility was not, so folding a pane away and
// reopening the app brought it back — the window forgetting a decision the reader
// had made about how they want to read.
func TestLayoutSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	// A reader who folds the library away, keeps the outline, and widens it.
	r := New()
	r.showLibrary = false
	r.showOutline = true
	r.sidebarW = 300
	r.outlineW = 260
	if err := r.SaveLayout(); err != nil {
		t.Fatalf("saving the layout: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, layoutFile)); err != nil {
		t.Fatalf("no layout file was written: %v", err)
	}

	// A fresh reader, as after a restart.
	layoutMu.Lock()
	layoutLoaded, layoutCache = false, layoutState{}
	layoutMu.Unlock()

	r2 := New()
	if r2.showLibrary {
		t.Error("the library came back open after being folded away")
	}
	if !r2.showOutline {
		t.Error("the outline was closed after being left open")
	}
	if r2.sidebarW != 300 {
		t.Errorf("library width = %.0f, want 300", r2.sidebarW)
	}
	if r2.outlineW != 260 {
		t.Errorf("outline width = %.0f, want 260", r2.outlineW)
	}
}

// A layout saved before the panes' visibility was kept has no answer for it, so
// the panes open rather than disappearing: an old file must not make the app
// look broken.
func TestOldLayoutOpensBothPanes(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	old := `{"sidebarWidth": 250, "outlineWidth": 200}`
	if err := os.WriteFile(filepath.Join(dir, layoutFile), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	layoutMu.Lock()
	layoutLoaded, layoutCache = false, layoutState{}
	layoutMu.Unlock()

	r := New()
	if !r.showLibrary || !r.showOutline {
		t.Errorf("an old layout closed a pane: library=%v outline=%v",
			r.showLibrary, r.showOutline)
	}
	if r.sidebarW != 250 || r.outlineW != 200 {
		t.Errorf("the widths were not read: %.0f and %.0f", r.sidebarW, r.outlineW)
	}
}

// What is written says both, so the file can be read by a person.
func TestLayoutFileRecordsVisibility(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	r := New()
	r.showLibrary = false
	r.showOutline = true
	if err := r.SaveLayout(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, layoutFile))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Library *bool `json:"library"`
		Outline *bool `json:"outline"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("the file is not valid JSON: %v\n%s", err, data)
	}
	if got.Library == nil || *got.Library {
		t.Errorf("library = %v, want false", got.Library)
	}
	if got.Outline == nil || !*got.Outline {
		t.Errorf("outline = %v, want true", got.Outline)
	}
}
