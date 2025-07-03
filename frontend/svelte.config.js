import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
export default {
  kit: {
		adapter: adapter({
      // Static needs a fallback page.
      fallback: 'index.html'
    }),
    alias: {
      "@/*": "./path/to/lib/*",
    },
	},
  compilerOptions: {
    // Disable accessibility warnings for now
    warningFilter: (warning) => {
      // Filter out a11y warnings
      if (warning.code.startsWith('a11y-')) return false;
      return true;
    }
  }
};
