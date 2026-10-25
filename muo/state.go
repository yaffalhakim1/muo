package muo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/egoist/mygo"
)

// Where the reader was. Each document keeps the heading it stopped at and the
// pixel offset as a fallback, so closing the app and opening it again returns
// to the same place in the same file.
//
// The state lives in the app's user data directory (mygo.PathUserData), which
// the desktop creates on first use; nothing here is a secret, so it is plain
// JSON rather than the keychain.
type readingState struct {
	// Position is the scroll offset in DIPs, the fallback when the heading
	// cannot be found again (a document edited since).
	Position float32 `json:"position"`
	// Heading is the block index of the last heading visited, or -1.
	Heading int `json:"heading"`
	// Title remembers the heading's words, so a lookup survives an edit that
	// shifted every block index.
	Title string `json:"title,omitempty"`
}

const stateFile = "reading-state.json"

var (
	stateMu    sync.Mutex
	stateCache map[string]readingState
	stateDir   string
)

// loadState reads the saved positions once.
func loadState() map[string]readingState {
	stateMu.Lock()
	defer stateMu.Unlock()
	if stateCache != nil {
		return stateCache
	}
	stateCache = map[string]readingState{}
	dir := stateDir
	if dir == "" {
		appDir, err := mygo.App.Path(mygo.PathUserData)
		if err != nil {
			return stateCache
		}
		dir = appDir
		stateDir = dir
	}
	data, err := os.ReadFile(filepath.Join(dir, stateFile))
	if err != nil {
		return stateCache
	}
	// A damaged file loses the positions, not the app.
	_ = json.Unmarshal(data, &stateCache)
	return stateCache
}

// savedPosition returns where a document was left.
func savedPosition(path string) (readingState, bool) {
	loadState()
	stateMu.Lock()
	defer stateMu.Unlock()
	s, ok := stateCache[path]
	return s, ok
}

// rememberPosition records where a document is, in memory; saveState writes it
// to disk when the app closes.
func rememberPosition(path string, s readingState) {
	loadState()
	stateMu.Lock()
	defer stateMu.Unlock()
	stateCache[path] = s
}

// saveState writes the positions to the user data directory. A failure is
// reported rather than swallowed: losing the reader's place silently is the
// bug this exists to prevent.
func saveState() error {
	loadState()
	stateMu.Lock()
	defer stateMu.Unlock()
	if stateDir == "" || stateCache == nil {
		return nil
	}
	data, err := json.MarshalIndent(stateCache, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(stateDir, stateFile)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
