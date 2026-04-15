// @ts-check
import { defineConfig } from 'astro/config';
import { readFileSync } from 'node:fs';

const glitchGrammar = JSON.parse(
  readFileSync(new URL('./src/glitch.tmLanguage.json', import.meta.url), 'utf-8')
);

// https://astro.build/config
export default defineConfig({
  site: 'https://8op.org',
  markdown: {
    shikiConfig: {
      theme: 'tokyo-night',
      langs: [
        {
          ...glitchGrammar,
          name: 'glitch',
        },
      ],
    },
  },
});
