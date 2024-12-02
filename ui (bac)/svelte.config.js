import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { phosphorSvelteOptimize } from 'phosphor-svelte/preprocessor';
import preprocess from 'svelte-preprocess';

import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: [
		phosphorSvelteOptimize(),
		vitePreprocess(),
		preprocess({
			postcss: {
				configFilePath: join(__dirname, 'postcss.config.cjs')
			},
			markupTagName: 'markup'
		})
	],
	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: undefined,
			precompress: false,
			strict: true
		}),
		alias: {
			$components: 'src/lib/components',
			'$components/*': 'src/lib/components/*'
		}
	}
};

export default config;
