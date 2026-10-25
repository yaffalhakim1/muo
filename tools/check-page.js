// Checks docs/ for the things a broken landing page gets wrong: dead local
// links, missing images, duplicate ids, and the design system's hard rules.
const fs = require('fs');
const path = require('path');

const pages = ['index.html', 'privacy.html', 'terms.html', '404.html'];
let problems = 0;

function fail(msg) {
  console.log('  FAIL ' + msg);
  problems++;
}

for (const page of pages) {
  const file = path.join('docs', page);
  if (!fs.existsSync(file)) { fail(page + ' missing'); continue; }
  const h = fs.readFileSync(file, 'utf8');
  console.log(page + ':');

  // Local links must resolve, unless they are in-page anchors or absolute.
  const links = [...h.matchAll(/href="([^"]+)"/g)].map(m => m[1]);
  for (const link of links) {
    if (link.startsWith('#') || link.startsWith('http') || link.startsWith('mailto:')) continue;
    const target = link.split('#')[0];
    if (!target) continue;
    const resolved = target.startsWith('/muo/')
      ? path.join('docs', target.replace('/muo/', ''))
      : path.join('docs', target);
    if (!fs.existsSync(resolved)) fail('dead link ' + link);
  }

  // Images must exist, and every one needs alt text.
  const imgs = [...h.matchAll(/<img\b([^>]*)>/g)];
  for (const img of imgs) {
    const attrs = img[1];
    const src = (attrs.match(/src="([^"]+)"/) || [])[1];
    if (!src) { fail('img without src'); continue; }
    if (!attrs.includes('alt=')) fail('img without alt: ' + src);
    const resolved = path.join('docs', src);
    if (!src.startsWith('http') && !fs.existsSync(resolved)) fail('missing image ' + src);
  }

  // Ids must be unique, and every getElementById must find one.
  const ids = [...h.matchAll(/id="([^"]+)"/g)].map(m => m[1]);
  const dup = ids.filter((v, i) => ids.indexOf(v) !== i);
  if (dup.length) fail('duplicate ids: ' + dup.join(', '));
  const refs = [...h.matchAll(/getElementById\('([^']+)'\)/g)].map(m => m[1]);
  for (const r of refs) if (!ids.includes(r)) fail('getElementById finds nothing: ' + r);

  // The design system's hard rules.
  if (/font-family[^;]*(Inter|Roboto|Arial|"Open Sans"|Helvetica)/.test(h)) fail('banned typeface');
  if (/font-style:\s*italic/.test(h)) fail('italic type');
  const weights = [...h.matchAll(/font-weight:\s*(\d{3})/g)].map(m => Number(m[1]));
  if (weights.some(w => w > 600)) fail('weight above 600');
  if (/transition-all/.test(h)) fail('transition-all');
  if (/scale\(0\)/.test(h)) fail('scale(0) entrance');
  if (/(^|[^-])ease-in\b/.test(h)) fail('ease-in');

  // Gradients are allowed on the hero heading only.
  const grads = [...h.matchAll(/(linear|radial|conic)-gradient/g)];
  if (grads.length > 1) fail(grads.length + ' gradients, only the hero heading may have one');

  console.log('  links ' + links.length + ', images ' + imgs.length + ', ids ' + ids.length);
}

// Pages the ship checklist requires.
for (const required of ['404.html', 'privacy.html', 'terms.html', 'icon.png', 'og.png']) {
  if (!fs.existsSync(path.join('docs', required))) fail('missing ' + required);
}

console.log(problems === 0 ? '\nall checks passed' : '\n' + problems + ' problems');
process.exit(problems === 0 ? 0 : 1);
