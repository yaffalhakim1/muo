package muo

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Snapshots for design work: MUO_SHOTS=1 go test ./muo -run TestShots renders
// each state of the window to shots/, so a change can be looked at rather than
// imagined.
func TestShots(t *testing.T) {
	if os.Getenv("MUO_SHOTS") == "" {
		t.Skip("set MUO_SHOTS=1 to render the window to shots/")
	}
	dir := filepath.Join("..", "shots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	sample, err := filepath.Abs(filepath.Join("..", "sample.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The kitchen-sink document, so the shots cover every block rather than the
	// handful the sample happens to use.
	all := filepath.Join(dir, "all-blocks.md")
	if err := os.WriteFile(all, []byte(everyBlock), 0o644); err != nil {
		t.Fatal(err)
	}

	type shot struct {
		name   string
		setup  func(r *Reader)
		dark   bool
		width  int
		height int
	}
	shots := []shot{
		{name: "empty-light", width: 1100, height: 760},
		{name: "empty-dark", width: 1100, height: 760, dark: true},
		{name: "doc-light", width: 1100, height: 760, setup: func(r *Reader) { r.OpenIn(sample) }},
		{name: "doc-wide-light", width: 1440, height: 860, setup: func(r *Reader) { r.OpenIn(sample) }},
		{name: "doc-wide-dark", width: 1440, height: 860, dark: true, setup: func(r *Reader) { r.OpenIn(sample) }},
		{name: "doc-dark", width: 1100, height: 760, dark: true, setup: func(r *Reader) { r.OpenIn(sample) }},
		{name: "blocks-light", width: 1100, height: 1500, setup: func(r *Reader) { r.OpenIn(all) }},
		{name: "blocks-dark", width: 1100, height: 1500, dark: true, setup: func(r *Reader) { r.OpenIn(all) }},
		{name: "settings-light", width: 1100, height: 760, setup: func(r *Reader) {
			r.OpenIn(sample)
			r.settings = true
		}},
		{name: "settings-dark", width: 1100, height: 760, dark: true, setup: func(r *Reader) {
			r.OpenIn(sample)
			r.settings = true
		}},
		{name: "find-light", width: 1100, height: 760, setup: func(r *Reader) {
			r.OpenIn(sample)
			r.finding = true
			r.query = "measure"
		}},
		{name: "narrow-light", width: 720, height: 620, setup: func(r *Reader) { r.OpenIn(sample) }},
	}

	for _, s := range shots {
		r := New()
		r.SetSidebarWidth(SidebarWidth)
		if s.setup != nil {
			s.setup(r)
		}
		tt := ui.NewTester(r.FrameView, s.width, s.height)
		tt.SetDark(s.dark)
		tt.Frame()
		tt.Frame()
		f, err := os.Create(filepath.Join(dir, s.name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}
