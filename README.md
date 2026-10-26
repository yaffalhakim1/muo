# Muo

A markdown reader that sizes type from your screen and your viewing distance, then
holds the line at 72 characters.

![Muo with a document open: the files panel on the left, the reading column in the middle, and the page outline on the right](docs/shots/reading-light.png)

## The idea

A point size means nothing on its own. What your eye sees is the angle a letter
covers, and that depends on how far away the screen is and how many pixels it packs
into an inch. Muo measures the display, takes your distance from it, and solves for
the size:

```
angle = 2 · atan( (capHeight / ppi / 2) / distance )
```

A 27-inch display at 65 cm and a 16-inch laptop at 42 cm need different point sizes
for the same apparent comfort. Neither of them is 16 px.

The column is capped in characters, not pixels. A pixel cap looks stable, but it
silently changes the character count whenever the type size changes, so it is right
on one screen and wrong on the other.

## What it does

- **A page, not a form.** One centred column of prose in a serif cut bundled with
  the app, with the files and the outline as panels either side. Both panels fold
  and slide, and the column keeps its measure while they do.
- **Find that marks the page.** Every match is highlighted where it sits, headings
  included. Enter walks them.
- **PDF search.** Find reads the page text and turns to the page the match is on.
- **Code you can tidy.** Go and JSON go through the standard library. Other languages
  get spacing and indentation that never reaches inside a string or a comment.
- **It remembers.** Your place in each document, which panels you folded, the files
  you had open, and when you last read each one.
- **One dependency.** Go, drawn on the GPU by [MyGo](https://github.com/egoist/mygo).
  No webview, no JavaScript runtime, no frontend build.

## The window

The header carries the document's name, what the type model solved for this screen,
and the commands. The two panels answer the two questions a reader has while
reading — which file, and where in this file — and either can be folded away with
⌘B or the header's toggle. Zen folds both panels and the status pill so the page
has the whole window; the header stays, so the way out is where you left it.

The measure is the same promise in every state: 72 characters, or as many as the
window can actually hold. The pill at the foot of the page reports the count, the
size, and how far through the document you are, and its two buttons step the type
size without leaving the page.

## Build

You need Go 1.24 or newer.

```sh
git clone https://github.com/yaffalhakim1/muo
cd muo
go build -o muo .
./muo sample.md
```

On Windows the binary is `muo.exe`.

For live reload while you edit:

```sh
go run github.com/egoist/mygo/cmd/mygo@v0.6.3 dev .
```

That rebuilds and restarts the window when a Go file changes.

## Release

`go.mod` pins the MyGo CLI as a tool, so `go tool mygo build` runs the version
this project was built against. One command makes every installer:

```sh
go tool mygo build -platform windows/amd64,darwin/universal,linux/amd64
```

That writes a Windows `.exe` and an NSIS setup, a signed macOS `.dmg`, and a
Debian package plus a `.tar.gz` with an install script. They land in `build/`,
one directory per platform.

Pushing a tag builds all of it on GitHub Actions and attaches the result to a
draft release you publish. See [`.github/workflows/release.yml`](.github/workflows/release.yml).

```sh
# Bump "version" in mygo.json first, then:
git tag v0.1.0
git push origin v0.1.0
```

The workflow refuses a tag that disagrees with `mygo.json`, so the release page
and the version the app reports cannot drift apart.

Installed apps update themselves. The manifests are signed with
`MYGO_UPDATER_PRIVATE_KEY`, which is set as a repository secret, and the public
half is in `mygo.json`. Keep the private key: losing it strands every installed
copy.

## Layout

```
muo/
├── pdf/       the PDF parser, interpreter and rasterizer, with no dependency on the app
├── code/      the code formatter and tab expansion
├── fonts/     the page's serif, bundled so the measure is the same on every platform
└── *.go       the reader: frame, panels, header, pill, settings, theme, find
```

`pdf` and `code` build standalone. The app imports them and nothing else.

## Tests

```sh
go test ./...
```

99 tests. They drive the window headlessly through MyGo's tester, so what is checked
is what the window actually draws rather than what the code intends. A few of them
exist because a bug got past a screenshot: icons off centre by six pixels, a Find that
counted matches and painted nothing, a heading smaller than the body text it
introduced.

## Design notes

[DESIGN_REVIEW.md](DESIGN_REVIEW.md) records the reasoning against Apple's Human
Interface Guidelines, including the two defects that shaped the current design: a page
and a sidebar cut from different neutrals, and a Format button that erased indentation
the page could not draw.

## Licence

MIT. See [LICENSE](LICENSE).
