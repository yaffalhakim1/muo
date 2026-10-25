# Animation opportunities, Muo's landing page

Swept with `find-animation-opportunities` against `docs/index.html`. Read with the
gate: frequency, purpose, speed, function. Most candidates were rejected.

## Part 1 — what survives

| # | Location | Today | Purpose | Frequency | Suggested motion |
| --- | --- | --- | --- | --- | --- |
| 1 | `.gauge-mark` (`docs/index.html:186`) | The marker snaps between grades | State indication | Occasional, only while dragging the slider | `transition: background-color 200ms var(--ease-out)`. Position stays untransitioned, 1:1 with the slider |
| 2 | `.readout .angle` (`docs/index.html:203`) | The degree number changes colour instantly | State indication | Same | `transition: color 200ms var(--ease-out)` |
| 3 | `input[type=range]::-webkit-slider-thumb` (`docs/index.html:237`) | No press feedback | Feedback | Tens of times per visit, but only on the page's one control | `transform: scale(0.94)` on `:active`, `transition: transform 160ms var(--ease-out)` |
| 4 | `.bar a.plain` (`docs/index.html:100`) | Nav links change colour with no transition | Feedback | Occasional | `transition: color 120ms var(--ease-out)` |
| 5 | `a` (`docs/index.html:64`) | Underline thickens on hover with no transition | Feedback | Occasional | `transition: text-decoration-thickness 120ms var(--ease-out)` |

Two things every row has in common: the properties are colour, opacity or transform,
and each is under 300ms. None of them move anything the reader is reading.

## Part 2 — rejected, and why

- **The gauge marker's position.** Tempting: a spring on `translateX` would feel
  physical. **Rejected on function.** The marker's position *is* the slider's value.
  A transition there lags the thumb the reader is dragging, which is the one thing
  the page exists to demonstrate: direct, 1:1 response.
- **The specimen text resizing.** I wrote `transition: font-size 40ms linear` in the
  first draft and removed it. **Rejected on properties.** `font-size` triggers layout
  and paint on every frame of the drag, and the value is the input, so any duration
  makes the text trail the thumb. Removed entirely.
- **Scroll-triggered reveals on each section.** **Rejected on purpose.** Nothing in
  the list names what it would demonstrate. The page is a spec sheet, not a story.
- **A staggered entrance for the four feature pairs.** **Rejected on frequency and
  function.** The section is reached once per visit, so frequency allows it, but the
  reader is there to read four short paragraphs. Motion would delay the reading.
- **The nav bar's blur on scroll.** Already a material, not an animation. **Rejected
  as a candidate** rather than a suggestion: a blur radius animating on scroll would
  be a large moving surface, which the reduced-motion guidance warns against.
- **A spring on the outline pane, matching the app.** This is the landing page, not
  the app. There is no outline pane here to fold. **Rejected as out of scope.**

## Part 3 — verdict

This page needs almost no motion, and that is correct for what it is. A reader
arrives to check one claim, that type size follows viewing distance, and drags one
slider to test it. The only interaction that repeats is the slider, and it must stay
uncoupled from animation or the demonstration breaks.

The highest-leverage row is #1: the gauge marker's colour transition. It costs 200ms
and one line, and it turns a jump between "in band" and "outside band" into a
noticeable change rather than a flicker. Everything else is polish.

Reduced motion ships with all of it: the page sets `transition-duration: 0.01ms` under
`prefers-reduced-motion`, which keeps colour changes and removes movement. Nothing
here moves, but the rule is stated so a later addition inherits it.

`prefers-reduced-transparency` also ships: the sticky bar drops its backdrop blur and
becomes solid.
