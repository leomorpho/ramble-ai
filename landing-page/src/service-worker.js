// Check if we're in development mode
const dev = !globalThis.location?.protocol.startsWith('https');

// Initialize variables
let build, files, version;
let CACHE = 'cache-dev';
let ASSETS = [];
let initialized = false;

// Initialize service worker modules asynchronously
async function initializeServiceWorker() {
	if (initialized || dev) return;
	
	try {
		const sw = await import('$service-worker');
		build = sw.build;
		files = sw.files;
		version = sw.version;
		CACHE = `cache-${version}`;
		ASSETS = [...build, ...files];
		initialized = true;
	} catch (e) {
		console.warn('Service worker modules not available:', e);
		initialized = true; // Mark as initialized even on error to prevent retries
	}
}

self.addEventListener('install', (event) => {
	// Create a new cache and add all files to it
	async function addFilesToCache() {
		await initializeServiceWorker();
		const cache = await caches.open(CACHE);
		if (ASSETS.length > 0) {
			await cache.addAll(ASSETS);
		}
	}

	event.waitUntil(addFilesToCache());
});

self.addEventListener('activate', (event) => {
	// Remove previous cached data from disk
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	event.waitUntil(deleteOldCaches());
});

self.addEventListener('fetch', (event) => {
	// ignore POST requests etc
	if (event.request.method !== 'GET') return;

	async function respond() {
		await initializeServiceWorker();
		const url = new URL(event.request.url);
		const cache = await caches.open(CACHE);

		// In development, just pass through to network
		if (dev) {
			return fetch(event.request);
		}

		// `build`/`files` can always be served from the cache
		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(url.pathname);

			if (response) {
				return response;
			}
		}

		// for everything else, try the network first, but
		// fall back to the cache if we're offline
		try {
			const response = await fetch(event.request);

			// if we're offline, fetch can return a value that is not a Response
			// instead of throwing - and we can't pass this non-Response to respondWith
			if (!(response instanceof Response)) {
				throw new Error('invalid response from fetch');
			}

			if (response.status === 200) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch (err) {
			const response = await cache.match(event.request);

			if (response) {
				return response;
			}

			// if there's no cache, then just error out
			// as there is nothing we can do to respond to this request
			throw err;
		}
	}

	event.respondWith(respond());
});