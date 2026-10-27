package muo

import (
	"testing"
)

// The measure is a promise in characters, not in pixels. The column is solved in
// whatever face the page is drawn in, and the two bundled faces are not the same
// width: if the column were solved in one and drawn in the other, the promise
// would be broken by however much they differ.
func TestTheMeasureIsSolvedInThePageFamily(t *testing.T) {
	if err := PageFontError(); err != nil {
		t.Fatalf("registering the page fonts: %v", err)
	}
	d := Display{Distance: 65}
	system := Model(d, 1600, 109, "", 0)
	page := Model(d, 1600, 109, pageFamily(), 0)

	// The page's face holds the measure...
	if page.ChPerLine != charsPerLine {
		t.Fatalf("the page column holds %.1f characters, want %.0f", page.ChPerLine, charsPerLine)
	}
	// ...and it holds it in a column of a different width from the system's own
	// font, which is the proof that the column was solved for the face it is
	// drawn in rather than for a default.
	if system.Ch == page.Ch {
		t.Fatalf("both faces measure %.2f per character: the family was ignored", system.Ch)
	}
	if system.ColumnW == page.ColumnW {
		t.Fatalf("both faces give a column of %.0f: the family was ignored", page.ColumnW)
	}
}

// Both bundled faces install, and each answers to its own family name.
func TestBothPageFacesInstall(t *testing.T) {
	if err := PageFontError(); err != nil {
		t.Fatalf("registering the page fonts: %v", err)
	}
	d := Display{Distance: 65}
	for _, family := range []string{familyInter, familyNewsreader} {
		typ := Model(d, 1600, 109, family, 0)
		if typ.Ch <= 0 {
			t.Errorf("%s did not measure a character", family)
		}
		if typ.ChPerLine != charsPerLine {
			t.Errorf("%s holds %.1f characters, want %.0f", family, typ.ChPerLine, charsPerLine)
		}
	}
	inter := Model(d, 1600, 109, familyInter, 0)
	news := Model(d, 1600, 109, familyNewsreader, 0)
	if inter.Ch == news.Ch {
		t.Errorf("Inter and Newsreader measure the same character width (%.3f): "+
			"one of them did not install", inter.Ch)
	}
	t.Logf("Inter ch=%.3f column=%.1f  Newsreader ch=%.3f column=%.1f",
		inter.Ch, inter.ColumnW, news.Ch, news.ColumnW)
}
