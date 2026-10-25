const fs = require('fs');
const h = fs.readFileSync('docs/index.html', 'utf8');

const ids = [...h.matchAll(/id="([^"]+)"/g)].map(m => m[1]);
const dup = ids.filter((v, i) => ids.indexOf(v) !== i);
console.log('ids:', ids.length, '| duplicates:', dup.length ? dup.join(',') : 'none');

const refs = [...h.matchAll(/getElementById\('([^']+)'\)/g)].map(m => m[1]);
const missing = refs.filter(r => !ids.includes(r));
console.log('js refs:', refs.length, '| missing:', missing.length ? missing.join(',') : 'none');

const imgs = [...h.matchAll(/src="([^"]+\.png)"/g)].map(m => m[1]);
let allThere = true;
for (const i of imgs) {
  if (!fs.existsSync('docs/' + i)) { console.log('MISSING IMAGE:', i); allThere = false; }
}
console.log('images:', imgs.length, '| all present:', allThere);

// Any colour literal that should be a token.
const hex = [...h.matchAll(/#[0-9a-fA-F]{3,8}\b/g)].map(m => m[0]);
console.log('hardcoded hex:', hex.length ? [...new Set(hex)].join(' ') : 'none');

// Animation hygiene.
const bad = [];
if (/transition:\s*all/.test(h)) bad.push('transition: all');
if (/ease-in[^-]/.test(h)) bad.push('ease-in');
if (/scale\(0\)/.test(h)) bad.push('scale(0)');
console.log('motion violations:', bad.length ? bad.join(', ') : 'none');
