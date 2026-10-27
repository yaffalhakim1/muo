# Why Muo does not feel like Typora yet

A diagnosis, not a plan of code. Every claim cites `file:line` in this repo or in
`mygo@v0.6.3`. Nothing here is implemented.

> **Provenance, stated plainly.** Muo claims in this file come from source read
> in this session. Typora claims in the first draft came from memory and were
> wrong in three places — corrected in "What Typora actually does" below, which
> now cites the shipped theme CSS (`typora/typora-default-themes`) and Typora's
> own documentation. Treat every Typora statement in this file as one that
> carries a URL; if it does not, it is an opinion, not a fact.

---

## The verdict

The direction is right and the engineering underneath it is stronger than most
readers. What is missing is not a feature and not a palette: **the page has no
composition.** Typora feels like a document because its measurement, its rhythm,
and its furniture all resolve to one grid. Muo solves a remarkable type model and
then spends it on a page whose blocks are each spaced by a different rule, whose
paragraph starts all line up, and whose chrome floats a progress pill on top of
the words.

Fix the vertical rhythm and the triad, and the same app will feel a full tier
higher. Everything else on this list is smaller than those two.

---

## What already agrees with Typora (keep it, do not "improve" it)

| Muo | Typora |
|---|---|
| Serif body, bundled, one file per platform (`serif.go:30`) | Serif body, consistent everywhere |
| Measure capped in characters, 72ch (`typography.go:141`) | ~fixed em measure, content column capped |
| Code block recessed from the page, never inverted (`mdtheme.go:87`) | Code block tinted, not inverted |
| Code block commands with no face until hover (`codeblock.go:45`) | Format/Copy quiet until hover |
| Gauge + band as the one instrument (`gauge.go:32`) | (Typora has no equivalent — this is Muo's own signature) |
| Two panels, both foldable, page keeps its measure (`frame.go:151`) | Outline pane, focus mode |

The thesis is intact. The gap is craft, not concept.

---

## The five things that cost the most

Ordered by how much of the "not nice enough" feeling each one explains.

### 1. The vertical rhythm is undefined (biggest)

This is the single reason the page reads as *rendered* rather than *set*.
Vertical space is currently decided in four independent places that share no
unit:

```
block spacing      md.BlockSpacing = 1.25 em      (mdtheme.go:43)
heading above      Top:   1.9 1.55 1.35 1.15 1.0 0.9  (mdtheme.go:61)
heading below      Bottom: .75 .55 .45 .36 .28 .24     (mdtheme.go:62)
list item spacing  md.ItemSpacing = 0.35 em       (mdtheme.go:44)
quote inner blocks the same 1.25 em, again          (view.go:238)
```

At 16.6 pt that is paragraph gaps of 20.8, and heading gaps of 31.5 below a
title and 14.3 below a body-size heading. Nothing relates those to the leading,
so there is no baseline and no ratio — the page is spaced by opinion at each
call site.

And the renderer's rule is `max()`, not addition (`view.go:243,246`):

```
h2 → p   = max(1.25, 0.55) = 1.25 em    ← the heading's own Bottom is discarded
p  → h2  = max(1.25, 1.55) = 1.55 em
```

So the space *below* a heading is always the paragraph gap, regardless of what
the heading table asks for. Every `Bottom` value in `mdtheme.go:62` is dead
code for the common case.

**What Typora does:** one vertical unit derived from the type, and every gap is
a multiple of it. Something like `--rythm: half a line`. A paragraph gap is one
unit, a heading is two above and one below, a section is four. That is why its
pages have a pulse you can feel without being able to point at it.

**The constraint that makes this hard here:** the renderer inserts margins as
`Margin(top, 0, 0, 0)` *between* blocks, and CSS — so also MyGo's flexbox —
collapses nothing. If you ask for the true gap you will get the true gap *plus*
the leading slack of the previous paragraph. The `1.25` is almost certainly
compensating for that. It is a fix that cannot work, because the compensation is
a constant and the slack is not.

**What it should be, concretely:**

```
fine unit      = em/4, the sub-rhythm for inside a block
block unit     = 1 lead                    (~30 DIP at 16.6 pt)
section unit   = 2–3 leads
heading above  = 2 leads (h2), 3 (h1)
heading below  = 1 fine unit  ← small, because leading does the rest
```

That inversion — heading almost tight to the paragraph it introduces, and the
*air above* carrying the split — is the Typora/print convention and it is
currently inverted here.

### 2. The chrome has three (arguably four) voices, and the page has none of them

Count the faces on the reading screen:

```
page prose     Newsreader serif
page headings  Newsreader serif
tabs           Segoe UI        (native tab strip, Windows)
sidebar rows   monospace       (panels.go:185,241)
sidebar labels monospace UPPER tracked  (panels.go:109,124)
outline rows   Segoe UI        (panels.go:388)
header meta    monospace       (header.go:173)
header name    monospace       (header.go:130)
footer         monospace       (footer.go:37,64)
pill           monospace       (pill.go:115)
code block     monospace
code label     Segoe UI        (codeblock.go:42)
```

Three families and at least six sizes inside a window that sells itself as
*type-first*. Typora also has three — serif body, mono code, sans chrome — but
it is ruthlessly consistent: **every piece of chrome is the same sans at the
same size**, and mono appears only inside code.

Muo's split is arbitrary rather than meaningful. Filenames are mono but outline
headings are sans. Header document name is mono but panel titles are mono
uppercase. There is no rule you could state, so a reader's eye has nothing to
learn.

**What it should be — the triad, with one rule each:**

| Voice | Family | Where | Never |
|---|---|---|---|
| Document | Newsreader | the column, nothing else | any chrome |
| Data | mono | numbers, code, paths, `.md` | prose, labels, names |
| Chrome | system sans | tabs, panel titles, rows, menus, buttons | numbers, code |

The current mono-on-sidebar-rows is the loudest offender: a list of files set in
monospace reads as a terminal listing, which fights "this is my manuscript".
Typora's file list is sans. Muo's own prompt (`AGENTS.md` §14) already names the
cap: *max 2 font families*, with data mono exempted as a third.

### 3. The `h1` and the measure are solving different problems

`headingScale = [2.35, ...]` (`measure.go:202`) at a 16.6 pt body is a **39 pt**
title. At 72ch with the same face, 39 pt makes `# The shape of a page` render as
344 DIP in a 659 DIP column — 52%. That is *not* why the shot looks odd; the shot
looks fine.

What is odd is the intent. `measure.go:196` says "The shape is the design's …
the rest step down quickly, because a serif carries hierarchy at a smaller size
difference than a sans does", and then the title is +95% and `h2` is +45%. The
comment and the table disagree, and the `h2` at +45% reads as *bold paragraph*,
not *section* — which you can see in `blocks-light.png`: `Title two`, `Title
three` and `Title four` are three near-identical grey-ruled lines.

Also visible there: every heading after the first is **weight 400 at the same
colour** (`headingScale` doc comment says weight carries hierarchy below level
four; `headingWeights` is `[500,500,500,500,500,500]` — it does not). So levels
four, five and six differ only by a point or two of size. That is the "unpolished
documentation" texture.

**Target:** three levels that are unmistakable and three that are deliberately
the same, with weight/colour — not size — doing the last work. `h1` somewhere
near `1.9–2.0×`, `h2` near `1.4×`, `h3` near `1.15×` with weight, `h4–h6` at body
size in 500/600/600 with colour stepping down.

### 4. Paragraphs start with no signal, and the gaps are doing the work for them

`sample.md` is hard-wrapped at ~78 columns (see `sample.md:9-17`). Hard wraps
are invisible in the source, and the renderer collapses them into one paragraph
— correct — but nothing then marks where a paragraph *starts*. The first line of
every paragraph sits flush left, at exactly the same x as the seventh line of
the previous one, and the only separator is 20.8 DIP of air.

Typora has the same flush-left paragraphs, and gets away with it because its
gap is a round `1.6 em` and its measure is fixed in em, so the gap-to-line ratio
is a constant you can feel. Muo's gap is a fraction of a *character* measure
(`BlockSpacing: 1.25` em of font size) against a leading of `1.8` — a ratio of
0.69. That is on the tight side of comfortable, and combined with the unmarked
starts it reads as one long run in `reading-light.png`.

**Two moves, either sufficient:**
- **Opening-line indent.** 1.5 em on every paragraph *after the first*, no gap.
  This is the print convention and it composes beautifully with a heading that
  sits tight below. It also means the ragged-left of the longest line stops
  being the only shape in the column.
- **Or** a wider, deliberate gap near `1.6 em` with the heading Below tightened
  to match.

Not both. Indent plus big gaps is a word-processor look.

Caveat: `ui` has no `TextIndent` on `Element` (no such method in
`mygo@v0.6.3/ui/element.go`). An indent would be `PaddingStart` on a wrapper, or
a first-span rewrite in `Block`. Worth knowing before choosing this.

### 5. The chrome floats over the words, twice

`reading-light.png`, bottom centre: the pill sits at `Bottom: Space(4)` =
16 DIP over the column (`pill.go:80`), centred on the text, opaque
(`pill.go:106`), with a border and a shadow. In the shot it covers the line
"Set the angle, and the point size follows from the hardware."

`frame.go:230` gives the page `pageBottomMargin = 64` DIP (`frame.go:43`) —
which is *less* than the pill's own height plus its inset plus the shadow. So the
last line of every document can sit under the pill, and the paragraph you read
at the end of any screen can sit under it too, because the pill is fixed while
the text scrolls.

`narrow-light.png` makes it worse: the pill is *wider* than the column at 720
wide, so it overhangs the text on both sides.

And there is a second reading-progress mechanism — the 2 DIP hairline across the
top (`pill.go:47`), plus "1 of 6 sections" in the outline footer
(`panels.go:429`), plus "0%" in the pill. **Three progress readouts**, one of
them a bar *in* the pill and one a percentage *in* the pill.

Typora has no progress UI at all, on the grounds that a reader knows where they
are. This is the one place where copying Typora's *restraint* buys the most:
pick one, and make it not touch the words.

---

## The next tier down

### 6. Tables have a box; the page has no lines at all

`table.go:66` draws `Border(1, ...)` around the grid **and** row rules (line 78)
**and** column rules (line 100). A full grid is the Excel texture, and it is the
only boxed thing on a page that otherwise has no rules inside it. The shot
`blocks-light.png` shows the contrast immediately: a 2×3 table rendered as a
boxed spreadsheet between two paragraphs of prose.

Typora, and every print convention, uses **horizontal rules only** — header
underline stronger, row rules hairline, no outer border, no vertical lines. That
is achievable *here*: the columns take `BorderWidth(0,0,0,1)` from `table.go:100`,
which the theme cannot reach, so the grid needs the `Block` hook (the parser path
`parser.Table` → `st.table()` in `view.go:337`) or an upstream change. Flagging
it because it is a visible box you cannot reach from `mdtheme.go`.

### 7. Callouts are cards

`callout.go:145`: `Radius(12)`, tinted fill, `Padding(3,3)`, plus a 3 DIP left
rule plus an icon plus a label. Five signals for a note. The shot
`callouts-light.png` shows five stacked cards of five different fills — a
dashboard, not a page. In Typora an alert is a quote bar and a tint, one signal.
Muo already gets the ordinary quote right (`callout.go:35`: hairline bar, muted
text); the alert is the over-designed sibling beside it.

### 8. The sidebar is a form

`panels.go:134-156`: "Open a File" is a bordered, filled, full-width button with
a plus icon *and* a `⌘O` badge. Then `OPEN FILES` with a row that carries a
green dot *and* the word "active" (`panels.go:184,188`). Then
`RECENT DOCUMENTS`. Then a footer rule with a path and `utf-8`
(`panels.go:260`).

That is six horizontal strata in a 256 DIP column holding one file. Typora's
sidebar is a tree: paths, chevrons, filenames, nothing else. The command to open
a file is ⌘O and the menu bar — it does not need to be a button in the sidebar
in a reader whose whole pitch is fewer controls.

Also: `.md` is printed with a green dot meaning "read from disk" (`header.go:136`)
and the same green dot appears in the sidebar meaning "active". One colour, two
meanings, in adjacent UI.

### 9. The header's middle is a marketing claim, permanently

`header.go:149-168`: `<name> · 72ch · 0.32° angle · 3 min read`, always on, in
mono, in the optical centre of the window.

At a glance this is the app talking about itself in the most valuable horizontal
real estate in the app. "3 min read" especially — Typora, Pages, Preview and
every print reader omit it, because a reader with a document open has decided to
read it and does not need a timer.

The gauge is the signature and should stay. The other two are the ones to argue
about. If they stay, they belong in the settings pane, or in the header only
until the first scroll.

### 10. Code-block chrome is a second toolbar

`codeblock.go:35`: an uppercase language label, then a `Format` button that is
**always present and disabled** when the language has no formatter
(`codeblock.go:57-59`). A visible-but-dead control is the anti-pattern the
project's own notes call out (`frame.go:22-25`: *"a command that appears to do
nothing is worse than a missing one"*). In the shot, `CODE` blocks show a
greyed `Format` and it reads as broken.

The renderer's own `DefaultCodeBlock` (`code.go:92-98`) puts language left,
Copy right, `Padding(em(0.35), ..., 0, em(pad))` — a thin row. Muo's version
adds a taller row with a second label style and a second button. That is one box
inside another, which `mdtheme.go` says it is avoiding.

### 11. Nothing on the page is clickable in a way you can see

No links in `sample.md`, so this is inference rather than observation — but
`mdtheme.go:31` sets `md.Link` to the accent and the palette's accent is ink
blue `#3a5a7d` on `#faf8f4`, which is 6.7:1 and reads as *quiet body text* at
16.6 pt in a serif. A link needs to be unambiguously a link. That usually means
either the accent is more chromatic than the ink, or links carry a decoration
(`DecorationThickness`, which `ui` supports — `element.go`).

---

## What Typora actually does

Sourced, not remembered. Files: [`night.css`](https://github.com/typora/typora-default-themes/blob/master/themes/night.css)
and [`newsprint.css`](https://github.com/typora/typora-default-themes/blob/master/themes/newsprint.css)
from `typora/typora-default-themes`; [`Width-of-Writing-Area`](https://support.typora.io/Width-of-Writing-Area/);
[`Write-Custom-Theme`](https://theme.typora.io/doc/Write-Custom-Theme/).

### Three corrections to what I said before

**1. "Typora fixes its measure in em, that's why you can feel the rhythm." — Wrong.**
Typora's default themes cap the measure in **pixels**, and the cap changes with
window width:

```css
/* night.css */
#write { max-width: 914px; }
@media only screen and (min-width: 1400px) { #write { max-width: 1024px; } }
@media only screen and (min-width: 1800px) { #write { max-width: 1200px; } }

/* newsprint.css */
#write { max-width: 40em; }
@media only screen and (min-width: 1400px) { #write { max-width: 914px; } }

/* github.css */
#write { max-width: 860px; }
@media only screen and (min-width: 1400px) { #write { max-width: 1024px; } }
@media only screen and (min-width: 1800px) { #write { max-width: 1200px; } }
```

So Typora grows the column from 914 to 1200 px past 1800 px of window, and its
`Night` theme sets `html { font-size: 16px }` and **no body `font-size` at all** —
the user's font-size preference is the only lever, applied through `rem` in the
theme (which is exactly what [Write-Custom-Theme](https://theme.typora.io/doc/Write-Custom-Theme/)
instructs theme authors to do: *"Put default font size into `html`, then for
elements like `h1` or `p`, use `rem`"*).

Consequence: **Typora's characters-per-line is not a constant.** It drifts as the
user changes type size and as the window crosses those breakpoints. This is the
failure mode `typography.go:141-155` and `DESIGN_REVIEW.md:165-167` exist to
prevent, and Muo is right and Typora is wrong here. My earlier "Typora = fixed em
measure" was the opposite of true.

**2. "Typora uses a serif body." — Only `newsprint.css`, and it is an optional
theme.** The default `Night` sets `font-family: "Helvetica Neue", Helvetica,
Arial, sans-serif` for `html, body, button, input, select, textarea`, and a
*separate* `"Lucida Grande", "Corbel", sans-serif` for `h1–h6`. Typora's
out-of-box look is sans-on-sans. The serif reading flavour people picture is
`newsprint` (`"PT Serif"`, `line-height: 1.5em`) or the `github` theme, which is
also PT Serif.

So "Typora = serif and Muo = serif, therefore same direction" is not the
comparison you are making. Your brief is closer to Typora's *newsprint/github*
themes than to Typora. Worth knowing before you tune to it.

**3. "Typora's rhythm is one vertical unit multiplied." — Half-wrong.** Its
default themes are a set of near-round pixel values that happen to be related,
not a derived scale:

```css
/* night.css */
html, body { line-height: 1.625rem; }            /* 26px */
p, ul, ol, pre, table, hr { margin-bottom: 1.5rem; }  /* 24px — NOT 26 */
h1 { font-size: 2.5rem; line-height: 2.75rem; margin-bottom: 1.5rem; margin-top: 2em; }
h2 { font-size: 1.63rem; line-height: 1.875rem; margin-bottom: 1.5rem; }
h3 { font-size: 1.17rem; line-height: 1.5rem; }
```

24 px of paragraph gap on 26 px of leading is a ratio of 0.92 of one line. That
is **much airier** than Muo's 20.8 DIP on ~30 DIP (0.69), and it is the single
most visible difference in the screenshots.

But `newsprint.css` does the opposite and is tighter and more disciplined:

```css
/* newsprint.css */
html, body { line-height: 1.5em; }
p, blockquote, .md-fences { margin-bottom: 1.5em; }
h1 { font-size: 1.875em; line-height: 1.3em; margin-bottom: 0.5em; }
h2, h3 { font-size: 1.3125em; line-height: 1.15; margin-top: 2.285714em; margin-bottom: .75em; }
h3 { font-weight: normal; }
h5, h6 { font-size: 1em; }
```

Look at that: `h1` `margin-bottom: 0.5em` and `h2/h3` `.75em` — **tight below the
heading, wide above it** — and `h3` is told apart by `font-weight: normal`, not
by size or colour. That is the exact inversion I described in item 1 as the
principal that Muo is missing, and here it is in the shipped theme CSS as a
number: `margin-top: 2.285714em` (48 px) against `margin-bottom: .75em` (12 px).
**A 4:1 ratio of air above to air below the heading.**

Muo's is 1.55 above and 1.25 below (`mdtheme.go:61-62`) — a 1.24:1 ratio, and
the 1.25 is not even honoured because of the `max()` in `view.go:243,246`. That is
the number that makes Muo's page read as a continuous list rather than as a set
of sections.

**4. What I said about tables, and what Typora does. — I was half right, and it
does not help you.** `night.css` has `th, td { border: solid 1px #474d54; }` —
a **full grid**, cells boxed. `newsprint.css` has `thead th { text-transform:
uppercase }`, `tr:nth-child(even) { background: #e8e7e7 }`, no cell borders. So
the "hairlines only" look is a *newsprint* property, not a Typora property.

Muo's table (`table.go:66,78,100`) is the `night` shape, on a page set in a serif
— which is the incoherent combination. The choice is whether to be a `night`
page (sans, grid, as Typora ships) or a `newsprint` page (serif, rule-and-grade).
Muo is the second thing everywhere *except* the table.

**5. What I said about callouts — Typora ships no alert rendering at all.**
Nothing in `night.css` or `newsprint.css` styles `[!NOTE]`. Typora draws it as a
plain quote containing the literal text. So Muo's callouts are a GitHub feature,
not a Typora one, and "make it like Typora" is not the available correction here.
"Make it one signal instead of five" is, and that stands on its own.

### What survives from the first draft

| Claim | Status after checking |
|---|---|
| Muo's vertical gaps come from four unrelated rules | ✓ source |
| Heading `Bottom` is silently discarded by `max()` | ✓ `view.go:243,246` |
| Muo has three families and no rule between them | ✓ source; **Typora also has 3** (`Night`: sans body, Lucida/Corbel headings, Monaco/Consolas code) |
| Typora keeps heading air above ≫ below | ✓ `newsprint.css`, 4:1 |
| Typora has no read-time and no progress UI | ✓ not in any default theme |
| Typora's sidebar is a file tree | ✓ `.file-library-node`, `.file-node-title` in `night.css` |
| Typora caps its measure in characters | ✗ **wrong** — pixels, 914/1024/1200 |
| Typora is a serif reader | ✗ **wrong by default** — sans; serif is `newsprint`/`github` |

### The uncomfortable conclusion

You asked me to build toward Typora. Checked against the shipped CSS, **two of
the three things you like about Typora are things Muo already does better**
(character measure, and a serif page as the *default* rather than an optional
theme), and the one thing Typora genuinely does better is **air**: 0.92 line of
paragraph gap instead of 0.69, and 4:1 heading air instead of 1.24:1.

So the honest target is not "become Typora". It is: **keep Muo's measure and
Muo's serif, take Typora's air budget, and drop the parts of Muo's chrome that
Typora would never ship** (the read-time, the three progress marks, the mono
file list).

---

## If you only do three things

> Superseded by **"You read on a 27″ and a 14″"** below, which reorders these for
> your actual setup. Kept because the reasoning for each still holds.

1. **One rhythm.** Derive every vertical gap from the leading, in multiples of
   a quarter-lead, and stop compensating with `BlockSpacing`. Decide the
   `max()` problem first, because it silently deletes every heading `Bottom`.
2. **The triad.** Chrome labels and rows to system sans; mono only for numbers,
   code, paths. Panel titles lose the uppercase-tracked treatment (it is an
   *app*, not a magazine folio).
3. **Get the pill off the words.** Either reserve its band in the page's bottom
   margin so text can never pass under it, or move it to the header and delete
   the hairline and the outline's "1 of 6" so there is one progress readout.

Then, in the second pass: table to hairlines, callouts to one signal, sidebar to
a tree, and the `h1`/`h2` step to something the eye believes.

---

## The ppi question, answered

Since you read on a 27″ and a 14″, this was the thing worth pinning down. It is
now pinned down, and the answer is better than I feared and worse than the
README claims.

### Finding A: MyGo exposes real per-monitor DPI, and Muo throws it away

`mygo.Display` carries `ScaleFactor float64` (`modules.go:217`), and it is the
**actual** scale factor, not a DIP guess:

| Platform | Source | What it is |
|---|---|---|
| Windows | `float64(dpi)/96`, dpi from `GetDpiForMonitor(m, MDT_EFFECTIVE_DPI)` (`internal/windows/system.go:341`, `window.go:353-361`) | the user's **scaling** for that monitor |
| macOS | `[NSScreen backingScaleFactor]` (`internal/darwin/system.go:169`) | the **backing scale** (1 or 2) |
| Linux | `gdk_monitor_get_scale_factor` (`internal/linux/system.go:239`) | GDK's integer scale |

Muo has this available at the point where it matters — `displayKey(w)` already
calls `mygo.Screen.DisplayMatching` (`reader.go:368`), and `displayScale(w)`
calls it again (`reader.go:478`) — and then **ignores `d.ScaleFactor`** and
re-derives it from `d.WorkArea.Width` instead (`reader.go:480-485`).

The bars:

```go
// reader.go, today — three bands, two thresholds
case d.WorkArea.Width > 2200: return 1
case d.WorkArea.Width > 1400: return 1.5
return 2
```

versus

```go
// what the platform already reports
d.ScaleFactor   // 1.0, 1.25, 1.5, 1.75, 2.0, 2.5, 3.0
```

This is a straight substitution, and the ratio values above (`1.5`) are exactly
Windows' common scaling presets, which is strong evidence the author was
approximating `ScaleFactor` by hand when this was written.

**Why this matters concretely:** your 27″ gives a different answer under the two
rules depending on how Windows is set up. If it is a 1440p at 100%, `ScaleFactor`
is 1.0 and `WorkArea.Width` is 2560 — both rules agree, and `displayScale`
happens to return the right thing. If it is a 4K at 150%, `ScaleFactor` is 1.5
and `WorkArea.Width` is again ~2560 — the real answer is 1.5, and the width rule
returns `1`. A 33% error in the number the whole angle model is built on.

### Finding B: but `ScaleFactor` is *scaling*, not panel density — and Muo needs density

This is the subtlety that decides the design, so it is worth being precise.

`ScaleFactor` tells you how many **device pixels** make one **DIP**. It does
**not** tell you how many pixels per inch the panel has. On Windows those are
unrelated numbers: a 27″ 1440p panel is ~109 ppi whether you run it at 100% or
150% scaling; changing scaling changes `ScaleFactor` and changes how many DIPs
the work area reports, but does not change the panel.

The angle model needs **ppi**:

```
heightCm = fontPx * capHeightRatio / ppi * 2.54      // typography.go:98
angle    = 2·atan(heightCm / 2 / distanceCm)          // typography.go:107
```

`fontPx` is in DIPs. Both the physical size of a DIP **and** the physical size of
a pixel matter, and they pull in opposite directions:

```
ppi_effective = 96 / dipCm * 2.54     where dipCm = 2.54 / (ppi_effective)
```

So what the model wants is the **effective ppi**: how many pixels of text per
physical inch on that panel at that scaling. That is `96 × scale` in the naive
reading — but only true if the OS scaling is 100%. On a 4K 27″ at 150%, one DIP
is still ~1/96 inch of *rendered* size, and `96 × 1.5 = 144` is the effective
ppi, which is correct — and close to the true ~163 ppi divided by panel scaling.

**Net:** `ppiFor` returning `96 × d.ScaleFactor` is *closer to right than the
current width heuristic*, but it is still an assumption: it assumes the DIP is
approximately 1/96 inch, which holds on macOS and on Windows/macOS displays at
their recommended scaling, and drifts on unusual layouts.

**There are exactly three ways to get it exactly right, and only one is available
to you:**

1. **Ask the OS for the panel's physical size.** MyGo does not expose EDID or
   physical mm — `platform.Display` has `ID, Label, Bounds, WorkArea,
   ScaleFactor, Rotation, Internal, Primary` and nothing else (`platform.go:863-871`).
   **Not available.**
2. **Have the reader state their screen diagonal, once per display.** Then
   `ppi = sqrt(w_px² + h_px²) / diagonal_inches`, using pixel dimensions you can
   recover from `Bounds × ScaleFactor`. **Available today**, three fields of
   friction, once per monitor.
3. **Have the reader calibrate with a physical object** — "hold a credit card to
   the screen and size this rectangle". This is the standard trick (it is what
   Android's display calibration and various typography tools do). **Available
   today** and accurate to a couple of percent without knowing the diagonal.
   It subsumes distance-measuring in the same gesture if you offer both.

**Recommendation, and the reason it matters:** do (1) trivially by using
`d.ScaleFactor` instead of the width bands — it is a strict improvement, it
deletes code, and it is correct on both of your machines at recommended scaling.
Then add (2) or (3) behind the existing settings pane, for the reader whose 27″
is not at its recommended scaling. Option (3) is the one that fits the app's
character: it is a *measurement*, and the app is a machine for measuring type.

### Finding C: there is no per-pixel control, so "true 1:1" is not on the table

`ui.Context.Size()` returns DIPs only (`context.go:342`). There is no
`PixelRatio()` or backing-size accessor anywhere in `ui`. The rendering is
DIP-driven with the backend multiplying by DPI (`architecture.md:534-535`), and
the only knob is the window's own scale.

Consequence: **Muo cannot offer "render this text at 1:1 physical size,
bypassing OS scaling".** Precisely, the scenario where a 27″ 4K at 150% and a
14″ laptop at 200% both get a 16.6 pt body: on the 27″ it renders at
`16.6 × 1.5 = 24.9` device pixels, on the 14″ at `33` device pixels — different
physical sizes, because DIPs are not inches.

The only way to make them physically equal is to **solve the DIP value per
display**, which is exactly what a correct `ppi` gives you. So the answer to
"can we bypass scaling" is: you do not want to; you want to solve *inside* it.
That is what the model already does — it just needs honest ppi input.

### Finding D: "crisp vs washed" has a concrete cause on Windows

`internal/text/dwrite_windows.go:32-35`:

> with ClearType where the system smooths fonts with it **(on opaque
> backgrounds)**

and `raster.go:242`:

```go
subpixel := opaque && s.subpixel
```

Muo draws the page as an **opaque** surface — `readingColumnView` puts
`pal.paperColor()` on the scroll (`frame.go:229-231`). So the page **does** get
ClearType subpixel antialiasing on Windows. Good.

But the panel chrome uses vibrancy:

```go
// panels.go:62-66
if c.Vibrancy() {
	pane.Background(t.Background.Alpha(0.5))
} else {
	pane.Background(t.Background)
}
```

Translucent → grayscale. And `frame.go:131-135` makes the root transparent when
vibrancy is on. So on Windows 11 the sidebar text is grayscale while the page
text is ClearType. Two different rendering treatments side by side, and the
chrome looks softer than the page. That is a real, findable cause for "some of
it looks unfinished" that has nothing to do with layout.

Also relevant: **kyro/Newsreader's optical size axis is not being used.**
MyGo's font API takes a family and a weight (`element.go:1142`), and there is no
variation-axis or `opsz` support anywhere in `ui` (no `Variation`/`OpticalSize`
symbols). The bundled `newsreader.ttf` is a variable font with an `opsz` axis
(this is what `serif.go:27-28` claims it exploits) — **it does not.** So the 39
pt `h1` and the 16.6 pt body are drawn from the same optical cut, which is the
one thing optical sizing exists to prevent: a display cut used at text size, or
a text cut blown up to display size.

That cannot be fixed through `ui`'s API. It can only be fixed by **bundling two
static cuts** (an `opsz`-high one for headings, an `opsz`-low one for body) and
registering them under different family names. That is a packaging change, not a
code change, and it is the single cheapest way to make the page look like it was
*set* instead of rendered.

### What all of this means for the plan

The ppi finding **improves** the plan rather than threatening it:

- Using `d.ScaleFactor` is a 3-line change that makes the angle honest on both
  your screens at recommended scaling, and deletes the width heuristic.
- A diagonal or credit-card calibration in Settings makes it honest at *any*
  scaling, and is the one feature in this app that deserves the word "signature"
  more than the gauge does — because it is the reader participating in the
  measurement.
- Two static font cuts for heading vs body give you optical sizing that
  `ui` cannot otherwise provide.
- The vibrancy/opaque split explains a crispness difference you can see and
  cannot currently name.

None of these is a redesign. All four are small, and all four are on the exact
axis the app claims to compete on.

This is the real question, and it changes the priority order. A 27″ and a 14″
panel are not two sizes of the same thing: they differ in **density**, and Muo's
density estimate is a guess.

### Finding: the ppi the model trusts is inferred from a width threshold

> **Answered in full in "The ppi question, answered" below.** Summary: MyGo does
> expose the real per-monitor scale (`Display.ScaleFactor`); Muo ignores it and
> re-derives it from a width band. It cannot give you true panel density, because
> no OS API MyGo wraps exposes physical size — so exact ppi needs one calibration
> value from the reader per display.

`reader.go:474-486`:

```go
func displayScale(w *mygo.Window) float64 {
	d := mygo.Screen.DisplayMatching(mygo.Rectangle(w.Bounds()))
	switch {
	case d.WorkArea.Width > 2200:   return 1
	case d.WorkArea.Width > 1400:   return 1.5
	}
	return 2
}
```

`ppiFor` is `96 * displayScale(w)` (`reader.go:472`). So the physical density is
never read — it is inferred from how many DIPs the work area reports. Three bands,
two thresholds, no exceptions.

Concrete consequences for exactly your two screens:

| Screen | Work area (DIPs) | Inferred scale | Implied ppi | Typical true ppi |
|---|---|---|---|---|
| 27″ 1440p @100% | 2560 | **1** | 96 | ~109 |
| 27″ 4K @150% | 2560 | **1** | 96 | ~163 (physical, 146 effective) |
| 27″ 5K @200% | 2560 | **1** | 96 | ~218 |
| 14″ FHD @150% | 1280 | **2** | 192 | ~157 |
| 14″ 2.8K @200% | 1440 | **2** | 192 | ~242 |

The 27″ row is the problem: **every 27″ display from a cheap 1440p to a 5K
reports a work area near 2560 DIPs**, so all of them get ppi 96. If yours is a 5K
(or a 4K at 200%), the model believes the pixels are twice as large as they are,
and then `comfortSize` (`typography.go:114-117`) solves for a size with that wrong
ppi — so the *angle* it reports is wrong too, and the gauge grades a number that
was computed from a guess. On a 1440p it is very nearly right, which is why this
has not bitten you visibly.

Meanwhile `distance` is set per display class by `defaultDistance` (`reader.go:412-422`):
`< 1900 DIPs → 42 cm`, `< 2600 → 55 cm`, else `65 cm`. So a 14″ laptop gets 42 cm
and a 27″ gets 55 cm — reasonable defaults, both editable in settings.

**The honest framing:** your two screens are the *use case*, and the app's weakest
link is the number that is supposed to make them equivalent. `ppi` should be read
from the OS (or from the user, once per display) rather than inferred from a
width. Until it is, "the same apparent size on both machines" is a claim the app
makes and cannot fully deliver — and reading it on the 14″ vs the 27″ is exactly
where you would notice.

### Finding: the measure holds, the *column* does not — by design, and that's right

`typography.go:141-155` gives the character count priority over the margin, and
falls back to the viewport only when 72ch genuinely cannot fit. On a 14″ at a
smaller window that fallback fires and you get 59–64ch (you can see it in
`shots/callouts-light.png`, which reports `64ch` at 900 px wide). On the 27″ you
get the full 72.

That is the correct behaviour and it is the app's best idea. But it means the
reader gets a *different* column on each machine and the only place that is
explained is a 10 px number in the header. If the measure is the promise, the
window should say so when it falls short of 72 — that is a one-line change and it
is the highest-value "it feels designed" move available on your two screens.

### Finding: the type size is sticky per display, and the reset is buried

`reader.go:432-435`: a manual size wins over the model, per display key
(`displayKey` uses work area width/height/x/y). Drag the window from the 27″ to
the 14″ and the key changes, so each keeps its own — good. But `⌘0` to hand it
back to the model (`frame.go:322-327`) is not in the menu bar, and the same reset
in Settings (`settings.go:145-152`) is behind a conditionally-shown field. On a
two-screen setup you will hit this. Worth a menu item at minimum.

### So the priority order changes

If the 27″/14″ jump is the point, then:

1. **ppi honesty** — read it, or ask once per display. Without this the gauge and
   the solved size are decoration on any high-density 27″.
2. **Air** — the 0.69 → 0.92 line gap is what makes both screens feel
   cramped-versus-Typora, and it is the same fix on both.
3. **Say when the measure falls short of 72ch** and **make ⌘0 discoverable** — the
   two things that make the promise visible across two screens.
4. Everything from the earlier list (pill, tables, callouts, sidebar) — real, but
   second order *for your use case*.

## Your question: is the readout medicine or signature?

Neither, if it stays as it is. Right now it is three claims bolted to one row:

```
72ch  ·  0.32° angle  ·  3 min read
```

- `72ch` is **the promise**. It is the thesis, it is verifiable, and it means
  something different on each of your screens. This is the signature.
- `0.32° angle` is **the proof** — it is only interesting *because* the ppi it is
  derived from can be wrong (see above). Once ppi is honest, this earns its place
  as the evidence. Until then it is a number reporting its own guess.
- `3 min read` is **neither**. It is a feature from a different app, on the one
  row where the app can argue its thesis. Typora does not have it. Cut it.

So my read: **keep `72ch`, keep `0.32°` once the input is trustworthy, cut the
read-time.** The gauge in Settings (`gauge.go`) is the same instrument given room;
the header row is the same instrument given a footnote's worth of space. One
place should win — and the header is the wrong place for all three, because the
header is where the *document* should be.

That is the [medicine/signature] answer you asked for: the measure and the angle
are the medicine (the app's reason to exist), and neither is a signature until
someone can see, on a 5K monitor, that it is still telling the truth.

## One thing to decide before any of this

The project's own `DESIGN_REVIEW.md:171-176` says the boldness is spent in one
place and everything else is quiet. That is right. But "quiet" has been read as
"small, mono, and bordered", and small-mono-bordered reads as *tooling*, not as
*quiet*. Typora's quiet is a large serif page, a great deal of air, and a
sidebar made of nothing but filenames.

So the real choice is this: **Muo currently looks like a developer tool that
renders Markdown well. Typora looks like a book that happens to scroll.** The
type model is the argument for the second. The fix list above is how to make the
window stop arguing against its own thesis.

---

## Evidence index

- Screenshots regenerated this session: `shots/*.png` (`MUO_SHOTS=1 go test ./muo -run TestShots`)
- `docs/shots/*.png` in the README are **stale** — they show a header and footer the
  code no longer builds (`shots/README.md` vs `docs/shots/`, 4 files not refreshed)
- **MyGo API surface, verified**: `modules.go:209-250` (`Display.ScaleFactor`),
  `internal/platform/platform.go:863-871` (no physical size — no EDID, no mm),
  `internal/windows/system.go:341` + `window.go:353-361` (`GetDpiForMonitor`,
  MDT_EFFECTIVE_DPI), `internal/darwin/system.go:169` (`backingScaleFactor`),
  `internal/linux/system.go:239` (`gdk_monitor_get_scale_factor`),
  `ui/context.go:342` (`Size()` returns DIPs; no backing-pixel accessor)
- **Text rendering**: `internal/text/dwrite_windows.go:32-35` (ClearType only on
  opaque backgrounds), `internal/text/raster.go:242` (`subpixel := opaque && ...`),
  Muo's `frame.go:131-135,229-231` and `panels.go:62-66`
- **Optical sizing**: `ui/element.go:1142-1154` (family + weight only; no variation
  axis API anywhere in `ui`), Muo's `serif.go:27-28` claims `opsz` and
  `mdtheme.go` relies on it
- Density inference: `reader.go:412-422,472-486`
- **Typora, sourced**: `typora/typora-default-themes` → `themes/night.css`,
  `themes/newsprint.css`, `themes/github.css`; `support.typora.io/Width-of-Writing-Area`;
  `theme.typora.io/doc/Write-Custom-Theme`
- Rhythm: `mdtheme.go:43-44,61-62`, `view.go:238-247`, `app.go:121-131`
- Scales: `measure.go:202,207`, `typography.go:141-155`
- Chrome type: `panels.go:109,124,185,388`, `header.go:130,173`, `footer.go:37`,
  `pill.go:115`, `theme.go:97-105`
- Overlay: `pill.go:47,80,92,106`, `frame.go:43,230`, `panels.go:429`
- Boxes: `table.go:66,78,100`, `callout.go:145-155`, `codeblock.go:35-59`,
  `panels.go:134-156,260-273`
- Renderer limits found in `mygo@v0.6.3`: no `TextIndent`, no column-less table
  option, `DefaultCodeBlock` padding is fixed at `em(0.85)` on the left,
  `block()` margins are `max()` not sum (`view.go:243,246`)