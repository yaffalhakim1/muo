# Rendered states of the window, for looking at a change rather than imagining it.

`MUO_SHOTS=1 go test ./muo -run TestShots` writes these. The page in `docs/` uses
four of them; the rest exist so a change to the theme, the outline or the find bar
can be compared against how it looked before.

| File | What it shows |
| --- | --- |
| `doc-wide-light.png`, `doc-wide-dark.png` | All three panes, at 1440 wide |
| `doc-light.png`, `doc-dark.png` | The page and the library, below the outline's threshold |
| `empty-light.png`, `empty-dark.png` | The empty window |
| `settings-light.png`, `settings-dark.png` | The reading settings pane |
| `find-light.png` | Find open, with matches marked |
| `blocks-light.png`, `blocks-dark.png` | One of every markdown block |
| `narrow-light.png` | 720 wide, where both panes yield |

`all-blocks.md` is the fixture `blocks-*.png` render: a document with a heading of
each level, lists, a table, code in two languages, a quote, a footnote and an image.
