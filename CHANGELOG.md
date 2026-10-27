# Changelog

Each `## <version>` section becomes the release notes for that version, and the
update window shows it when an installed app offers the update.

## 0.3.0

The page is set in Inter, and the page's spacing has a rhythm.

- **Inter, and the whole window in one voice.** The reading column and the chrome
  are now the same face. Newsreader is still bundled and one line away: the two
  lead differently, weight differently and measure differently, so which one the
  page is set in is a decision about what the window is.
- **A page that reads in sections.** The space above a heading is now several
  times the space below it, so a heading belongs to the text it introduces. Every
  heading previously sat closer to the paragraph above than to the one beneath:
  the renderer takes the *maximum* of two adjacent margins, and the theme asked for
  a value smaller than the block spacing, so its own setting was discarded.
- **Leading follows the face.** A sans is set tighter than a serif at the same
  measure, and the page adapts rather than holding one number for both.
- **The sidebar is opaque.** A translucent panel let the page show through it, so
  words ran under the file list and the chrome lost its ClearType.

Fixed:

- **Find moved nothing when there was one match.** The hook that scrolls to a match
  only ran for blocks the renderer builds, so a match that was a heading was
  counted, marked, and never scrolled to.
- **Find counted matches it could not mark.** A match inside a code fence is not
  highlighted — marking it would show the tags — but the bar counted it, so `1 of
  2` sent you to a match with nothing lit. The bar now counts what the page can
  show, and walking it never lands on a fence.
- **A code block with no language wore a "CODE" label and a dead Format button.**
  An untagged fence has no language to name and nothing to tidy.
- **The apparent size was solved from a guess.** The display's scale was inferred
  from how wide the work area reported in DIPs, which is wrong by a third on a 4K
  monitor at 150%. It is read from the platform now.
- **The header carried a reading time** counted from the source's whitespace, which
  counts code and table cells as prose.

Removed: the reading time, the "CODE" placeholder, the dead Format command, and the
outline's duplicate progress label.

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
