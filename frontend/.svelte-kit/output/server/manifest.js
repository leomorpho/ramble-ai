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
		client: {start:"_app/immutable/entry/start.D9LPwsPz.js",app:"_app/immutable/entry/app.BacEljVf.js",imports:["_app/immutable/entry/start.D9LPwsPz.js","_app/immutable/chunks/BG208bVx.js","_app/immutable/chunks/C7DE31cZ.js","_app/immutable/chunks/DRR-r2a_.js","_app/immutable/entry/app.BacEljVf.js","_app/immutable/chunks/DRR-r2a_.js","_app/immutable/chunks/C7DE31cZ.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/Co015kgV.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
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
