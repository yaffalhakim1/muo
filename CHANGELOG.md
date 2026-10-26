# Changelog

Each `## <version>` section becomes the release notes for that version, and the
update window shows it when an installed app offers the update.

## 0.2.0

The window is a page now.

- **A serif page.** The reading column is set in Newsreader, bundled with the app
  so the measure is the same on every platform. Headings are drawn in it too, at
  the size and weight the design calls for.
- **One header, two panels.** The document's name, the measure, the apparent size
  and the reading time sit above the page; the files and the outline are panels
  either side of it. Both fold with ⌘B or the header's toggle, and slide.
- **Zen.** Folds both panels and the status pill so the page has the whole window.
  The header stays, so the way out is where you left it.
- **A status pill** at the foot of the page: the measure, A- and A+ for the type
  size, and how far through the document you are.
- **Reading time**, from the document's word count.
- **Recent documents** in the files panel, with how long ago each was read.
- **GitHub callouts.** `> [!NOTE]` and the other four alerts are drawn as marked
  passages rather than showing their marker as text.
- **A new palette.** Warm paper and two inks — a darker one for headings, a softer
  one for prose — held inside the contrast band reading wants rather than at
  maximum.

Fixed:

- A heading's position doubled as the page scrolled, so the outline's idea of where
  you were drifted further from the page the further you read, and a saved place
  could point past the end of the document.
- The outline now marks the section you are actually in, measured against the
  header rather than the raw scroll.
- A table column as narrow as a short heading broke it mid-word.
- The header clears the window's controls on Windows, where they are on the right
  rather than the left.

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
