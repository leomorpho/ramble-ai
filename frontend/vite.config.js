import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
// import path from 'path'

/** @type {import('vite').UserConfig} */
const config = {
  server: {
    fs: {
      // To allow serving files from the frontend project root.
      //
      // allow: ['.'],
    },
    hmr: {
      // Better hot-reload
      overlay: true,
    },
    watch: {
      // Poll for file changes (useful for some systems)
      usePolling: false,
      interval: 100,
    },
  },
	plugins: [sveltekit(), tailwindcss()],
  resolve: {
    alias: {
      // This alias finishes the ability to reference our
      // frontend dirctory with "@path/to/file."
      // You also need to add the path to jsconfig.json.
      //
      // '@': path.resolve(__dirname, './'), 
    },
  },
};

export default config;

