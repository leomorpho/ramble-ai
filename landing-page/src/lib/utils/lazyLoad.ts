/**
 * Enhanced lazy loading utility with intersection observer fallback
 * Provides progressive enhancement for older browsers
 */

interface LazyLoadOptions {
	rootMargin?: string;
	threshold?: number;
	fallbackDelay?: number;
}

export function lazyLoad(
	node: HTMLImageElement,
	options: LazyLoadOptions = {}
) {
	const {
		rootMargin = '50px',
		threshold = 0.1,
		fallbackDelay = 100
	} = options;

	// Check if browser supports Intersection Observer
	if ('IntersectionObserver' in window) {
		const observer = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						loadImage(entry.target as HTMLImageElement);
						observer.unobserve(entry.target);
					}
				});
			},
			{
				rootMargin,
				threshold
			}
		);

		observer.observe(node);

		return {
			destroy() {
				observer.unobserve(node);
			}
		};
	} else {
		// Fallback for older browsers - load after a delay
		const timeoutId = setTimeout(() => {
			loadImage(node);
		}, fallbackDelay);

		return {
			destroy() {
				clearTimeout(timeoutId);
			}
		};
	}
}

function loadImage(img: HTMLImageElement) {
	// Get the actual src from data-src attribute if present
	const src = img.dataset.src || img.src;
	
	if (src && img.src !== src) {
		img.src = src;
		
		// Add loaded class for animations
		img.addEventListener('load', () => {
			img.classList.add('lazy-loaded');
		}, { once: true });
		
		// Handle loading errors
		img.addEventListener('error', () => {
			img.classList.add('lazy-error');
			console.warn('Failed to load image:', src);
		}, { once: true });
	}
}

// CSS classes for styling loading states
export const lazyLoadCSS = `
	.lazy-loading {
		opacity: 0;
		transition: opacity 0.3s ease-in-out;
	}
	
	.lazy-loaded {
		opacity: 1;
	}
	
	.lazy-error {
		opacity: 0.5;
		filter: grayscale(1);
	}
`;