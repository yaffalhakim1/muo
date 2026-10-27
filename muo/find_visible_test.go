package muo

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// A query with one match, whose block is already off screen, must still move
// the page: there is no second match to move to, so `match` stays 0 and the
// bar's own choice never changes. Typing "what is left out" in a document whose
// heading of that name is below the fold used to count 1 of 1 and go nowhere.
func TestFindMovesToOneMatch(t *testing.T) {
	body := strings.Repeat("filler prose that is not the needle.\n\n", 40)
	src := "# Top\n\n" + body + "## What is left out\n\nend.\n"
	r := openString(t, "one.md", src)
	tt := ui.NewTester(r.FrameView, 700, 400)
	tt.Frame()
	tt.Frame()

	if r.contentScroll.MaxY <= 0 {
		t.Fatalf("the page does not scroll (maxY=%.0f)", r.contentScroll.MaxY)
	}

	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("what is left out")
	for i := 0; i < 6; i++ {
		tt.Frame()
	}

	if got := r.paintableMatches(); got != 1 {
		t.Fatalf("the bar counts %d paintable matches, want 1", got)
	}
	if r.contentScroll.Y <= 0 {
		t.Fatalf("the page did not move to its only match (Y=%.0f)", r.contentScroll.Y)
	}
}

// A match inside a code fence is counted by a plain source search but can never
// be painted, because marking it would show the tags. The bar must not count it:
// "1 of 2" for a query with one visible match makes the reader walk to a match
// and find nothing lit.
func TestFindCountsOnlyVisibleMatches(t *testing.T) {
	src := "# The measure comes first\n\n" +
		"Prose mentions the measure here.\n\n" +
		"```go\n// the measure in a comment\nfunc f() {}\n```\n\n" +
		"- [ ] Read the measure section again\n"
	r := openString(t, "code.md", src)
	tt := ui.NewTester(r.FrameView, 900, 700)
	tt.Frame()
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("the measure")
	tt.Frame()
	tt.Frame()

	all := len(r.matches())
	if all != 4 {
		t.Fatalf("the source holds %d occurrences, want 4", all)
	}
	visible := r.paintableMatches()
	if visible != 3 {
		t.Errorf("the bar would count %d visible matches, want 3 (the one in the fence is not paintable)", visible)
	}
}

// Every match the bar counts must be a match it can light. Walking the whole
// set with Next must never land on one that the page cannot paint.
func TestFindNeverLandsOnAnUnpaintableMatch(t *testing.T) {
	src := "# The measure\n\n" +
		"the measure in prose.\n\n" +
		"```go\n// the measure in code\n```\n\n" +
		"## the measure again\n"
	r := openString(t, "walk.md", src)
	tt := ui.NewTester(r.FrameView, 900, 700)
	tt.Frame()
	tt.Frame()

	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Type("the measure")
	tt.Frame()
	tt.Frame()

	visible := r.paintableMatches()
	if visible != 3 {
		t.Fatalf("the bar counts %d visible matches, want 3", visible)
	}
	// Walking every match the bar offers must give a real source offset every
	// time — never a fence, which is what an untranslated index would hit.
	seen := map[int]bool{}
	for step := 0; step < visible; step++ {
		at, ok := r.matchAt(r.match)
		if !ok {
			t.Fatalf("step %d: the bar chose match %d of %d, which has no source offset",
				step, r.match+1, visible)
		}
		if seen[at] {
			t.Fatalf("step %d: match %d repeats offset %d", step, r.match+1, at)
		}
		seen[at] = true

		skip := unmarkable(r.doc.Source)
		if overlaps(skip, at, at+len(r.query)) {
			t.Fatalf("step %d: the bar chose a match at %d that the page cannot paint", step, at)
		}
		if err := tt.Click("Next"); err != nil {
			t.Fatalf("no Next button: %v", err)
		}
		tt.Frame()
		tt.Frame()
	}
	if len(seen) != visible {
		t.Fatalf("walking the bar reached %d distinct matches, want %d", len(seen), visible)
	}
}
