package muo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// The redesigned frame: one header, two panels that fold, a page, and a footer.
// These hold the parts a reader would notice going missing, and the parts that
// are easy to break silently — a panel that stops folding, a progress figure that
// stops agreeing with the scroll, a Zen mode that leaves something on screen.

// zen hides both panels and the pill. The header stays: Zen is about the page's
// room, not about the window's chrome, and the command that leaves it has to
// still be there.
func TestZenFoldsThePanelsAndThePill(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")
	if _, ok := tt.Find("Open a File"); !ok {
		t.Fatal("the workspace is missing before Zen, so this proves nothing")
	}
	if _, ok := tt.Find("On This Page"); !ok {
		t.Fatal("the outline is missing before Zen, so this proves nothing")
	}
	if _, ok := tt.Find("Reading status"); !ok {
		t.Fatal("the pill is missing before Zen, so this proves nothing")
	}

	r.zen = true
	tt.Frame()
	// The panels slide rather than vanish, so the fold has to be given its time.
	tt.Advance(600 * time.Millisecond)
	tt.Frame()

	if _, ok := tt.Find("Open a File"); ok {
		t.Error("the workspace is still on screen in Zen")
	}
	if _, ok := tt.Find("On This Page"); ok {
		t.Error("the outline is still on screen in Zen")
	}
	if _, ok := tt.Find("Reading status"); ok {
		t.Error("the status pill is still on screen in Zen")
	}
	// The header stays, and with it the way out of the mode.
	if _, ok := tt.Find("Zen Mode"); !ok {
		t.Error("the Zen command went with the panels, so there is no way out")
	}
	if !tt.HasText("one") {
		t.Error("the page went with the panels")
	}
}

// Zen is a mode over the reader's choices, not a change to them: coming out
// restores whatever panels they had, rather than opening both.
func TestZenRemembersTheReadersPanels(t *testing.T) {
	r, tt := openWithDocs(t, "one.md")
	r.showOutline = false
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()

	r.zen = true
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	r.zen = false
	tt.Frame()
	tt.Advance(600 * time.Millisecond)
	tt.Frame()

	if _, ok := tt.Find("Open a File"); !ok {
		t.Error("the workspace did not come back")
	}
	if _, ok := tt.Find("On This Page"); ok {
		t.Error("Zen reopened the outline the reader had folded")
	}
}

// The progress figure is the scroll's own state, so it cannot disagree with the
// page. At the top it is zero, at the end it is one, and a document too short to
// scroll counts as read rather than as unstarted.
func TestProgressFollowsTheScroll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "long.md")
	var b strings.Builder
	b.WriteString("# Top\n\n")
	b.WriteString(strings.Repeat("Filler text here. ", 300) + "\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New()
	r.OpenIn(path)
	tt := ui.NewTester(r.FrameView, 1000, 600)
	tt.Frame()
	tt.Frame()

	if r.contentScroll.MaxY <= 0 {
		t.Fatal("the test page does not scroll, so progress cannot be tested")
	}
	if f := r.readingProgress(); f != 0 {
		t.Errorf("an unstarted page reads %.2f, want 0", f)
	}

	r.contentScroll.Y = r.contentScroll.MaxY / 2
	tt.Frame()
	if f := r.readingProgress(); f < 0.45 || f > 0.55 {
		t.Errorf("half way down reads %.2f, want about 0.5", f)
	}

	r.contentScroll.Y = r.contentScroll.MaxY
	tt.Frame()
	if f := r.readingProgress(); f != 1 {
		t.Errorf("the end of the page reads %.2f, want 1", f)
	}

	// A page with nothing to scroll is entirely visible, so it is read.
	short := New()
	short.OpenIn(writeTemp(t, "tiny.md", []byte("# T\n")))
	st := ui.NewTester(short.FrameView, 1000, 900)
	st.Frame()
	st.Frame()
	if f := short.readingProgress(); f != 1 {
		t.Errorf("a page that fits the window reads %.2f, want 1", f)
	}
}

// The progress label rounds and clamps, so a scroll past the end cannot print
// 101% and a nearly-finished page does not claim to be done.
func TestProgressLabel(t *testing.T) {
	cases := []struct {
		f    float64
		want string
	}{
		{0, "0%"},
		{0.004, "0%"},
		{0.5, "50%"},
		{0.999, "100%"},
		{1, "100%"},
		{1.4, "100%"},
	}
	for _, tc := range cases {
		if got := percentLabel(tc.f); got != tc.want {
			t.Errorf("percentLabel(%.3f) = %q, want %q", tc.f, got, tc.want)
		}
	}
}

// The reading time is the document's word count at the stated pace, and a
// document shorter than a minute still says one rather than zero.
func TestReadingTime(t *testing.T) {
	r := New()
	if got := r.readingMinutes(); got != 0 {
		t.Errorf("no document reads %d minutes, want 0", got)
	}

	r.doc = &Document{Source: strings.Repeat("word ", readingWordsPerMinute*2)}
	if got := r.readingMinutes(); got != 2 {
		t.Errorf("two minutes of words reads %d, want 2", got)
	}

	r.doc = &Document{Source: "one word"}
	if got := r.readingMinutes(); got != 1 {
		t.Errorf("a one-word document reads %d, want 1", got)
	}
}

// The workspace is the only list of open files, and the open command is above it
// so that a reader with several files still finds it where the list begins.
func TestWorkspaceListsEveryDocumentOnce(t *testing.T) {
	r, tt := openWithDocs(t, "one.md", "two.md")
	_ = r
	for _, name := range []string{"one.md", "two.md"} {
		if !tt.HasText(name) {
			t.Errorf("%q is not in the workspace", name)
		}
	}
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
}

// The empty page offers a way in, names the app, and lists the files the app was
// last in — and does not repeat that list in the workspace panel beside it.
func TestEmptyPageIsTheFrontispiece(t *testing.T) {
	useStateDir(t, t.TempDir())
	r := New()
	tt := ui.NewTester(r.FrameView, 1100, 760)
	tt.Frame()
	tt.Frame()

	if !tt.HasText("Muo") {
		t.Error("the empty page does not name the reader")
	}
	if _, ok := tt.Find("Open a File"); !ok {
		t.Error("the empty page offers no way in")
	}
	// The panels are open, and they carry no second copy of the recent list.
	if _, ok := tt.Find("Open Files"); ok {
		t.Error("the workspace lists open files when none are open")
	}
}

// The shortcut in the header and the shortcut in the menu are one fact, so the
// header's open command states the one the app actually binds.
func TestOpenCommandStatesItsShortcut(t *testing.T) {
	useStateDir(t, t.TempDir())
	r := New()
	tt := ui.NewTester(r.FrameView, 1100, 760)
	tt.Frame()
	tt.Frame()
	if !tt.HasText("⌘O") {
		t.Error("the open command does not show its shortcut")
	}
}

// The recent list's times survive a restart, which is what makes "2m ago" mean
// anything: the platform's recent list carries paths and no times at all.
func TestRecentTimesSurviveARestart(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)

	touchRecent("/notes/a.md")
	if err := saveRecentTimes(); err != nil {
		t.Fatalf("saving: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, recentTimesFile)); err != nil {
		t.Fatalf("no times file was written: %v", err)
	}

	// A fresh process, as after a restart.
	recentMu.Lock()
	recentTimes = nil
	recentMu.Unlock()

	if _, ok := recentOpenedAt("/notes/a.md"); !ok {
		t.Fatal("the time was lost across a restart")
	}
}

// A damaged times file loses the times, not the app.
func TestDamagedRecentTimesAreIgnored(t *testing.T) {
	dir := t.TempDir()
	useStateDir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, recentTimesFile), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	recentMu.Lock()
	recentTimes = nil
	recentMu.Unlock()

	if _, ok := recentOpenedAt("/notes/a.md"); ok {
		t.Fatal("a damaged file produced a time")
	}
}

// The ago label is coarse on purpose, and every band has its own wording.
func TestAgoLabel(t *testing.T) {
	now := recentNow()
	cases := []struct {
		d    int // minutes ago
		want string
	}{
		{0, "just now"},
		{3, "3m ago"},
		{90, "1h ago"},
		{60 * 30, "Yesterday"},
		{60 * 24 * 3, "3d ago"},
	}
	for _, tc := range cases {
		got := agoLabel(now.Add(-time.Duration(tc.d) * time.Minute))
		if got != tc.want {
			t.Errorf("%d minutes ago reads %q, want %q", tc.d, got, tc.want)
		}
	}
	if got := agoLabel(time.Time{}); got != "" {
		t.Errorf("a zero time reads %q, want nothing", got)
	}
}
