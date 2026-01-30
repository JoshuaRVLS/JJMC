import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		// adapter-static only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
		// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
		// See https://svelte.dev/docs/kit/adapters for more information about adapters.
		adapter: adapter({
			fallback: 'index.html' // Enable SPA mode
		}),
		files: {
			hooks: {
				client: 'frontend/hooks.client',
				server: 'frontend/hooks.server'
			},
			lib: 'frontend/lib',
			params: 'frontend/params',
			routes: 'frontend/routes',
			serviceWorker: 'frontend/service-worker',
			appTemplate: 'frontend/app.html',
			errorTemplate: 'frontend/error.html'
		}
	}
};

export default config;
