import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark';

// Initialize theme from localStorage or system preference
function getInitialTheme(): Theme {
	if (!browser) return 'light';
	
	const stored = localStorage.getItem('theme') as Theme;
	if (stored) return stored;
	
	return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export const theme = writable<Theme>(getInitialTheme());

// Update localStorage and document class when theme changes
export function setTheme(newTheme: Theme) {
	if (!browser) return;
	
	theme.set(newTheme);
	localStorage.setItem('theme', newTheme);
	document.documentElement.classList.toggle('dark', newTheme === 'dark');
}

// Toggle between light and dark
export function toggleTheme() {
	theme.update(current => {
		const newTheme = current === 'light' ? 'dark' : 'light';
		setTheme(newTheme);
		return newTheme;
	});
}

// Initialize theme on page load
if (browser) {
	const currentTheme = getInitialTheme();
	setTheme(currentTheme);
}