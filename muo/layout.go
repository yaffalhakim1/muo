package muo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/egoist/mygo"
)

// The window's layout, as opposed to the documents' places: how wide the reader
// made each pane, and which panes they had open. It is kept beside
// reading-state.json, in the app's user data directory, so a reader who widened
// the library for long file names — or folded it away to read — finds the window
// as they left it.
//
// The visibility is saved as well as the width, and that is the point: a reader
// who folds a pane away has made a decision about how they want to read, and
// reopening the window to find it back is the window forgetting.
//
// It is written when the window closes, not as it is dragged: a file write per
// frame of a drag would be hundreds of writes for one gesture.
type layoutState struct {
	SidebarWidth float32 `json:"sidebarWidth,omitempty"`
	OutlineWidth float32 `json:"outlineWidth,omitempty"`
	// Library and Outline are the panes' own visibility. They are pointers so
	// that a layout saved before these existed reads as "not stated" rather than
	// as "closed", and so takes the default of open.
	Library *bool `json:"library,omitempty"`
	Outline *bool `json:"outline,omitempty"`
}

const layoutFile = "layout.json"

var (
	layoutMu     sync.Mutex
	layoutLoaded bool
	layoutCache  layoutState
)

// layoutDir is where the layout is kept: the state directory when one is set
// (a test), else the app's user data directory.
func layoutDir() string {
	stateMu.Lock()
	dir := stateDir
	stateMu.Unlock()
	if dir != "" {
		return dir
	}
	appDir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return ""
	}
	return appDir
}

// loadLayout reads the saved layout once. A damaged file loses the layout, not
// the app.
func loadLayout() layoutState {
	layoutMu.Lock()
	defer layoutMu.Unlock()
	if layoutLoaded {
		return layoutCache
	}
	layoutLoaded = true
	dir := layoutDir()
	if dir == "" {
		return layoutCache
	}
	data, err := os.ReadFile(filepath.Join(dir, layoutFile))
	if err != nil {
		return layoutCache
	}
	_ = json.Unmarshal(data, &layoutCache)
	return layoutCache
}

// saveLayout writes the layout. A failure is reported rather than swallowed:
// the reader set that width on purpose, and losing it silently is the bug this
// exists to prevent.
func saveLayout(l layoutState) error {
	layoutMu.Lock()
	layoutCache, layoutLoaded = l, true
	layoutMu.Unlock()

	dir := layoutDir()
	if dir == "" {
		return nil
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, layoutFile)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// SaveLayout writes the window's layout, as the app closes.
func (r *Reader) SaveLayout() error {
	library, outline := r.showLibrary, r.showOutline
	return saveLayout(layoutState{
		SidebarWidth: clampSidebar(r.sidebarW),
		OutlineWidth: clampOutline(r.outlineW),
		Library:      &library,
		Outline:      &outline,
	})
}
