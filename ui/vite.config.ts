import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		hmr: {
			// In dev mode, this forces the socket to use a port other than the default client
			// port, which will be that of the golang backend. It allows HMR to by pass the
			// backend and connect directly to the vite server
			port: 5174
		}
	}
});
