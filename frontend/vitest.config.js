import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { resolve } from 'path';

export default defineConfig({
  plugins: [svelte({ hot: !process.env.VITEST })],
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.js'],
    globals: true,
    include: ['src/**/*.{test,spec}.{js,ts}']
  },
  resolve: {
    alias: {
      '$lib': resolve(__dirname, './src/lib'),
      '$lib/wailsjs/go/main/App': resolve(__dirname, './src/test/mocks/wailsjs.js'),
      'svelte-sonner': resolve(__dirname, './src/test/mocks/svelte-sonner.js')
    }
  }
});