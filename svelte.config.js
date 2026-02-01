import adapter from '@sveltejs/adapter-static';

 
const config = {
	kit: {
		 
		 
		 
		adapter: adapter({
			fallback: 'index.html'  
		}),
		files: {
			lib: 'frontend/lib',
			routes: 'frontend/routes',
			appTemplate: 'frontend/app.html'
		}
	}
};

export default config;
