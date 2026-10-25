# Changelog

Each `## <version>` section becomes the release notes for that version, and the
update window shows it when an installed app offers the update.

## 0.1.0

First release.

- Type size solved from the display's density and the distance you sit from it,
  with the column held at 72 characters.
- Three panes: files, page, and the document's own outline. Every divider drags,
  and every pane folds away and comes back where you left it.
- Find marks every match on the page, headings included, and walks them with
  Enter. In a PDF it reads the page text and turns to the page the match is on.
- Code blocks format in place: Go and JSON through the standard library, other
  languages through a spacing and indent pass that never reaches inside a string
  or a comment.
- Remembers your place in each document, the pane widths, and the files you had
  open.
- Markdown, PDF, and plain text.
