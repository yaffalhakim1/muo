package pdf

import (
	"strings"
	"testing"
)

// The text of a page is what the page draws, run by run, with each run's place
// on it. Find needs both: the words to search for, and where to scroll.
func TestPageText(t *testing.T) {
	content := "BT /F1 12 Tf 72 720 Td (Hello, reader) Tj ET"
	path := writeTemp(t, "text.pdf", simplePDF(content))
	doc, ok := Load(path)
	if !ok {
		t.Fatal("the file did not load")
	}
	runs, err := doc.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) == 0 {
		t.Fatal("no text runs were recorded")
	}
	var all strings.Builder
	for _, r := range runs {
		all.WriteString(r.Text)
	}
	if !strings.Contains(all.String(), "Hello, reader") {
		t.Fatalf("the text is %q, want it to hold \"Hello, reader\"", all.String())
	}
	// The run sits where the page put it: 72 points in from the left, and 72
	// points up from a 792-point page, which is 72 from the top.
	r := runs[0]
	if r.X < 71 || r.X > 73 {
		t.Errorf("X = %.1f, want about 72", r.X)
	}
	if r.Y < 71 || r.Y > 73 {
		t.Errorf("Y = %.1f, want about 72 (the page is 792 high)", r.Y)
	}
	if r.Size < 11 || r.Size > 13 {
		t.Errorf("size = %.1f, want about 12", r.Size)
	}
}

// A run drawn lower on the page has a larger Y, because the app's Y grows
// downward and the PDF's grows up.
func TestPageTextYIsFlipped(t *testing.T) {
	content := "BT /F1 12 Tf 72 700 Td (high) Tj ET\n" +
		"BT /F1 12 Tf 72 100 Td (low) Tj ET"
	path := writeTemp(t, "flip.pdf", simplePDF(content))
	doc, _ := Load(path)
	runs, err := doc.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) < 2 {
		t.Fatalf("%d runs, want 2", len(runs))
	}
	high, low := runs[0], runs[1]
	if !strings.Contains(high.Text, "high") || !strings.Contains(low.Text, "low") {
		t.Fatalf("runs are %q and %q", high.Text, low.Text)
	}
	if high.Y >= low.Y {
		t.Errorf("\"high\" is at Y=%.0f and \"low\" at Y=%.0f; the lower text should have the larger Y",
			high.Y, low.Y)
	}
}

// Text joins the runs of a page into lines, one per baseline.
func TestTextJoinsLines(t *testing.T) {
	content := "BT /F1 12 Tf 72 720 Td (first line) Tj ET\n" +
		"BT /F1 12 Tf 72 700 Td (second line) Tj ET"
	path := writeTemp(t, "lines.pdf", simplePDF(content))
	doc, _ := Load(path)
	pages := doc.Text()
	if len(pages) != 1 {
		t.Fatalf("%d pages, want 1", len(pages))
	}
	lines := strings.Split(strings.TrimSpace(pages[0]), "\n")
	if len(lines) != 2 {
		t.Fatalf("%d lines, want 2: %q", len(lines), pages[0])
	}
	if !strings.Contains(lines[0], "first line") {
		t.Errorf("first line = %q", lines[0])
	}
	if !strings.Contains(lines[1], "second line") {
		t.Errorf("second line = %q", lines[1])
	}
}

// Asking for a page the file does not have is an error, not a panic.
func TestPageTextOutsideTheFile(t *testing.T) {
	path := writeTemp(t, "one.pdf", simplePDF("BT /F1 12 Tf 72 720 Td (x) Tj ET"))
	doc, _ := Load(path)
	if _, err := doc.PageText(7); err == nil {
		t.Fatal("a page outside the file was not an error")
	}
	if _, err := doc.PageText(-1); err == nil {
		t.Fatal("a negative page was not an error")
	}
}

// A page with no text at all yields no runs, and no error.
func TestPageTextOfABlankPage(t *testing.T) {
	path := writeTemp(t, "blank.pdf", simplePDF("0 0 m 100 100 l S"))
	doc, _ := Load(path)
	runs, err := doc.PageText(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 0 {
		t.Fatalf("%d runs on a page with no text", len(runs))
	}
}
