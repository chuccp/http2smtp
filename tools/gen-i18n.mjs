#!/usr/bin/env node
// Generates webui/i18n/*.json from the Vue locale files (view/src/locales/*.ts).
// The locale .ts files are plain `export default {...}` objects, which Node imports
// directly (type stripping) — no build step, no esbuild. Run from the repo root:
//
//   node tools/gen-i18n.mjs
//
// The JSON output is committed; the web server embeds it and never needs Node.

import { readFile, writeFile, mkdir } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const repoRoot = join(dirname(fileURLToPath(import.meta.url)), '..');
const localesDir = join(repoRoot, 'view', 'src', 'locales');
const outDir = join(repoRoot, 'webui', 'i18n');
const langs = ['zh-cn', 'zh-tw', 'en', 'ja'];

function flatten(obj, prefix, into) {
  for (const key of Object.keys(obj)) {
    const value = obj[key];
    const path = prefix ? `${prefix}.${key}` : key;
    if (value !== null && typeof value === 'object') {
      flatten(value, path, into);
    } else {
      into[path] = String(value);
    }
  }
}

// The key set every language must carry; a mismatch is an error, not a warning —
// the four files have to stay in key-for-key sync.
let reference = null;
let referenceLang = null;

await mkdir(outDir, { recursive: true });

for (const lang of langs) {
  const file = join(localesDir, `${lang}.ts`);
  if (!existsSync(file)) {
    console.error(`missing locale file: ${file}`);
    process.exit(1);
  }
  const mod = await import(pathToFileURL(file).href);
  const flat = {};
  flatten(mod.default, '', flat);

  const keys = Object.keys(flat).sort();
  if (reference === null) {
    reference = new Set(keys);
    referenceLang = lang;
  } else {
    const extra = keys.filter((k) => !reference.has(k));
    const absent = [...reference].filter((k) => !(k in flat));
    if (extra.length || absent.length) {
      console.error(`${lang}: keys out of sync with ${referenceLang}: extra=[${extra.join(', ')}] absent=[${absent.join(', ')}]`);
      process.exit(1);
    }
  }

  const outFile = join(outDir, `${lang}.json`);
  await writeFile(outFile, JSON.stringify(flat, null, 2) + '\n', 'utf8');
  console.log(`${lang}: ${keys.length} keys -> webui/i18n/${lang}.json`);
}
