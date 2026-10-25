# Muo — design review against Apple's HIG

> Reviewed with the `apple-design` skill against Apple's Human Interface
> Guidelines, and with Emil Kowalski's `apple-design` skill for the motion and
> craft rules. References read: `sidebars.md`, `toolbars.md`, `color.md`,
> `typography.md`, `layout.md`, `designing-for-macos.md`, `the-menu-bar.md`,
> `find-bar.md`, `form.md`, `button.md`, `accessibility.md`.

## Summary

**Good.** The thesis is unchanged and still the strongest thing about the app:
**the reading column is the product**, and the type model that sizes it is the
one idea in the window that no other reader has. This pass made that idea
visible — the readout is a gauge now, not a coloured string — and removed the
defect that was quietly working against it.

The defect was colour. The page was warm cream and the sidebar was the
toolkit's cool grey, so the two neutrals met at a hairline as two materials
glued together. A reader sits in this window for an hour; the seam was the one
thing that said "assembled". There is now one palette, and the page, the
sidebar, the title row and the settings pane are all cut from it.

## What changed

### The palette is one material (Critical → fixed)

- **What**: the page was `#faf8f5` and the sidebar `#f4f4f5`. Two neutrals from
  different hue families, 1 point apart in lightness, meeting at a 1-pixel rule.
- **Why**: `color.md › Best practices`: *"Avoid using the same color to mean
  different things… make sure all your app's colors work well in light, dark,
  and increased contrast contexts."* A warm surface and a cool one in the same
  window is the same colour doing two jobs.
- **Fix**: `theme.go › palette` is the only place a colour is written. The page
  reads its own surface from it, so `readingSurface` and `ShellTheme` cannot
  drift. Every text colour clears 4.5:1 against both the paper and the chrome,
  computed rather than estimated: ink 15.9:1, muted 5.6:1, accent 6.7:1 in the
  light; 10.5:1, 5.2:1, 6.6:1 in the dark.

### The accent was a brand, and the app has no brand (High → fixed)

- **What**: the accent was warm orange `#b4530a`, on the "warm paper" theory.
  It fought the paper instead of sitting on it, and it filled the Done button
  in a pane where nothing is being committed.
- **Why**: `branding.md`: brand colour is for primary actions and status, and it
  should defer to content. A reader has no brand; it has a pen.
- **Fix**: an ink blue, `#3a5a7d` light and `#7aa3cc` dark. It marks links, the
  chosen row, the focus ring and the one real primary action. In the dark, the
  *filled* control takes a deeper blue (`#35618f`) than the *text* accent, so
  white on it is 5.9:1 while links on the page stay 6.6:1 — one accent, two
  weights, because a pale block of blue was the brightest thing on screen.

### The readout is the signature, so it became one (High → fixed)

- **What**: `0.32° · 69ch` in green 11 pt text. The one number the app computes
  was legible to someone who already knew what it meant.
- **Why**: the model does not produce a number, it produces **a position in a
  band**. `0.26°`–`0.38°` is the band; the grade is whether you are in it.
- **Fix**: `gauge.go` draws the band and a marker on it, so a glance answers
  "is this comfortable" before the number is read. The number stays, because a
  reader who wants it should have it, and it is in tabular figures so it does
  not shuffle as the type steps. The row is named for assistive technology:
  *"Apparent character height 0.32 degrees, in the comfortable band; 69
  characters per line."*

### Settings showed a form where it could show type (High → fixed)

- **What**: three bare groups of controls on a blank page, with a
  accent-filled **Done** at the top right.
- **Why**: `settings.md`: a settings pane holds options; `form.md`: a form is
  labels and controls. Neither says the pane should be only that. These
  settings are *about how type looks*, and the pane asked the reader to change a
  number and then go back to the page to see what they had done.
- **Fix**: a specimen at the top, set at the size and measure being chosen, so
  the effect is where the control is. Each group sits on a card, so where one
  concern ends is a surface rather than a line of bold text. Each slider carries
  its value beside it, so a slider is read rather than estimated. **Done** is a
  plain button: nothing is being committed, since every change applies as it is
  made, so the way out is not the loudest thing in the pane.

### The heading scale had no top and a bottom below the floor (High → fixed)

- **What**: `1.62, 1.36, 1.16, 1.02, 0.95, 0.9` of the body size. At a 16.6 pt
  body that is 21.5, 22.6, 19.3, 16.9, 15.8, 14.9 — three near-identical
  levels at the top and two levels *smaller than the paragraph they introduce*.
- **Why**: `typography.md › Conveying hierarchy`: *"Adjust font weight, size,
  and color as needed to emphasize important information and help people
  visualize hierarchy."* A heading smaller than its body text is a footnote
  pretending to be a heading.
- **Fix**: `1.62, 1.36, 1.18, 1.06, 1.0, 1.0`, stopping at the body size. Past
  level four weight carries the hierarchy and size stays where the text is; the
  deepest level is set in the muted ink instead, which says "sub-heading"
  without shrinking it. Tracking follows the size (`-size × 0.012`), because one
  tracking value for every size is wrong at one end or the other.

### The sidebar truncated every heading (Medium → fixed)

- **What**: a 220 pt column against heading text that runs to 30 characters, so
  `Two colors, not maximum co…` was the normal case.
- **Why**: `sidebars.md`: *"If you need to include two levels of hierarchy in a
  sidebar, use succinct, descriptive labels."* The app cannot shorten a
  document's headings — they are the document's words — so it must give them
  room.
- **Fix**: 236 pt, and headings below the first are indented by their depth.
  Indentation is what makes the list read as the document's shape rather than
  as a flat list of strings, and it is the one place the app shows a hierarchy
  it knows.

### Find was promised and never built (Critical → fixed)

- **What**: `Cmd+F`, the Edit menu's **Find…**, and a **Find in Document**
  button all set `r.finding = true`. Nothing read it. There was no bar, no
  query, no count. A command that does nothing is worse than a missing one.
- **Why**: `the-menu-bar.md`: every command the window offers is reachable from
  the menu bar — and the corollary, that a menu item must do what it says.
- **Fix**: `find.go` counts matches in the source and puts the chosen one on the
  page by interpolating between the headings it has already measured: each
  heading's source offset comes from the parser, its pixel offset from the frame
  it was laid out in, so a match lands within a line. The bar is `ui.FindBar`,
  in the chrome's own row under the title, and the page makes room for it — a
  bar laid over the text hides the first match it is counting.

### Icons sat against the edge of their buttons (Medium → fixed)

- **What**: `ui.Button(c, "").Size(30, 26)` keeps the button's 14 pt horizontal
  padding, so a 16 pt icon is laid out in the 2 pt that are left. Every
  icon-only control in the title row was 6–7.5 DIPs off centre.
- **Why**: `layout.md › Visual hierarchy`: *"Align elements to make them easier
  to scan."* An icon that is almost centred reads as a mistake, not as a style.
- **Fix**: `icons.go › iconButton` zeroes the padding, and every icon-only
  command goes through it. `icons_test.go` measures the ink, not the intent: it
  renders a frame and compares the bounding box of the non-background pixels
  with the button's centre, so the offset is a number and not a hunch.

### The window could not shrink to fit a smaller display (Medium → fixed)

- **What**: `StateKey` restores the saved bounds. Opened on a 1536-wide display
  after being used on a 1920-wide one, the 1550-wide window puts the right-hand
  end of the title row — the readout and both commands — off the screen edge.
- **Why**: `windows.md`: a window resizes fluidly and never keeps information
  where it cannot be reached.
- **Fix**: `main.go › fitToDisplay` keeps the saved size but brings it inside
  the work area of the display it opens on.

### Two smaller ones

- **The empty page was a form.** A heading, a sentence, a button, and a
  three-line shortcut list under a rule. It is now the reading column's own
  measure, centred, with the recent files — the one thing there that does
  something the menu bar cannot. The shortcut is a keycap beside the button
  rather than a line of text.
- **Code blocks had a toolbar.** A bordered header with two bordered buttons
  above a bordered block is a box inside a box. The header is quiet now: the
  language in small caps, the commands with no face until hovered.

## What works, and should survive the next pass

- **One title bar.** `TitleBarHiddenInset` with the window controls over the
  sidebar; the content pane draws its own title row. No stacked bars.
- **The scroll edge, not a rule.** `glass.ScrollEdge` fades the content under
  the title bar. `toolbars.md`: *"use a ScrollEdgeEffectStyle when necessary to
  distinguish the toolbar area from the content area."*
- **The page is the only opaque surface.** The material shows through the
  sidebar. With one palette that separation is now a step in lightness rather
  than a change of hue, which is what it should always have been.
- **The measure is capped in characters, not pixels.** The one decision in this
  app that is genuinely better than the usual one, and every other choice
  should defer to it.
- **The gauge is the only thing in the window that measures anything.** Keep it
  singular. A second gauge would make the first one decoration.

## Craft notes

**The boldness is spent in one place.** The gauge is the signature: a small
instrument in the title bar that says whether this page is comfortable to read,
which is the question the whole app exists to answer. Everything around it is
quiet — one accent, one icon weight, one radius family, no ornament.

**Motion is a spring, and only where it is grabbed.** The sidebar slides on a
spring rather than a curve, because the command can be given again mid-flight
and the width should turn around where it is instead of stopping dead. That is
the only motion in the app, and it is the only one that needed to exist.

**Type is the content, so the chrome's type is small and still.** The chrome
runs 10–15 pt in five named steps; the page runs at whatever the model solves
for, which is the point. Nothing in the chrome competes with the page.

**What was removed.** The shortcut list on the empty page, the borders on the
code-block commands, the second accent weight, the accent fill on Done, the
rule between the title row and the page. Each was a small thing that, together,
made the window look busier than the page it holds.

## Platform notes

- Windows and macOS both get the inset title bar; on Windows the window controls
  are drawn by the platform into the strip the view leaves for them, which is
  why every pane's title starts at `titleLeft` rather than at its own padding.
- The sidebar hides itself below 820 DIPs of window width
  (`sidebars.md › Desktop`: *"Consider automatically hiding and revealing a
  sidebar when its container window resizes"*), and a reader who asks for it at
  that width gets it: a command that appears to do nothing is worse than a
  cramped page.
- Reduce Motion is honoured: the sidebar's spring becomes an instant move.
