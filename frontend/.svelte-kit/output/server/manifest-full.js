export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.png"]),
	mimeTypes: {".png":"image/png"},
	_: {
		client: {start:"_app/immutable/entry/start.B-rAsPDi.js",app:"_app/immutable/entry/app.BAsEc0xQ.js",imports:["_app/immutable/entry/start.B-rAsPDi.js","_app/immutable/chunks/CF7KS7oH.js","_app/immutable/chunks/DJY-c6uu.js","_app/immutable/chunks/CdXqVUbr.js","_app/immutable/chunks/C8TegXoY.js","_app/immutable/entry/app.BAsEc0xQ.js","_app/immutable/chunks/CdXqVUbr.js","_app/immutable/chunks/C8TegXoY.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/DJY-c6uu.js","_app/immutable/chunks/B2OhTg8B.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js'))
		],
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
