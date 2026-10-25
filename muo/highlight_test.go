package muo

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
)

// Every token kind the lexer can emit must have a color of its own on the code
// block's surface; a zero color falls back to the block's text color and the
// highlighting silently disappears for that kind.
func TestCodeThemeColorsEveryTokenKind(t *testing.T) {
	var ct *editor.Theme
	tt := ui.NewTester(func(c *ui.Context) {
		pal := pageColors(c)
		ct = codeTheme(c.Theme(), pal, pal.paper)
		ui.Text(c, "x")
	}, 400, 300)
	tt.Frame()

	kinds := []editor.TokenKind{
		editor.TokenKeyword, editor.TokenType, editor.TokenFunction,
		editor.TokenVariable, editor.TokenProperty, editor.TokenConstant,
		editor.TokenNumber, editor.TokenString, editor.TokenEscape,
		editor.TokenRegexp, editor.TokenComment, editor.TokenOperator,
		editor.TokenPunctuation, editor.TokenTag, editor.TokenAttribute,
		editor.TokenMeta, editor.TokenHeading, editor.TokenEmphasis,
		editor.TokenStrong, editor.TokenLink, editor.TokenCode,
		editor.TokenInserted, editor.TokenDeleted, editor.TokenInvalid,
	}
	missing := []string{}
	for _, k := range kinds {
		if ct.Tokens[k].Color.A == 0 {
			missing = append(missing, tokenName(k))
		}
	}
	if len(missing) > 0 {
		t.Fatalf("token kinds with no color: %v", missing)
	}
}

func tokenName(k editor.TokenKind) string {
	names := map[editor.TokenKind]string{
		editor.TokenKeyword: "Keyword", editor.TokenType: "Type",
		editor.TokenFunction: "Function", editor.TokenVariable: "Variable",
		editor.TokenProperty: "Property", editor.TokenConstant: "Constant",
		editor.TokenNumber: "Number", editor.TokenString: "String",
		editor.TokenEscape: "Escape", editor.TokenRegexp: "Regexp",
		editor.TokenComment: "Comment", editor.TokenOperator: "Operator",
		editor.TokenPunctuation: "Punctuation", editor.TokenTag: "Tag",
		editor.TokenAttribute: "Attribute", editor.TokenMeta: "Meta",
		editor.TokenHeading: "Heading", editor.TokenEmphasis: "Emphasis",
		editor.TokenStrong: "Strong", editor.TokenLink: "Link",
		editor.TokenCode: "Code", editor.TokenInserted: "Inserted",
		editor.TokenDeleted: "Deleted", editor.TokenInvalid: "Invalid",
	}
	if n, ok := names[k]; ok {
		return n
	}
	return "Token"
}
