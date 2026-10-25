# The icon

The mark is the letter M drawn as two columns of text: the M's stems are the
columns, its middle vertex the valley between them. It is the app's thesis in one
glyph, two pages side by side.

## Files

| File | What it is |
| --- | --- |
| `icon.svg` | The design, at 1024, as a reference. Not used by the build |
| `gen-icon.go` | Draws `icon-*.png` at every size |
| `gen-icns.go` | Packs those PNGs into `../resources/icon.icns` for macOS |
| `icon.png` | The 1024, which `mygo.json` names for Windows and Linux |

## Regenerate

```sh
go run assets/gen-icon.go
go run assets/gen-icns.go
cp assets/icon.png resources/icon.png
```

Then rebuild. The window icon is embedded in the binary through
`resources/icon.png` in `main.go`, so the taskbar and title bar carry it in a
plain `go build` too, not only in a packaged one.

## Why it is drawn in Go rather than rendered from the SVG

The icon needs a gradient plate, a cast shadow, and antialiased rounded corners.
The SVG parsers available here leave out filters, and a headless renderer on
Windows has no Cairo. Drawing it directly gives exact control at every size, and
Go is already the project's language.

## The sizes

The mark has three weights, because an icon is read at 16 pixels as often as at
512:

- **128 and up**: every line of text, six per column.
- **64 and 32**: three lines per column, every stroke 1.9× wider. At 32 pixels six
  hairlines and a stroke turn to mush.
- **16**: the letter alone, no text lines. Eight pixels of stroke cannot carry six
  hairlines.

The plate also grows from Apple's 824/1024 grid to nearly the full canvas at the
small sizes, because the standard margin eats a 32-pixel mark.
