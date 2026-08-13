import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Tauri serves the frontend as static files, so we use the static adapter
		// with an SPA fallback. See https://v2.tauri.app/start/frontend/sveltekit/
		adapter: adapter({ fallback: 'index.html' }),
		alias: {
			'@/': './src/lib',
			'@/*': './src/lib/*'
		}
	}
};

export default config;
