package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

// Highlighting only reads as highlighting if each kind differs from plain text.
// A theme whose Type color equals the block's text color paints types as if the
// lexer had not marked them at all.
func TestTokenColorsAreDistinct(t *testing.T) {
	var ct *editor.Theme
	var ink ui.Color
	tt := ui.NewTester(func(c *ui.Context) {
		pal := pageColors(c)
		ct = codeTheme(c.Theme(), pal, pal.paper)
		ink = pal.inkColor()
		ui.Text(c, "x")
	}, 400, 300)
	tt.Frame()

	for _, tc := range []struct {
		name string
		kind editor.TokenKind
	}{
		{"Keyword", editor.TokenKeyword},
		{"Type", editor.TokenType},
		{"Function", editor.TokenFunction},
		{"String", editor.TokenString},
		{"Number", editor.TokenNumber},
		{"Comment", editor.TokenComment},
	} {
		got := ct.Tokens[tc.kind].Color
		if got == ink {
			t.Errorf("%s shares the body ink #%02x%02x%02x: it will not read as highlighted", tc.name, got.R, got.G, got.B)
			continue
		}
		t.Logf("%-9s #%02x%02x%02x (ink #%02x%02x%02x)", tc.name, got.R, got.G, got.B, ink.R, ink.G, ink.B)
	}

	// Type and Variable must differ, or a type name and a parameter look alike.
	typ := ct.Tokens[editor.TokenType].Color
	vr := ct.Tokens[editor.TokenVariable].Color
	if typ == vr {
		t.Errorf("type and variable share #%02x%02x%02x", typ.R, typ.G, typ.B)
	}
}
