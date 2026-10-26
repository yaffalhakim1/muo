package muo

import (
	"math"
	"testing"
)

// The typography model is the reader's whole idea, so it is tested as a
// contract: the same comfort across two very different screens, and a measure
// that stays in characters.

func TestModelHoldsTheAngleAcrossScreens(t *testing.T) {
	// A 27" monitor at 65 cm and a 16" laptop at 42 cm, at the densities those
	// panels actually report.
	cases := []struct {
		name         string
		distance     float64
		ppi          float64
		wantSizeBand [2]float32
	}{
		{"27-inch at 65 cm, 1x", 65, 109, [2]float32{20, 25}},
		{"16-inch at 42 cm, 2x", 42, 224, [2]float32{27, 32}},
		{"24-inch at 55 cm, 1x", 55, 92, [2]float32{14, 18}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := Display{Distance: tc.distance}
			got := Model(d, 1600, tc.ppi, "", 0)
			if got.Comfort != ComfortGood {
				t.Fatalf("comfort = %v at %.2f°, want good", got.Comfort, got.Angle)
			}
			if math.Abs(got.Angle-targetAngle) > 0.002 {
				t.Fatalf("angle = %.4f°, want %.2f°", got.Angle, targetAngle)
			}
			if got.FontPx < tc.wantSizeBand[0] || got.FontPx > tc.wantSizeBand[1] {
				t.Fatalf("size = %.1f, want within %v", got.FontPx, tc.wantSizeBand)
			}
			if got.ChPerLine < 71 || got.ChPerLine > 73 {
				t.Fatalf("measure = %.1f characters, want 72", got.ChPerLine)
			}
		})
	}
}

func TestModelAsksForDifferentPixelsOnDifferentScreens(t *testing.T) {
	// The same apparent size on a laptop and a desktop needs different pixels.
	laptop := Model(Display{Distance: 42}, 1600, 224, "", 0)
	desktop := Model(Display{Distance: 65}, 1600, 109, "", 0)
	if laptop.FontPx <= desktop.FontPx {
		t.Fatalf("laptop %.1f should need more points than desktop %.1f for the same angle: %.3f vs %.3f",
			laptop.FontPx, desktop.FontPx, laptop.Angle, desktop.Angle)
	}
}

func TestColumnStaysInCharactersWhenTheWindowIsNarrow(t *testing.T) {
	// A narrow window cannot show 72 characters at a large size: the column
	// gives way, and what it shows is fewer characters, never a different
	// character count on the big screen.
	narrow := Model(Display{Distance: 65}, 520, 109, "", 0)
	if narrow.ChPerLine >= 72 {
		t.Fatalf("a 520 pt window reported %.1f characters", narrow.ChPerLine)
	}
	if narrow.ColumnW > 520 {
		t.Fatalf("column %.1f overflows the window", narrow.ColumnW)
	}
}

func TestManualSizeOverridesTheModel(t *testing.T) {
	d := Display{Distance: 65}
	auto := Model(d, 1600, 109, "", 0)
	manual := Model(d, 1600, 109, "", 28)
	if manual.FontPx != 28 {
		t.Fatalf("manual size = %.1f, want 28", manual.FontPx)
	}
	if math.Abs(manual.Angle-auto.Angle) < 0.01 {
		t.Fatal("the manual size should have changed the apparent angle")
	}
	if manual.PreferredSize != auto.FontPx {
		t.Fatalf("preferred size %.1f should still be the model's %.1f", manual.PreferredSize, auto.FontPx)
	}
}

func TestComfortGradeFollowsTheBand(t *testing.T) {
	cases := []struct {
		px    float32
		ppi   float64
		want  Comfort
		about string
	}{
		{8, 109, ComfortSmall, "8 pt on a desktop screen is too small"},
		{14, 72, ComfortGood, "14 pt at 72 ppi and 65 cm is in the band"},
		{60, 109, ComfortLarge, "60 pt on a desktop screen is too large"},
	}
	for _, tc := range cases {
		got := grade(ComfortAngle(tc.px, tc.ppi, 65))
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.about, got, tc.want)
		}
	}
}

func TestLineHeightTightensAsLinesShorten(t *testing.T) {
	if readingLineHeight(15) >= readingLineHeight(22) {
		t.Fatal("small type should get the tighter leading")
	}
	if h := readingLineHeight(22); h < 1.7 || h > 1.9 {
		t.Fatalf("leading %.2f is outside the measured band", h)
	}
}

func TestReadingSurfaceAvoidsMaximumContrast(t *testing.T) {
	bg, fg, muted, _, _ := readingSurface(false)
	if bg == (rgbColor{255, 255, 255}) || fg == (rgbColor{0, 0, 0}) {
		t.Fatal("a pure white page or pure black ink is maximum contrast")
	}
	if contrastRatio(bg, fg) < 7 || contrastRatio(bg, fg) > 16 {
		t.Fatalf("light page contrast %.1f is outside the 7–16 band reading wants", contrastRatio(bg, fg))
	}
	if contrastRatio(bg, muted) < 4.5 {
		t.Fatalf("muted text contrast %.1f falls below the 4.5 accessibility floor", contrastRatio(bg, muted))
	}

	dbg, dfg, _, _, _ := readingSurface(true)
	if dbg == (rgbColor{0, 0, 0}) || dfg == (rgbColor{255, 255, 255}) {
		t.Fatal("a pure black page with pure white ink is maximum contrast")
	}
	if c := contrastRatio(dbg, dfg); c < 6 || c > 14 {
		t.Fatalf("dark page contrast %.1f is outside the 6–14 band", c)
	}
}

func TestComfortSizeInvertsTheAngle(t *testing.T) {
	for _, distance := range []float64{35, 42, 55, 65, 90} {
		px := comfortSize(96, distance)
		if angle := ComfortAngle(px, 96, distance); math.Abs(angle-targetAngle) > 0.001 {
			t.Fatalf("%.0f cm: %.2f pt gives %.3f°, want %.2f°", distance, px, angle, targetAngle)
		}
	}
}

// contrastRatio is the WCAG ratio of two colors, used to keep the page inside
// the band the reading research describes.
func contrastRatio(a, b rgbColor) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func luminance(c rgbColor) float64 {
	channel := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
}





