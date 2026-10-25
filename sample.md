# The shape of a page

A reader is a machine for one job: getting words from a screen into a head with
as little friction as possible. Everything else — the fonts, the colors, the
window chrome — exists to keep that one job from being interrupted.

## The measure comes first

The single largest effect on reading comfort is how many characters sit on a
line. Too few and the eye finishes before it has parsed; too many and the
return sweep to the next line loses its place, so the reader re-reads the same
line without noticing. The band that works is about 45 to 75 characters, and
dense technical prose does well at the top of it.

That is why this column is capped in characters and not in pixels. A pixel cap
looks stable, but it silently changes the character count whenever the type
size changes — so it is right on one screen and wrong on the other.

## Size is an angle, not a number

A point size means nothing on its own. What the eye experiences is the angle a
letter subtends, and that depends on how far away the screen is and how many
pixels it packs into an inch.

    angle = 2 · atan( (capHeight / ppi / 2) / distance )

Set the angle, and the point size follows from the hardware. That is the whole
trick behind switching between a large monitor and a laptop without the text
feeling suddenly wrong.

> A 27-inch display at 65 cm and a 16-inch laptop at 42 cm need different point
> sizes for the same apparent comfort. Neither of them is "16 px".

## Two colors, not maximum contrast

Pure black on pure white is the maximum contrast a screen can produce, and it
is not the most comfortable. At that ratio the edges of letters bloom — a
halation effect that is especially strong for readers with astigmatism.

| Surface | Background | Ink       | Ratio |
| ------- | ---------- | --------- | ----- |
| Light   | `#faf8f5`  | `#1f1d1b` | 13:1  |
| Dark    | `#1a1a1c`  | `#cdc8c2` | 9:1   |

Both sit in the band that is easy to look at for an hour without being washed
out. Dark mode is fine; dark mode at 21:1 is a handicap.

## Leading

Line height has a sweet spot too. Below about 1.3 the ascenders and descenders
crowd each other. Above about 1.8 each line stops reading as part of a
paragraph and starts reading as its own small object. Between them — closer to
1.5 on short lines and 1.65 on long ones — the paragraph keeps its shape.

```go
// Model turns a display's calibration into a reader's type size.
func Model(d Display, viewport float64, ppi float64) Typography {
	size := comfortSize(ppi, d.Distance)
	const charsPerLine = 72.0
	if size > 40 {
		log.Fatal("an unreasonable distance")
	}
	return Typography{
		FontPx:    size,
		Line:      leading(size),
		ColumnW:   size * charsPerLine,
		ChPerLine: charsPerLine,
	}
}
```

- [ ] Read the measure section again
- [ ] Try the settings surface
- [x] Open a document

## What is left out

No theme gallery. No plugin system. No syncing. A reader that does one thing
has room to do it carefully, and every feature that does not serve the reading
column is a place where the column can get worse.
