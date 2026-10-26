package muo

import (
	"testing"
)

// The measure is a promise in characters, not in pixels. The page is set in the
// bundled serif, and a serif is not the width of the system's sans: if the
// column were solved in one face and drawn in the other, the promise would be
// broken by however much the two differ — about two percent, which is a
// character and a half a line.
func TestTheMeasureIsSolvedInThePageFamily(t *testing.T) {
	if err := PageFontError(); err != nil {
		t.Fatalf("registering the page font: %v", err)
	}
	d := Display{Distance: 65}
	sans := Model(d, 1600, 109, "", 0)
	serif := Model(d, 1600, 109, pageFamily, 0)

	// Both hold the measure...
	if serif.ChPerLine != charsPerLine {
		t.Fatalf("the serif column holds %.1f characters, want %.0f", serif.ChPerLine, charsPerLine)
	}
	// ...and they hold it in columns of different widths, which is the proof
	// that the column was solved for the face it is drawn in.
	if sans.Ch == serif.Ch {
		t.Fatalf("both faces measure %.2f per character: the family was ignored", sans.Ch)
	}
	if sans.ColumnW == serif.ColumnW {
		t.Fatalf("both faces give a column of %.0f: the family was ignored", serif.ColumnW)
	}
}
