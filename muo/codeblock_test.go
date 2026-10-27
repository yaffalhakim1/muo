package muo

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Format rewrites the block in the document, so what is shown and what Copy
// copies are the formatted code.
func TestFormatBlockRewritesTheSource(t *testing.T) {
	src := "# T\n\n```go\nfunc main(){x:=1\n}\n```\n"
	r := New()
	r.doc = &Document{Format: FormatMarkdown, Source: src, Title: "t.md"}
	tt := ui.NewTester(r.FrameView, 900, 600)
	tt.Frame()

	if err := tt.Click("Format"); err != nil {
		t.Fatalf("no Format button: %v", err)
	}
	tt.Frame()

	if !strings.Contains(r.doc.Source, "func main() {") {
		t.Fatalf("the block was not formatted:\n%s", r.doc.Source)
	}
	if !strings.HasPrefix(r.doc.Source, "# T") {
		t.Fatal("the rest of the document was lost")
	}
}

// A fence with no language has nothing for the app to tidy, so the block must
// not offer the command at all. It used to draw a Format button that could only
// ever be disabled — a control on the page that does nothing, on roughly half
// the code blocks in a real document.
func TestUntaggedFenceOffersNoFormatCommand(t *testing.T) {
	src := "# T\n\n```\nno language here\n```\n"
	r := New()
	r.doc = &Document{Format: FormatMarkdown, Source: src, Title: "t.md"}
	tt := ui.NewTester(r.FrameView, 900, 600)
	tt.Frame()

	if err := tt.Click("Format"); err == nil {
		t.Fatal("an untagged fence drew a Format button")
	}
}
