package code

import (
	"bytes"
	"encoding/json"
	"go/format"
	"strings"
)

// Formatting a code block.
//
// Go and JSON get the formatters that ship with Go: go/format is the printer
// gofmt runs, and encoding/json lays out JSON the way the standard library does.
// Their output is exact, so it is used as it comes.
//
// Every other language gets the formatter below, and its rules are worth
// stating, because the difference between a formatter and a vandal is whether
// it can be trusted with code it does not fully understand:
//
//   - It never reorders, renames, removes or adds a token. The output says
//     exactly what the input said.
//   - It never looks inside a string, a comment, a template or a raw literal.
//     Those are copied byte for byte. A regex that rewrote the inside of a
//     quoted string is how a formatter quietly breaks code.
//   - It changes two things: the whitespace between tokens on a line, and the
//     indentation of a line from the brackets and block markers around it.
//
// That is a smaller promise than prettier makes, and it is one this app can
// keep without a parser for every language it might be shown.

// Format returns code laid out for its language, and whether it changed.
//
// The output has its tabs expanded to spaces, and that is not cosmetic. gofmt
// and json.Indent emit tabs; this page's text engine draws a tab at zero width,
// so a block formatted by the standard library would be *stored* indented and
// *drawn* flat — which is what made the Format button look like it was erasing
// the indentation it had just added. Expanding on the way out keeps one
// invariant: what is in the document is what is on the page.
//
// It is also what makes the button idempotent. The document holds spaces, the
// formatter is given spaces and returns spaces, so a second press finds nothing
// to change and says so, truthfully.
func Format(lang, code string) (string, bool) {
	// A block with nothing in it has nothing to lay out, and saying it changed
	// would be a lie the caller acts on.
	if strings.TrimSpace(code) == "" {
		return code, false
	}
	out := FormatRaw(lang, code)
	out = ExpandTabs(out)
	return out, out != code
}

// FormatRaw is the formatter proper, whose output may hold tabs.
func FormatRaw(lang, code string) string {
	switch NormalizeLang(lang) {
	case "go":
		// A whole file formats exactly. A snippet without its package clause
		// does not parse on its own, so it is wrapped, formatted and unwrapped,
		// which keeps a document's fragment indented as gofmt would indent it.
		if formatted, err := format.Source([]byte(code)); err == nil {
			return string(formatted)
		}
		return formatGoSnippet(code)
	case "json":
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(code), "", "  "); err == nil {
			return buf.String()
		}
		return layout(lang, code)
	case "html", "xml", "svg", "markdown", "text", "":
		// Markup and prose are not code, and this formatter's rules would
		// mangle them: "<div>" is not the operator "<" beside a word, and a
		// Markdown heading is not a comment. Better to leave the block exactly
		// as it was written, and say so, than to reformat it into something
		// that is no longer the document.
		return code
	default:
		return layout(lang, code)
	}
}

// NormalizeLang reads a fence's language the way the renderer does.
func NormalizeLang(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexAny(l, "{ "); i > 0 {
		l = l[:i]
	}
	switch l {
	case "golang":
		return "go"
	case "py":
		return "python"
	case "js", "mjs", "cjs", "jsx":
		return "javascript"
	case "ts", "tsx":
		return "typescript"
	case "sh", "bash", "zsh", "shell", "shellscript", "console":
		return "shell"
	case "yml":
		return "yaml"
	case "md":
		return "markdown"
	case "rs":
		return "rust"
	case "kt":
		return "kotlin"
	case "rb":
		return "ruby"
	case "c++", "cc", "cxx", "hpp":
		return "cpp"
	}
	return l
}

// formatGoSnippet formats Go that has no package clause, as a document's
// fragment has none: it is wrapped in a package and a function, formatted by
// go/format, and unwrapped again.
func formatGoSnippet(code string) string {
	indented := layout("go", code)
	if !strings.Contains(indented, "func ") {
		return indented
	}
	wrapped := "package p\n" + indented
	formatted, err := format.Source([]byte(wrapped))
	if err != nil {
		return indented
	}
	return strings.TrimPrefix(string(formatted), "package p\n\n")
}

// ---------------------------------------------------------------------------
// The general formatter

// tokenKind is what a token is, as far as laying out code is concerned.
type tokenKind uint8

const (
	// tokAtom is a word: an identifier, a number, a keyword. It is never split.
	tokAtom tokenKind = iota
	// tokLiteral is a string, a character, a template or a raw literal, kept
	// exactly as it was written.
	tokLiteral
	// tokComment is a comment, kept exactly as it was written.
	tokComment
	// tokPunct is an operator or a separator.
	tokPunct
	// tokNewline is a line ending the source had, and the formatter keeps:
	// it decides where lines start, not how many lines there are.
	tokNewline
)

// token is one piece of source and what it is.
type token struct {
	kind tokenKind
	text string
}

// layout re-indents and re-spaces code for its language.
//
// It works in three passes, and keeping them apart is what keeps each one
// simple: split the source into tokens, decide the spacing between the tokens on
// a line, then decide how far in each line starts.
func layout(lang, code string) string {
	lang = NormalizeLang(lang)
	toks := lex(lang, code)
	var b strings.Builder
	depth := 0
	// lineStart is true until the first token of a line has been written, which
	// is when its indentation is decided.
	lineStart := true
	// prev is the last token written, for the spacing rule.
	var prev token
	havePrev := false

	indentUnit := "\t"
	if spaceIndented(lang) {
		indentUnit = "    "
	}

	writeIndent := func() {
		if depth > 0 {
			b.WriteString(strings.Repeat(indentUnit, depth))
		}
	}

	for i, tok := range toks {
		if tok.kind == tokNewline {
			// Trailing spaces are never wanted, and a blank line is blank.
			trimTrailingSpace(&b)
			b.WriteByte('\n')
			lineStart = true
			havePrev = false
			continue
		}
		if lineStart {
			// A line that begins by closing a bracket sits one level out, so a
			// trailing "}" lines up with the statement that opened it.
			if isCloser(tok.text) && depth > 0 {
				depth--
			}
			writeIndent()
			lineStart = false
		} else if havePrev && needsSpace(lang, prev, tok) {
			b.WriteByte(' ')
		}

		b.WriteString(tok.text)

		// The depth moves by what the token leaves open. A closer that started
		// its line was already counted above, so it is not counted twice.
		switch {
		case isOpener(tok.text):
			depth++
		case isCloser(tok.text):
			if depth > 0 && !startedLine(toks, i) {
				depth--
			}
		case tok.text == ":" && opensBlock(lang) && endsLine(toks, i):
			depth++
		}
		prev, havePrev = tok, true
	}
	return finish(b.String())
}

// startedLine reports whether the token at i is the first on its line, which is
// where a closer was already accounted for.
func startedLine(toks []token, i int) bool {
	for j := i - 1; j >= 0; j-- {
		switch toks[j].kind {
		case tokNewline:
			return true
		case tokComment:
			// A comment before a closer on the same line means the closer did
			// not start the line.
			return false
		default:
			return false
		}
	}
	return true
}

// spaceIndented reports whether a language indents with spaces.
func spaceIndented(lang string) bool {
	switch lang {
	case "python", "yaml", "ruby":
		return true
	}
	return false
}

// opensBlock reports whether a trailing colon opens a block.
func opensBlock(lang string) bool {
	switch lang {
	case "python", "yaml", "ruby":
		return true
	}
	return false
}

// endsLine reports whether the next token that matters is a line ending, so a
// colon at the end of a line opens a block and one inside an expression, as in a
// dict or a ternary, does not.
func endsLine(toks []token, i int) bool {
	for j := i + 1; j < len(toks); j++ {
		switch toks[j].kind {
		case tokNewline:
			return true
		case tokComment:
			continue
		default:
			return false
		}
	}
	return true
}

// needsSpace is the spacing rule between two tokens on one line.
//
// It is written as what wants a space rather than what does not, because the
// exceptions are far fewer that way: after a comma, around a binary operator,
// before an opening brace, and between two things that are each a word.
func needsSpace(lang string, prev, cur token) bool {
	// Nothing ever wants a space before these.
	switch cur.text {
	case ",", ";", ")", "]", ".", "?.", "::":
		return false
	}
	// Nothing ever wants a space after these.
	switch prev.text {
	case "(", "[", ".", "?.", "::", "@", "#":
		return false
	}
	// A call's parenthesis hugs its name; a keyword's does not.
	if cur.text == "(" {
		return keywordTakesParen(lang, prev.text)
	}
	if cur.text == "[" {
		// An index hugs its name; a list literal after a keyword does not.
		return keywordTakesParen(lang, prev.text)
	}
	// After a comma or a semicolon, always.
	if prev.text == "," || prev.text == ";" {
		return true
	}
	// A colon is tight before and spaced after, in the languages where it
	// separates a key from a value or a case from its body. In YAML the space
	// after it is not a style choice: without it, "c: 1" becomes "c:1", which
	// is a different value.
	if cur.text == ":" {
		return false
	}
	if prev.text == ":" {
		return colonSpacedAfter(lang)
	}
	// Around a binary operator, always.
	if isBinary(prev.text) || isBinary(cur.text) {
		return true
	}
	// Before an opening brace: in languages that brace their blocks, and after a
	// word in any language, since "return {" is a word and a literal while "{"
	// on its own is not.
	if cur.text == "{" {
		return bracesBlocks(lang) || prev.kind == tokAtom
	}
	// Two atoms side by side are two words: "return x", "type Foo".
	if prev.kind == tokAtom && cur.kind == tokAtom {
		return true
	}
	// A word after a closing bracket is a keyword: "} else", ") {", "] as T".
	if (prev.text == "}" || prev.text == ")" || prev.text == "]") && cur.kind == tokAtom {
		return true
	}
	// A literal or a comment beside anything readable.
	if prev.kind == tokLiteral && (cur.kind == tokAtom || cur.kind == tokLiteral) {
		return true
	}
	if prev.kind == tokAtom && cur.kind == tokLiteral {
		return true
	}
	if prev.kind == tokComment || cur.kind == tokComment {
		return true
	}
	return false
}

// colonSpacedAfter reports whether a colon takes a space after it. It does
// wherever the colon separates a key from a value or a label from its body;
// it does not in the languages that write it as part of a type, as Go's
// "map[string]int" and Rust's "Vec<T>".
func colonSpacedAfter(lang string) bool {
	switch lang {
	case "python", "yaml", "javascript", "typescript", "css", "scss", "less",
		"json", "ruby", "kotlin", "swift", "csharp", "java", "scala", "php":
		return true
	}
	return false
}

// bracesBlocks reports whether a language writes blocks with braces.
func bracesBlocks(lang string) bool {
	switch lang {
	case "python", "yaml", "ruby", "shell", "toml", "ini":
		return false
	}
	return true
}

// keywordTakesParen reports whether a word is a keyword whose parenthesis is
// separated from it by a space, as "if (" is in C-like code.
func keywordTakesParen(lang, word string) bool {
	switch lang {
	case "python", "yaml", "ruby":
		return false
	}
	switch word {
	case "if", "for", "while", "switch", "catch", "return", "typeof", "delete",
		"void", "throw", "case", "else", "do", "try", "finally", "with", "assert",
		"until", "unless", "elif", "match", "sizeof", "new":
		return true
	}
	return false
}

// isBinary reports whether an operator wants a space on either side.
func isBinary(op string) bool {
	switch op {
	case "=", "==", "===", "!=", "!==", "<", ">", "<=", ">=", "+", "-", "*", "/",
		"%", "&&", "||", "??", "=>", "->", ":=", "+=", "-=", "*=", "/=", "%=",
		"&=", "|=", "^=", "<<", ">>", "&", "|", "^", "**", "|>", "<>", "..":
		return true
	}
	return false
}

func isOpener(s string) bool { return s == "{" || s == "(" || s == "[" }
func isCloser(s string) bool { return s == "}" || s == ")" || s == "]" }

// ---------------------------------------------------------------------------
// Lexing

// lex splits source into atoms, literals, comments and line endings.
//
// The one thing it must never get wrong is where a literal ends. If a quote is
// misread, every rule after it applies to the inside of a string, which is how a
// formatter corrupts code that was fine. So each literal form the common
// languages have is spelled out, and an unterminated one stops at the end of its
// line rather than swallowing the rest of the block.
func lex(lang, src string) []token {
	var toks []token
	r := []rune(src)
	i := 0
	for i < len(r) {
		c := r[i]
		switch {
		case c == '\n':
			toks = append(toks, token{tokNewline, "\n"})
			i++
		case c == ' ' || c == '\t' || c == '\r':
			// Whitespace is the formatter's to decide, so it is dropped here.
			for i < len(r) && (r[i] == ' ' || r[i] == '\t' || r[i] == '\r') {
				i++
			}
		case c == '/' && at(r, i+1) == '/':
			j := i
			for j < len(r) && r[j] != '\n' {
				j++
			}
			toks = append(toks, token{tokComment, string(r[i:j])})
			i = j
		case c == '/' && at(r, i+1) == '*':
			j := i + 2
			for j+1 < len(r) && !(r[j] == '*' && r[j+1] == '/') {
				j++
			}
			if j+1 < len(r) {
				j += 2
			} else {
				j = len(r)
			}
			toks = append(toks, token{tokComment, string(r[i:j])})
			i = j
		case c == '#' && hashComment(lang):
			j := i
			for j < len(r) && r[j] != '\n' {
				j++
			}
			toks = append(toks, token{tokComment, string(r[i:j])})
			i = j
		case c == '-' && at(r, i+1) == '-' && dashComment(lang):
			j := i
			for j < len(r) && r[j] != '\n' {
				j++
			}
			toks = append(toks, token{tokComment, string(r[i:j])})
			i = j
		case c == '"' || c == '\'' || c == '`':
			j := literalEnd(lang, r, i, c)
			toks = append(toks, token{tokLiteral, string(r[i:j])})
			i = j
		case isWordRune(c):
			j := i
			for j < len(r) && isWordRune(r[j]) {
				j++
			}
			toks = append(toks, token{tokAtom, string(r[i:j])})
			i = j
		default:
			if op := longestOperator(r[i:]); op != "" {
				toks = append(toks, token{tokPunct, op})
				i += len([]rune(op))
				continue
			}
			toks = append(toks, token{tokPunct, string(c)})
			i++
		}
	}
	return toks
}

// at returns the rune at i, or 0.
func at(r []rune, i int) rune {
	if i < 0 || i >= len(r) {
		return 0
	}
	return r[i]
}

// hashComment reports whether a language writes comments with #.
func hashComment(lang string) bool {
	switch lang {
	case "python", "yaml", "shell", "ruby", "perl", "toml", "ini", "r", "makefile":
		return true
	}
	return false
}

// dashComment reports whether a language writes comments with --.
func dashComment(lang string) bool { return lang == "sql" || lang == "lua" }

// literalEnd returns the index just past the literal that starts at i, whose
// opening quote is q.
func literalEnd(lang string, r []rune, i int, q rune) int {
	n := len(r)
	// A triple-quoted string, in the languages that have one, ends at three of
	// its own quotes.
	if tripleQuoted(lang, q) && at(r, i+1) == q && at(r, i+2) == q {
		j := i + 3
		for j+2 < n {
			if r[j] == q && r[j+1] == q && r[j+2] == q {
				return j + 3
			}
			j++
		}
		return n
	}
	// A raw or template literal has no escapes: its own quote ends it.
	raw := q == '`'
	j := i + 1
	for j < n {
		switch {
		case r[j] == '\\' && !raw:
			j += 2
			continue
		case r[j] == q:
			return j + 1
		case r[j] == '\n' && !raw:
			// Unterminated: stop at the line rather than eat the block.
			return j
		}
		j++
	}
	return n
}

// tripleQuoted reports whether a language writes """ or ''' strings.
func tripleQuoted(lang string, q rune) bool {
	if q == '`' {
		return false
	}
	switch lang {
	case "python", "kotlin", "swift", "scala", "groovy":
		return true
	}
	return false
}

// operators are the multi-character operators read as one token, longest first
// so that "===" is not read as "==" and "=".
var operators = []string{
	"...", "===", "!==", "**=", "<<=", ">>=", "??=", "&&=", "||=", "?.",
	"->", "=>", "==", "!=", "<=", ">=", "&&", "||", "??", "::", ":=",
	"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "++", "--", "**",
	"<<", ">>", "|>", "<>", "..",
}

// longestOperator returns the longest operator at the start of r, or "".
func longestOperator(r []rune) string {
	for _, op := range operators {
		o := []rune(op)
		if len(o) <= len(r) && string(r[:len(o)]) == op {
			return op
		}
	}
	return ""
}

// isWordRune reports whether c can be part of an identifier or a number.
func isWordRune(c rune) bool {
	return isLetter(c) || (c >= '0' && c <= '9') || c == '_' || c == '$'
}

func isLetter(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c > 127
}

// ---------------------------------------------------------------------------
// Finishing

// trimTrailingSpace drops the spaces at the end of the current line, which a
// line that ended after a space would otherwise keep.
func trimTrailingSpace(b *strings.Builder) {
	s := b.String()
	trimmed := strings.TrimRight(s, " \t")
	if len(trimmed) == len(s) {
		return
	}
	b.Reset()
	b.WriteString(trimmed)
}

// finish drops the blank lines at either end of the block and makes it end with
// exactly one newline.
func finish(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// HasFormatter reports whether a language has a formatter. Markup and prose do
// not: the rules below would mangle them, so the command is not offered at all
// rather than failing when pressed.
func HasFormatter(lang string) bool {
	switch NormalizeLang(lang) {
	case "html", "xml", "svg", "markdown", "text", "":
		return false
	}
	return true
}
