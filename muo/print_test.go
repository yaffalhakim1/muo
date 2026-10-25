package muo

import "testing"

func TestPrintModel(t *testing.T) {
	for _, tc := range []struct {
		name string
		d    float64
		ppi  float64
	}{
		{"27in @65cm @109ppi", 65, 109},
		{"16in @42cm @224ppi", 42, 224},
		{"16in @45cm @144ppi", 45, 144},
		{"24in @55cm @92ppi", 55, 92},
	} {
		m := Model(Display{Distance: tc.d}, 1600, tc.ppi, "", 0)
		t.Logf("%-22s size=%.1f angle=%.2f ch=%.1f line=%.2f", tc.name, m.FontPx, m.Angle, m.ChPerLine, m.Line)
	}
}
