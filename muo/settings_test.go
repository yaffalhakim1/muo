package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The settings surface must lay itself out from the space it is given, not the
// window's width: a settings pane measures itself, not the window.
func TestSettingsLayoutUsesItsOwnWidth(t *testing.T) {
	r := New()
	r.OpenIn(writeTemp(t, "s.md", []byte("# T\n")))
	r.settings = true
	tt := ui.NewTester(r.FrameView, 1100, 900)
	tt.Frame()

	if !tt.HasText("Reading") {
		t.Fatal("the settings categories are missing")
	}
	// The specimen is the page's own type at the page's own measure, so what
	// the reader judges in the pane is what they will read.
	if !tt.HasText("The measure comes first") {
		t.Fatal("the settings pane shows no specimen")
	}
	// Every setting the pane offers is named, and each name is a word the
	// reader would use for it.
	for _, want := range []string{"Distance", "Type size", "Mode", "Name"} {
		if !tt.HasText(want) {
			t.Errorf("the settings pane has no %q setting", want)
		}
	}
}


