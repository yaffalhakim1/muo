package muo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/egoist/mygo"
)

// When each document was last opened.
//
// The recent list in the workspace and on the empty page shows how long ago a
// file was last read, and that is a fact the app has to keep: the operating
// system's recent-documents list carries the paths and nothing else, by design,
// so it cannot answer the question. This is a small side table keyed by path, in
// the same directory as the reading positions, written the same way.
//
// It is deliberately not a replacement for mygo's list. That list is what the
// Dock, the jump list and the Open Recent menu show, and it is the platform's;
// this is only the app's own display detail.

const recentTimesFile = "recent-times.json"

var (
	recentMu    sync.Mutex
	recentTimes map[string]time.Time
	recentNow   = time.Now
)

// loadRecentTimes reads the table once.
func loadRecentTimes() map[string]time.Time {
	recentMu.Lock()
	defer recentMu.Unlock()
	if recentTimes != nil {
		return recentTimes
	}
	recentTimes = map[string]time.Time{}
	dir := stateDirOf()
	if dir == "" {
		return recentTimes
	}
	data, err := os.ReadFile(filepath.Join(dir, recentTimesFile))
	if err != nil {
		return recentTimes
	}
	// Stored as Unix seconds rather than as a formatted time, so the file does
	// not depend on a layout and survives a timezone change.
	var raw map[string]int64
	if err := json.Unmarshal(data, &raw); err != nil {
		// A damaged file loses the times, not the app.
		return recentTimes
	}
	for path, sec := range raw {
		recentTimes[path] = time.Unix(sec, 0)
	}
	return recentTimes
}

// touchRecent records that path was opened now.
func touchRecent(path string) {
	times := loadRecentTimes()
	recentMu.Lock()
	times[path] = recentNow()
	recentMu.Unlock()
}

// recentOpenedAt returns when path was last opened here, if it was.
func recentOpenedAt(path string) (time.Time, bool) {
	times := loadRecentTimes()
	recentMu.Lock()
	defer recentMu.Unlock()
	t, ok := times[path]
	return t, ok
}

// saveRecentTimes writes the table. It is called with the rest of the app's
// state as the window closes.
func saveRecentTimes() error {
	times := loadRecentTimes()
	recentMu.Lock()
	raw := make(map[string]int64, len(times))
	for path, t := range times {
		raw[path] = t.Unix()
	}
	recentMu.Unlock()

	dir := stateDirOf()
	if dir == "" {
		return nil
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, recentTimesFile)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// stateDirOf is the directory the app's own state lives in: the one a test set,
// else the app's user data directory.
func stateDirOf() string {
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

// agoLabel says how long ago a time was, in the form a recent list uses: a count
// of minutes, then of hours, then "Yesterday", then a date.
//
// It is coarse on purpose. A recent list is scanned, not read, and "3m ago"
// answers "is this the one I was just in" better than a timestamp does.
func agoLabel(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := recentNow().Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + "m ago"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d.Hours())) + "h ago"
	case d < 48*time.Hour:
		return "Yesterday"
	case d < 7*24*time.Hour:
		return strconv.Itoa(int(d.Hours()/24)) + "d ago"
	}
	return t.Format("Jan 2")
}
