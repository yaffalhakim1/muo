package code

import "strings"

// Tabs in a page.
//
// The text engine lays a tab out at zero width: measured through the same
// shaper the page draws with, "a\tb" and "ab" are the same width. That is fine
// for a word processor and wrong for code, where the indentation *is* the
// structure — a Go file drawn with its tabs invisible is a wall of statements
// with no blocks in it.
//
// So a document's tabs are expanded to spaces as it is loaded. Four columns is
// CommonMark's own tab stop, which is what the parser already uses to read an
// indented code block, so the expansion agrees with the document's meaning
// rather than imposing a new one.
//
// It is done once, at load, rather than per frame: the expansion is the same
// every time, and doing it here means the display, Find, the outline and the
// formatter all read one string and cannot disagree about where a character is.

// TabStop is how many columns a tab advances, as CommonMark's.
const TabStop = 4

// ExpandTabs replaces every tab with the spaces that reach the next tab stop,
// counting columns from the start of each line.
func ExpandTabs(s string) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + len(s)/4)
	col := 0
	for _, r := range s {
		switch r {
		case '\n':
			b.WriteRune(r)
			col = 0
		case '\t':
			n := TabStop - col%TabStop
			for i := 0; i < n; i++ {
				b.WriteByte(' ')
			}
			col += n
		default:
			b.WriteRune(r)
			col++
		}
	}
	return b.String()
}
