package code

import (
	"strings"
	"testing"
)

// Go goes through go/format, the same printer gofmt runs, so its output is
// exact — with one change on the way out: gofmt indents with tabs and this page
// cannot draw a tab, so the indentation is spaces by the time it is returned.
func TestFormatGoUsesTheStdlib(t *testing.T) {
	in := "package main\nfunc main(){x:=1\nif x>0{println(x)}}\n"
	out, changed := Format("go", in)
	if !changed {
		t.Fatal("unformatted Go was left alone")
	}
	want := "package main\n\nfunc main() {\n    x := 1\n    if x > 0 {\n        println(x)\n    }\n}\n"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

// Code that does not parse is still laid out rather than rejected: a snippet
// without its package clause is common in docs.
func TestFormatGoFallsBackForSnippets(t *testing.T) {
	in := "func f(){\nreturn 1\n}"
	out, _ := Format("go", in)
	if !strings.Contains(out, "    return 1") {
		t.Fatalf("the snippet was not indented:\n%s", out)
	}
}

// JSON is re-indented with the standard encoder's layout.
func TestFormatJSON(t *testing.T) {
	out, changed := Format("json", `{"a":1,"b":[1,2]}`)
	if !changed {
		t.Fatal("minified JSON was left alone")
	}
	if !strings.Contains(out, "\n  \"a\": 1") {
		t.Fatalf("got:\n%s", out)
	}
}

// JavaScript gets more than indentation: the space before a block's brace and
// around an operator are what make a pasted snippet read as code.
func TestFormatJavaScriptSpacing(t *testing.T) {
	in := "function f(){\nif(x){\nreturn 1+2\n}\n}"
	out, _ := Format("javascript", in)
	for _, want := range []string{"function f() {", "if (x) {", "return 1 + 2"} {
		if !strings.Contains(out, want) {
			t.Errorf("got:\n%s\nwant it to contain %q", out, want)
		}
	}
}

// A call keeps its parenthesis against its name, while a keyword takes a space.
func TestFormatDistinguishesCallsFromKeywords(t *testing.T) {
	out, _ := Format("javascript", "if(a){f(b)}\n")
	if !strings.Contains(out, "if (a) {") {
		t.Errorf("a keyword's parenthesis should be spaced:\n%s", out)
	}
	if !strings.Contains(out, "f(b)") {
		t.Errorf("a call's parenthesis should hug its name:\n%s", out)
	}
}

// JavaScript is indented from its brackets; a closer on a same line as an
// opener is one level, not two.
func TestFormatJavaScriptIndent(t *testing.T) {
	in := "function f() {\nreturn [\n1,\n2\n]\n}"
	out, _ := Format("javascript", in)
	lines := strings.Split(out, "\n")
	if lines[1] != "    return [" {
		t.Errorf("line 1 = %q, want one level", lines[1])
	}
	if lines[2] != "        1," {
		t.Errorf("line 2 = %q, want two levels", lines[2])
	}
	if lines[5] != "}" {
		t.Errorf("last line = %q, want no indent", lines[5])
	}
}

// A Python block indents after its colon, and a dict literal does not.
func TestFormatPythonBlocks(t *testing.T) {
	in := "def f(x):\nif x:\nreturn {\"a\": 1}\n"
	out, _ := Format("python", in)
	if !strings.Contains(out, "    if x:") {
		t.Fatalf("the if was not indented under def:\n%s", out)
	}
	if !strings.Contains(out, "        return {") {
		t.Fatalf("the return was not indented under if:\n%s", out)
	}
}

// A line that closes its own bracket is not left over-indented.
func TestFormatDoesNotDoubleIndentClosers(t *testing.T) {
	in := "if (a) {\nb();\n} else {\nc();\n}\n"
	out, _ := Format("javascript", in)
	if strings.Contains(out, "\t} else {") {
		t.Fatalf("the closing brace kept an indent:\n%s", out)
	}
}

// Formatting an already-formatted block changes nothing, so the button can be
// pressed twice without drifting.
func TestFormatIsIdempotent(t *testing.T) {
	for _, lang := range []string{"javascript", "python", "rust", "css", "yaml", "sql", "go", "json"} {
		in := "function f() {\n\tif (x) {\n\t\treturn 1\n\t}\n}"
		once, _ := Format(lang, in)
		twice, changed := Format(lang, once)
		if changed {
			t.Errorf("%s: a second format changed the block:\n%s\nto\n%s", lang, once, twice)
		}
	}
}

// The language's fence name is read the way the renderer reads it.
func TestNormalizeLang(t *testing.T) {
	for in, want := range map[string]string{
		"Go": "go", "golang": "go", "js": "javascript", "ts": "typescript",
		"py": "python", "bash": "shell", "yml": "yaml", "sh": "shell",
		"rs": "rust", "kt": "kotlin", "rb": "ruby", "c++": "cpp",
	} {
		if got := NormalizeLang(in); got != want {
			t.Errorf("NormalizeLang(%q) = %q, want %q", in, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// What must never change

// A string literal is copied byte for byte. This is the single most important
// property the formatter has: a rule that reached inside a quoted string would
// change what the code says, which is worse than leaving it badly spaced.
func TestFormatNeverTouchesLiterals(t *testing.T) {
	cases := []string{
		`x="a  ,  b"`,
		`x='if (a) {b}'`,
		`x="a:b,c;d"`,
		"x=`a  b  ${c}`",
		`x="escaped \" quote and \\"`,
		`x="unterminated`,
	}
	for _, in := range cases {
		out, _ := Format("javascript", in)
		// The literal, as written, must appear in the output unchanged.
		start := strings.IndexAny(in, `"'` + "`")
		if start < 0 {
			continue
		}
		lit := in[start:]
		if !strings.Contains(out, lit) {
			t.Errorf("%q: the literal was rewritten\nin:  %s\nout: %s", in, in, out)
		}
	}
}

// A comment is copied as written, including the spacing inside it.
func TestFormatNeverTouchesComments(t *testing.T) {
	in := "x=1 // a  comment ,  with   spacing\n"
	out, _ := Format("javascript", in)
	if !strings.Contains(out, "// a  comment ,  with   spacing") {
		t.Fatalf("the comment was rewritten:\n%s", out)
	}
}

// YAML's space after a colon is not decoration: "c: 1" and "c:1" are different
// documents, so the formatter must not remove it.
func TestFormatKeepsYAMLColonSpace(t *testing.T) {
	out, _ := Format("yaml", "a:\nb:\n  c: 1\n")
	if !strings.Contains(out, "c: 1") {
		t.Fatalf("the value lost its space:\n%s", out)
	}
}

// Markup is not code. Its rules would turn "<div>" into "< div >", so the block
// is left exactly as it was written.
func TestFormatLeavesMarkupAlone(t *testing.T) {
	for _, lang := range []string{"html", "xml", "svg", "markdown", ""} {
		in := "<div class=\"a\"><p>hi</p></div>"
		out, changed := Format(lang, in)
		if changed || out != in {
			t.Errorf("%s: markup was reformatted:\n%s", lang, out)
		}
	}
}

// The output says what the input said: no token is added, removed or moved.
// This is checked by comparing the non-space characters of both, which catches
// the whole class of "the formatter ate a bracket" bugs.
func TestFormatPreservesEveryToken(t *testing.T) {
	cases := []struct{ lang, code string }{
		{"javascript", "const f=(a,b)=>{return a+b}"},
		{"typescript", "interface A{x:number}\nfunction g<T>(a:T):T{return a}"},
		{"rust", "fn main(){let v:Vec<i32>=vec![1,2];}"},
		{"python", "def f(a,b):\nreturn {'k':a+b}"},
		{"css", "a{color:red;background:blue}"},
		{"sql", "SELECT a,b FROM t WHERE a=1 AND b<>2"},
		{"yaml", "a:\n  - 1\n  - 2"},
	}
	for _, c := range cases {
		out, _ := Format(c.lang, c.code)
		if squeeze(c.code) != squeeze(out) {
			t.Errorf("%s: tokens changed\nin:  %q\nout: %q", c.lang, squeeze(c.code), squeeze(out))
		}
	}
}

// squeeze removes every space and tab, so two texts can be compared for having
// the same tokens.
func squeeze(s string) string {
	return strings.NewReplacer(" ", "", "\t", "").Replace(s)
}

// A blank line inside a block stays a blank line, and no line keeps a trailing
// space.
func TestFormatKeepsBlankLinesAndDropsTrailingSpace(t *testing.T) {
	in := "a=1   \n\n\nb=2   \n"
	out, _ := Format("javascript", in)
	if !strings.Contains(out, "a = 1\n\n\nb = 2") {
		t.Fatalf("blank lines were collapsed:\n%q", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if line != strings.TrimRight(line, " \t") {
			t.Errorf("a line kept a trailing space: %q", line)
		}
	}
}

// The bug this test exists for: gofmt emits tabs, the page draws a tab at zero
// width, so formatting a Go block stored indentation that could never be seen —
// the block went flat the moment the button was pressed. Every formatter's
// output must therefore be free of tabs.
func TestFormatOutputHasNoTabs(t *testing.T) {
	cases := []struct{ lang, code string }{
		{"go", "func f() {\nif true {\nreturn\n}\n}"},
		{"go", "package p\n\nfunc g() {\n\tx := 1\n\t_ = x\n}\n"},
		{"json", `{"a":{"b":[1,2]}}`},
		{"javascript", "function f(){\nif(x){\nreturn 1\n}\n}"},
		{"python", "def f():\nif x:\nreturn 1"},
		{"rust", "fn main(){\nif true {\nlet x = 1;\n}\n}"},
		{"yaml", "a:\n  b: 1"},
	}
	for _, c := range cases {
		out, _ := Format(c.lang, c.code)
		if strings.Contains(out, "\t") {
			t.Errorf("%s: the output holds a tab, which the page cannot draw:\n%q", c.lang, out)
		}
	}
}

// And the case that started it: a Go block already formatted by gofmt, which is
// tab-indented in the file. Pressing Format must either leave it alone or change
// it into something visibly equivalent — never strip its indentation.
func TestFormatGoKeepsIndentationVisible(t *testing.T) {
	// A document's block holds spaces, as a loaded document's does.
	in := "func Model(d Display) Typography {\n    if d.Distance > 0 {\n        return Typography{}\n    }\n    return Typography{}\n}"
	out, _ := Format("go", in)
	if !strings.Contains(out, "        return Typography{}") {
		t.Fatalf("the nested statement lost its indentation:\n%s", out)
	}
	// The indentation is spaces, not tabs, so the page can draw it.
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "\t") {
			t.Fatalf("a line starts with a tab:\n%q", line)
		}
	}
}

// Formatting an empty or whitespace-only block is not an error and changes
// nothing.
func TestFormatEmptyBlock(t *testing.T) {
	for _, in := range []string{"", "   ", "\n\n", " \n \n"} {
		out, changed := Format("javascript", in)
		if changed {
			t.Errorf("%q was changed to %q", in, out)
		}
	}
}
