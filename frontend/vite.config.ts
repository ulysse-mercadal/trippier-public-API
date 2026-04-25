import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	// Read from the repo-root .env so all services share a single env file
	envDir: '..',
	server: {
		port: 3000,
	},
});
