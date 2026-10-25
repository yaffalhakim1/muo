package muo

import (
	"path/filepath"
	"strings"
)

// classify picks the renderer from the file name. It is deliberately dumb:
// a reader opens what the user points at and shows it one of three ways.
func classify(path string) Format {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return FormatPDF
	case ".md", ".markdown", ".mdown", ".mkd", ".mdx":
		return FormatMarkdown
	}
	return FormatText
}

// plainText strips what a plain text file does not need shown as markup.
func plainText(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// titleOf derives a window title from a path.
func titleOf(path string) string {
	base := filepath.Base(path)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return path
	}
	return base
}
