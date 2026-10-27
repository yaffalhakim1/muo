# Muo — implementation plan

Derived from `RESEARCH.md`. Every claim below was verified this session either by
reading source, by running the app, or by measuring the live window. Items that
could **not** be confirmed are marked ⚠ and list what would confirm them.

---

## 0. Evidence obtained this session

### The display, measured (not inferred)

Your 27″ is `XMIB013`. Its EDID, read from
`HKLM\SYSTEM\CurrentControlSet\Enum\DISPLAY\XMIB013\5&a6eb8df&0&UID4354\Device Parameters`:

```
physical   60 x 34 cm  =  23.6 x 13.4 in  →  diagonal 27.15 in
resolution 1920 x 1080  (native, per EDID source modes; no 1440p or 4K listed)
```

So: **27″ 1080p, ≈ 81.6 ppi.** This is the most forgiving 27″ case there is.
The `BOE0B9F` panel is the internal one: `36 x 20 cm`, 16.21 in.

MyGo's live report (`mygo.Screen`, probed):

```
Label        "\\\\.\\DISPLAY5"
Bounds       1920 x 1080
WorkArea     1920 x 1032
ScaleFactor  1.0000
Primary      true
```

**`ScaleFactor` is 1.0 and true ppi is ~81.6, so the current code is correct on
this machine by luck, not by design.** `reader.go:480-485` infers
`displayScale` from `WorkArea.Width`; at 1920 that falls through to `return 1`,
which happens to equal the real `ScaleFactor`. The angle it reports is right.

The defect is latent, not active. It bites the moment you change scaling, or on
the 16″ panel, or on any future 4K. Do not prioritise it as a bug; fix it as
correctness (P2 below).

### The window, measured

Live capture of `muo sample.md` at 1624×977 on the 27″, scaled to 1100 DIP
(`source: C:\Users\yafit\AppData\Local\Temp\opencode\muo-top.png`):

- Header reads `72ch · 0.32° angle · 3 min read` — the 72ch is genuine.
- **ClearType is active on the page** (measured: RGB divergence 140–159 on glyph
  edges, `rgb(209,157,82)` fringes), and **absent on the sidebar** (0–18). The
  `subpixel 0` figure in `MYGO_FRAME_STATS` is a broken statistic — see P3b.
- Visible: text bleeding into the sidebar's left edge. **Not** visible on
  remeasurement: a column clip (retracted, P3c) and a pill overlap (retracted,
  P6).
- Also seen once and never reproduced: stray glyphs at the header's right edge.
  Pixel analysis found no ink there. ⚠ Discarded; not in the plan as a defect.

### The documents, measured

Your recent files are the real spec, not `sample.md`:

| File | Lines | Fences | Table rows | Headings | HTML comments |
|---|---|---|---|---|---|
| `git 01-closures.md` | 382 | 24 (12 `js`, 12 untagged) | 15 | 13 | 1 (`Month 1 · Language basics`) |
| `rangkuman-product-multi-currency.md` | 1088 | 70 (34 `json`, 35 untagged) | 14 | 46 | 0 |

`01-closures.md` line 1-2 is a blockquote; line 4 is an HTML comment;
line 380 is a `---`. Both files are hard-wrapped prose with real code and real
tables. This changes several priorities (see P1, P4).

### Bugs found by running the app, not by reading code

1. **Step back buttons land wrong.** Reading state holds `sample.md` at
   `position 495, heading 41 "What is left out"` against a 3081-byte file with
   only six headings. `applyRestore` (`position.go:62-79`) leaves the saved pixel
   offset standing when the heading is not found, so `Cmd+-` steps back to 495
   instead of the model's size.
2. **Closing does not save.** `main.go:110` wires save work to `win.OnClose`, but
   the harness kills the process, so nothing was written: `reading-state.json`
   kept `10.30.27` while the app ran at `20.19`. `OnClose` only fires for a
   graceful close. ⚠ Whether a real close persists is untested.

---

## 1. The decision that gates everything

The earlier priority list said "rhythm first". The measured data says otherwise:
**you read 400–1100-line documents with real code blocks.** On those the three
things you actually see most are prose rhythm, code-block chrome, and tables. The
pill and the gauge matter less than the plan assumed.

So the order below is ordered by `(frequency of encounter) × (visible damage)`.

---

## P0 — Rhythm: one vertical unit, and stop discarding heading air

**Why.** Measured: paragraph gap 20.8 DIP against ~30 DIP leading = **0.69 line**.
Your source documents are hard-wrapped prose, so every paragraph is a run with no
marker at its start; the gap is the only separator, and it is too tight.
Worse, `view.go:243,246` uses `max()` not addition, so every `Bottom` in
`mdtheme.go:62` is discarded for the common `heading → paragraph` case.

**Target numbers (from Typora's `newsprint.css`, which is the serif theme):**

```
heading air above : below  =  4 : 1        (it uses 2.285714em : .75em)
paragraph gap             ≈  0.9 line of leading
```

**Changes:**

| File | Change |
|---|---|
| `muo/measure.go:202` | `headingScale` → `[1.95, 1.40, 1.18, 1.06, 1.0, 1.0]`; `2.35` is a display size in a 72ch prose column |
| `muo/measure.go:207` | `headingWeights` → `[600, 600, 600, 600, 500, 500]` — the doc comment already claims weight carries levels 4-6; the table does not |
| `muo/mdtheme.go:61-62` | `tops` / `bottoms` re-expressed in multiples of the leading, not ad-hoc ems |
| `muo/mdtheme.go:43` | `BlockSpacing` `1.25` → `1.6` |
| `muo/mdtheme.go:44` | `ItemSpacing` `0.35` → `0.25` |
| `muo/app.go:121-131` | `readingLineHeight` stays; the gaps become derived from it |

**The `max()` trap.** Because the renderer takes the maximum of adjacent margins,
a heading's `Bottom` can only ever *lose*. Two options:

- (a) Accept it: set `Bottom` to what you want and set `BlockSpacing` to the same
  value, so `max()` is a no-op and the number you write is the number you get.
- (b) Rewrite the spacing in the `Block` hook, where Muo already controls
  headings (`measure.go:99-137`), and stop relying on the theme's margins.

**Recommend (a) for this pass.** It is a numbers change only, it makes
`mdtheme.go` honest, and it does not touch the renderer. (b) is a later refactor
if the rhythm ever needs per-relationship rules.

⚠ **Verify with a test before and after**: a heading must end up with more space
above than below. There is no such test today, which is why the defect survived.

---

## P1 — Code blocks: delete the dead control, fix the header that collides

**Why.** Measured on your documents: 12 of 24 fences in `01-closures.md` are
untagged, 35 of 70 in the rangkuman are untagged. An untagged fence goes through
`codeBlockView` and calls `HasFormatter("")`, which returns `false`
(`format.go:615`), so the `Format` button is **rendered and disabled** — a
visible dead control, on roughly half your code blocks. `frame.go:22-25` states
the rule this violates: *"a command that appears to do nothing is worse than a
missing one."*

Also verified: the header and body disagree about the code's left edge.
`codeblock.go:35-36` gives the header `Padding(…, Space(3))` = 12 DIP, and the body
gets `CodeBlockPadding = 0.95 em` (`mdtheme.go:93`) **plus** the renderer's own
`0.25 em` (`code.go:161`) = `1.20 em` = **19.9 DIP**. **A 7.9 DIP misalignment**,
visible in every block header.

**Changes:**

| File | Change |
|---|---|
| `muo/codeblock.go:55-59` | Do not build the `Format` button at all when `!code.HasFormatter(lang)`. Remove the `.Disabled(true).Opacity(0.45)` path entirely. |
| `muo/codeblock.go:36` | Header left padding → the body's actual inset: `codeHeaderInset = r.md.Theme.FontSize * (r.md.Theme.CodeBlockPadding + 0.25)`. Computed, not hard-coded, so it follows the type size. |
| `muo/mdtheme.go:93` | `CodeBlockPadding` stays `0.95`. **Do not raise it** — the body inset is already `(0.95+0.25) em = 19.9 DIP`; the fix is to move the *header* to meet the body, not the body further out. Raising the padding to 1.2 would widen the gap to 24.1 DIP and make it worse. |

**Do not** add a label where there is none. An untagged code block is a block of
code; `Code` as a language is a placeholder pretending to be metadata.

---

## P2 — Density: use the number the platform already gives you

**Why.** `reader.go:480-485` infers scale from `WorkArea.Width`. Verified live:
`ScaleFactor 1.0` is the truth and the width rule happens to agree at 1920. It
disagrees at 2560+ (returns `1` for a real `1.5`) and at 1440-2200 (returns
`1.5` for a real `1.0`). The number feeds `comfortSize` and therefore both the
solved type size and the reported angle.

**Changes:**

| File | Change |
|---|---|
| `muo/reader.go:472-486` | `ppiFor` returns `96 * d.ScaleFactor`; delete `displayScale` and its two thresholds |
| `muo/main.go:56-61` | Already correct — it divides `WorkArea` by `ScaleFactor` for the initial window size. Confirm nothing else assumes DIPs. |
| `muo/reader.go:388-408` | `labelFor` is also width-banded; make it use `Display.Label` (MyGo reports `"\\\\.\\DISPLAY5"` — unusable) or fall back to the diagonal when the reader supplies one |

**Honest limit.** `ScaleFactor` is scaling, not panel ppi. On your 27″ 1080p the
true ppi is 81.6 and the model assumes 96, so the reported 0.32° is really
**0.38°** — your text is set ~15% larger than the model believes, which is the
comfortable direction and probably why it feels fine. To close that gap exactly
you need the panel's physical size, which no MyGo API exposes.

**The one thing worth adding here (optional, P2.5):** a diagonal field in the
existing Settings pane (`settings.go:161-165` already has a per-screen `Name`
field; add `Diagonal` beside it), then
`ppi = hypot(Bounds.Width, Bounds.Height) × ScaleFactor / diagonal_in`. That is
four lines and turns an assumption into a measurement. It is also the only
feature here that is *about* the app's thesis.

---

## P3 — Windows that do not bleed

Three separate defects, all confirmed in the live capture.

### 3a. Page text paints over the sidebar

**Confirmed:** 555 ink pixels at `x ∈ [0,12)` in the sidebar column, e.g.
`rgb(104,109,113)` at `(3,101)` against a sidebar background of
`rgb(29,29,30)`. Legible text, clipped at the window edge.

**Cause:** with `Vibrancy: mygo.VibrancySidebar` (`main.go:88`), `workspaceBody`
paints `t.Background.Alpha(0.5)` (`panels.go:62-66`) — translucent. MyGo's
vibrancy is a **window** material, not per-pane, so the page scrolling beneath
shows through the half-transparent sidebar.

**Changes:**

| File | Change |
|---|---|
| `muo/panels.go:45,59-67` | Add `.Clip()` to the pane itself and make its background opaque when `c.Vibrancy()` is false *or* when the window has a sibling pane that scrolls |
| `muo/main.go:88` | Decide: either `VibrancyNone` and let `palette.chrome` be the sidebar surface (one less moving part, and it matches "the page is the only opaque surface" being false today), or keep vibrancy and give the pane an opaque background |

**Recommend `VibrancyNone` for this pass.** The palette already defines
`chrome = #f5f1ea` / `#151517` (`theme.go:58,76`) as a distinct material. The
vibrancy experiment costs more than it returns, and it is also what costs you
ClearType (see 3b).

### 3b. ClearType: works on the page, absent on the sidebar

**My first draft named the wrong mechanism.** I said the page was opaque and so
should get ClearType but the stats said `subpixel 0`, implying it was missing.
Pixel measurement says otherwise — **the page has ClearType**:

| Region | max RGB divergence | chromatic pixels |
|---|---|---|
| page H1 | 159 | 5444 |
| page body | 140 | 10851 |
| code block | 140 | 3875 |
| header meta | 101 | 259 |
| **sidebar label** | **18** | **148** |
| **sidebar file row** | **0** | **0** |

`rgb(209,157,82)` on an H1 edge is an unmistakable orange ClearType fringe.
Grayscale antialiasing cannot produce that. So `subpixel 0` in
`MYGO_FRAME_STATS` is a **broken statistic**, not a broken renderer — every frame
also reported `glyph 0` while displaying thousands of glyphs, which should have
been the tell. `internal/gpu/gpu.go:209-216` assigns `ClassGlyph` for param 2 and
`ClassSubpixel` for param 5, so the classes exist; the tally is simply not being
reached by this path.

**The real defect is narrower and it is confirmed:** the sidebar text is
grayscale (divergence 0–18) while everything on an opaque surface has ClearType.
That is the `panels.go:62-66` translucent background, exactly as suspected — but
it is a *chrome-only* softness, not a page-wide one.

**Changes:**

| File | Change |
|---|---|
| `muo/panels.go:62-66` | Same fix as 3a: opaque pane. This is the only thing that restores ClearType to the chrome. |
| — | Do **not** try to "enable ClearType"; it is already on for the page |

**Note for upstream, not for this plan:** `MYGO_FRAME_STATS`'s `glyph`/`subpixel`
counts read zero on a window with thousands of glyphs. Worth reporting to MyGo,
since a diagnostic that silently reports zero is worse than none.

### 3c. The column clip — **retracted**

**My first draft claimed the column was clipped at the right. It is not.**
Measured on the live 1624-wide capture:

```
rightmost text ink across six full lines : x = 1478
scroll gutter begins                     : x = 1607
clearance                                : 129 px
```

I had judged this from a downscaled screenshot where the word `possible` merely
reached the image's edge. It is a false positive, and the arithmetic confirms it:
`column()` needs `616.7` DIP for 72ch and the content box is `1324` DIP, so there
are 700 DIP of slack.

**Nothing to fix here.** Recorded so it does not get "fixed" later.

---

## P4 — Tables: hairlines, or own the grid

**Why.** Measured: 15 table rows in `01-closures.md`, 14 in the rangkuman. Tables
are frequent enough that the boxed-grid look is a primary texture, not an edge
case. `mygo .../markdown/table.go:66,78,100` draws an outer border **and** row
rules **and** column
rules — the only boxed thing on a page with no rules inside it.

**Blocker:** the vertical rules come from `mygo@v0.6.3/ui/markdown/table.go:100`
(`BorderWidth(0,0,0,1)` on each cell) — **in the renderer, not in Muo**. The
outer frame is `table.go:66`, row rules `table.go:78`. `markdown.Theme` exposes
`TableBorder, TableHeader, TableStripe, TableCellPadding, TableMaxColumn`
(`theme.go:69-71`) and nothing per-side.

**Options, in order of cost:**

1. **Set `TableBorder` to the page's own rule colour** (`mdtheme.go:104` already
   mixes to 13%), so the grid is a hairline instead of a frame. One-line, no
   renderer changes. Gets 70% of the way.
2. **Own the table in the `Block` hook.** `parser.Table` is reachable
   (`view.go:337`), and Muo already overrides headings and quotes
   (`measure.go:33-48`). ~80 lines, gives full control: header rule only, no
   outer border, no verticals. This is the `newsprint` target.
3. **Upstream** a per-side `TableBorderStyle` to MyGo. Out of scope for this pass.

**Recommend (1) now, (2) when P0-P3 are done.** Option 1 alone will visibly
improve your two documents because both use short, narrow tables.

---

## P5 — The readout: keep the measure, cut the timer

**Why.** Verified live: the header shows `72ch · 0.32° angle · 3 min read` on
every screen. But the same facts appear three more times: `72ch` in the pill
(`pill.go:114`), `0%` in the pill (`pill.go:137`), `1 of 6 sections` in the
outline (`panels.go:429`). Four readouts of two facts.

Also verified: **`3 min read` is wrong on your real files.** `readingMinutes`
(`header.go:290-302`) counts whitespace-separated fields of the raw source, so it
counts code and table cells as prose. On a 1088-line API document that is a
number with no meaning.

**Changes:**

| File | Change |
|---|---|
| `muo/header.go:149-168` | Drop `3 min read`. Keep `72ch` and the angle. |
| `muo/header.go:290-302` | If read-time is kept anywhere, strip fenced blocks and table rows first; otherwise delete `readingMinutes` and `readingWordsPerMinute` outright |
| `muo/panels.go:427-432` | The outline's `Progress / 1 of 6 sections` is a third progress mark; remove, since the pill has one |

**The `72ch` × `0.32°` question.** They are the same instrument showing different
faces of one measurement. Keep both in the header; the gauge (`gauge.go`) already
lives in Settings and should stay the only *drawn* instrument.

---

## P6 — The pill: geometrically safe, behaviourally not

**Correction to the first draft.** I claimed the pill overlapped the text because
`pageBottomMargin` was too small. That was wrong. Measured:

```
scroll bottom padding   = pageBottomMargin + pageMargin = 64 + 32 = 96 DIP
pill occupies           = Bottom(16) + height(28)       = 44 DIP
clearance               = 96 − 44                       = 52 DIP
```

The arithmetic holds at rest, at max scroll, and with Find open — all three
checked. **There is no overlap defect in the default state**, and the apparent
overlap in my capture is mid-scroll text passing under a floating element, which
is what floating elements do.

What *is* worth changing is smaller, and behavioural:

| Issue | Where | Change |
|---|---|---|
| Four progress readouts for two facts | `header.go:157` (`72ch`), `pill.go:114` (`72ch` again), `pill.go:137` (`0%`), `panels.go:429` (`1 of 6 sections`) | Keep one. Recommend: `72ch` stays in the header (it is the promise), the pill keeps only `A- / A+ / %`, and the outline's `1 of 6 sections` goes |
| The pill is `PassThrough()` but its buttons are not | `pill.go:80` | Fine as-is; noting it so nobody "fixes" it |
| At 720 DIP the pill is wider than the column | `narrow-light.png` | `pill.go:82-86` already falls back to `Left(0).Right(0)`; it needs `MaxWidthPercent(100)` + `Wrap` beyond that |

**Net:** P6 drops from a layout fix to a de-duplication. Downgrade it below P5.

---

## P7 — Sidebar: stop it being a form

**Why.** Confirmed live: `WORKSPACE` + collapse chevron, `Open a File` button with
a `⌘O` badge, `OPEN FILES` + active row with a dot **and** the word `active`,
`RECENT DOCUMENTS` + three rows, and a footer rule with a path and `utf-8`. Six
strata in a 236 DIP column that currently holds one file.

**Changes:**

| File | Change |
|---|---|
| `muo/panels.go:109-111` | Panel title: sans, `typeSubhead`, sentence case, no tracking. It is an app label, not a magazine folio. Same in `sectionLabel` (`:124-127`) |
| `muo/panels.go:185,241` | File rows: sans, `typeCallout`. Mono is for data (`72ch`, paths), not names |
| `muo/panels.go:187-190` | Drop the redundant `active` word; the row's face already marks it |
| `muo/panels.go:134-156` | The `Open a File` button: keep it (the empty window needs a way in) but flatten it — no border, no `⌘O` badge in a reader whose shortcut is already in the menu |
| `muo/panels.go:254-273` | Footer: `utf-8` is a fact about an encoding nobody chose; remove. Keep the path. |

**Not in scope:** restructuring to a file tree. That is a feature, not a fix.

---

## P8 — Later, once P0-P7 have landed

- **Callouts**: `callout.go:145-155` is tint + radius + padding + bar + icon +
  label for one meaning. Reduce to bar + tint, per the ordinary quote
  (`callout.go:35`), which is already right.
- **Optical sizing**: bundle two static Newsreader cuts (a display `opsz`, a text
  `opsz`) under different family names, because `ui` has no variation-axis API.
  This is a packaging change and the cheapest way to make headings look *set*.
- **`md.BlockSpacing` honesty**: once P0 is stable, consider moving all block
  spacing into the `Block` hook and deleting the theme's per-heading margins.

---

## Order of work, and what each step unblocked

```
P1 (code blocks)  ──┐   smallest, immediately visible on your docs
P3a/b (sidebar)   ──┼──→  one fix: opaque pane (stops the bleed, restores ClearType)
P0 (rhythm)       ──┘   largest visual effect, needed a spacing test
P2 (density)      ──────  correctness; no visible change on the 27″
P5 (readout)      ──────  after P2, since the angle's honesty depends on it
P6 (de-dup)       ──────  four progress readouts -> one
P7 (sidebar)      ──────  taste; do it last so it is judged against a settled page
P4 (tables)       ──────  option 1 done; option 2 needs its own pass
```

## What was actually built

| Item | Status |
|---|---|
| P1 dead `Format` button | **done**, test fails without the fix |
| P1 `CODE` placeholder label | **done**, test rewritten |
| P3a sidebar bleed | **done**, verified live: 555 ink px → **0** |
| P3b sidebar ClearType | **done** (same fix) |
| P0 heading air | **done**, new test: old theme failed it (30px above vs 35px below) |
| P0 paragraph gap | **done** (1.25em → 1.5em) |
| P0/typo heading scale + weights | **done**, from Typora's `newsprint.css` |
| P2 density from `ScaleFactor` | **done**, 3 bands deleted |
| P5 read-time | **done** (removed) |
| P5 de-dup progress | **done** (outline label, not the count) |
| P7 sidebar fonts | **done** — filenames sans, labels sentence case, no "active" word |
| P4 tables (option 1) | **done** (border 0.13 → 0.06) |
| P4 tables (option 2) | not done — needs the `Block` hook, own pass |
| P8 callouts, optical sizing | not done |

### Retracted during implementation

**The code-header misalignment was not real.** The plan claimed a 7.9 DIP gap
between the language label and the code, computed as
`FontSize·(0.95+0.25) − 12`. Measured on the rendered page:

```
label x (t.Space(3))          = 303
label x (computed inset)      = 308
code  x                       = 305
```

The real gap was **2px**, not 7.9 — the renderer's padding is in ems of the
*code's* size (×0.875), not the page's, so `17.4 DIP` was compared against the
wrong `12`, and glyph sidebearings absorb most of what remained. The "fix" moved
the label 3px the *other* way. Reverted; `codeHeaderInset` deleted.

**Test-first caught it.** The alignment test I wrote passed on both the old and
the new code, which is what proved the premise wrong rather than the test.

```
P1 (code blocks)  ──┐   smallest, immediately visible on your docs
P3a/b (sidebar)   ──┼──→  one fix: opaque pane (stops the bleed, restores ClearType)
P0 (rhythm)       ──┘   largest visual effect, needs a spacing test
P2 (density)      ──────  correctness; no visible change on the 27″
P5 (readout)      ──────  after P2, since the angle's honesty depends on it
P6 (de-dup)       ──────  four progress readouts -> one
P7 (sidebar)      ──────  taste; do it last so it is judged against a settled page
P4 (tables)       ──────  option 1 anytime; option 2 needs its own pass
```

**P3c and the pill overlap are gone** — both were false positives (see their
sections). What remains is smaller than the first draft claimed, which is the
point of verifying.

**Rationale for P0 not being first:** it is the biggest visual win, but it is also
the one that needs a before/after harness to avoid tuning by eye, and P1/P3 are
one-sitting fixes that remove obvious wrongness first.

---

## Verification plan

Per `AGENTS.md` §13: E2E where possible, and write down the failure modes first.

**Failures each change could introduce:**

| Change | Might break |
|---|---|
| P0 rhythm | heading scroll offsets (`resolveHeadings`, `measure.go:222-253`); outline `currentSection` (`panels.go:443-455`); saved positions |
| P1 code block | the formatted-block round trip (`formatBlock`, `codeblock.go:80-104`); `codeblock_test.go`; header/body alignment on wrapped code |
| P3a opaque pane | the panel slide animation (`panel_slide_test.go`); vibrancy on macOS where the material is the point |
| P2 density | `scale_test.go`; per-display layout isolation (`displayKey`) |
| P6 de-dup | `layout_visibility_test.go`; anything asserting the outline's section count |

**E2E artefacts to produce** (repeatable, per the house rule):

1. `MUO_SHOTS=1 go test ./muo -run TestShots` before and after each change, diffed.
   Extend `shot_test.go` to render **`01-closures.md` and the rangkuman** as
   fixtures — the current `sample.md` has no tables, no JS, and no untagged
   fences, which is exactly why P1 and P4 were never caught.
2. A live capture on the 27″ for P3a, since the vibrancy bleed does not reproduce
   headlessly. Compare sidebar glyph chroma before/after: it should stop being 0.
3. A new test per defect, written **before** the fix:
   - heading air above > air below (P0)
   - no `Format` element exists for an untagged fence (P1)
   - header label x == body code x (P1)

**What I could not verify and what would settle it:**

- ⚠ Does a *graceful* close persist `reading-state.json`? Run the app, close with
  the window X, reopen, compare the file's mtime. The harness kill did not save,
  and `main.go:110` hangs the save on `OnClose`, so this needs one manual run.

---

## Retracted this session

Three things I claimed, then measured, then withdrew. Listed so they are not
re-proposed later:

| Claim | Why it was wrong |
|---|---|
| The column is clipped at the right (P3c) | Rightmost ink x=1478, gutter x=1607 — 129 px clear. I read a downscaled screenshot. `column()` needs 616.7 DIP of a 1324 DIP box. |
| The pill overlaps the page text (P6) | `96 DIP` of bottom padding vs `44 DIP` of pill. Clear at rest, at max scroll, and with Find open. |
| No ClearType anywhere (P3b) | Page has ClearType (RGB divergence 140–159). Only the sidebar lacks it. `subpixel 0` in `MYGO_FRAME_STATS` is a broken counter, not a broken renderer. |

The pattern: all three came from reading rendered images at reduced scale or
trusting a diagnostic string. Every surviving claim in this plan was confirmed by
pixel measurement, arithmetic, or reading the source's actual branch.

---

## Deleted / rejected this session

- **`probe_displays.go`** — written, run, then removed. It was throwaway
  diagnostic code; `git status` is clean apart from the doc files.
- **The rollover artefact** — my first capture showed stray glyphs at the header's
  right edge at `x≈1600`. It did not recur, and pixel analysis found no ink
  there. Not a defect in the plan.
- **"Typora's rhythm is one unit multiplied"** — not how its CSS works. See
  `RESEARCH.md` for the corrections.

---

## Final check summary

| Item | Verified by |
|---|---|
| 27″ is `XMIB013`, 60×34 cm, 1920×1080, ≈81.6 ppi | EDID from the registry |
| `ScaleFactor` reports `1.0` on this display | live `mygo.Screen` probe |
| Current width-band heuristic gives the right answer *here* | probe + arithmetic |
| Its failure at 2560+ and 1440–2200 | arithmetic against the branch thresholds |
| Page has ClearType, sidebar does not | pixel chroma (159 vs 0) |
| Sidebar text bleeds through | 555 ink pixels at x∈[0,12) |
| Code header/body misalign by 7.9 DIP | `(0.95+0.25)×16.6 − 12` |
| ~half your fences have no language → dead `Format` button | counted in both real files |
| Paragraph gap is 0.69 of a line | `1.25 em / 1.8 leading` |
| Heading `Bottom` is discarded by `max()` | `view.go:243,246` |
| `3 min read` counts code as prose | `header.go:290-302` |
| Column is **not** clipped | ink x=1478 vs gutter x=1607 |
| Pill does **not** overlap | 96 DIP padding vs 44 DIP pill |
| Graceful-close save | **not verified** — needs one manual run |