// The gauge's arithmetic, checked the way the page runs it: the marker must stay
// inside the 58px track at every slider value, and the band must match the
// comfortable range it claims to show.
const FROM = 0.18, TO = 0.46, BAND_LO = 0.26, BAND_HI = 0.38;
const PPI = 96, CM = 65, CAP = 0.70, RAD2DEG = 57.29577951308232;
const GAUGE = 58, MARK = 3;

function angleFor(px) {
  const heightCm = (px * CAP) / PPI * 2.54;
  return 2 * Math.atan(heightCm / 2 / CM) * RAD2DEG;
}

let worst = null;
for (let px = 14; px <= 30; px += 0.5) {
  const a = angleFor(px);
  let t = (a - FROM) / (TO - FROM);
  t = Math.max(0, Math.min(1, t));
  const x = t * (GAUGE - MARK);
  const right = x + MARK;
  if (x < -0.01 || right > GAUGE + 0.01) {
    console.log(`OUT OF BOUNDS at ${px}pt: left=${x.toFixed(1)} right=${right.toFixed(1)}`);
  }
  if (!worst || x < worst.x) worst = { px, x };
}

// The band's own placement must match the degrees it claims.
const lo = (BAND_LO - FROM) / (TO - FROM);
const hi = (BAND_HI - FROM) / (TO - FROM);
console.log(`band: ${(lo * 100).toFixed(1)}% .. ${(hi * 100).toFixed(1)}%`);
console.log(`css says left/right 28.6% -> band ${(100 - 28.6).toFixed(1)}% wide`);

// Sanity: 18pt at 96ppi/65cm should land near the target 0.32.
console.log(`18pt -> ${angleFor(18).toFixed(3)} deg`);
console.log(`14pt -> ${angleFor(14).toFixed(3)} deg (below band)`);
console.log(`30pt -> ${angleFor(30).toFixed(3)} deg (above band)`);
