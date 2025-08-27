// Application configuration
export const config = {
	app: {
		name: 'My SvelteKit App',
		description: 'A modern full-stack application built with SvelteKit and PocketBase'
	},
	
	// Get current year dynamically
	getCurrentYear: () => new Date().getFullYear()
} as const;