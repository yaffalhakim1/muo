package muo

import (
	"image/png"
	"os"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Every state the window can show, rendered, so the whole app can be reviewed at
// once rather than one screenshot at a time.
func TestRenderStates(t *testing.T) {
	states := []struct {
		name  string
		setup func(r *Reader)
	}{
		{"reading", func(r *Reader) {
			r.OpenIn("Y:/Learn/Go/muo/sample.md")
		}},
		{"empty", func(r *Reader) {}},
		{"settings", func(r *Reader) {
			r.OpenIn("Y:/Learn/Go/muo/sample.md")
			r.settings = true
		}},
	}

	for _, st := range states {
		t.Run(st.name, func(t *testing.T) {
			r := New()
			r.SetSidebarWidth(sidebarWidth)
			st.setup(r)
			tt := ui.NewTester(r.FrameView, 1000, 700)
			tt.Frame()
			tt.Frame()
			f, err := os.Create("Y:/Learn/Go/muo/state-" + st.name + ".png")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			png.Encode(f, tt.Image())
		})
	}
}

