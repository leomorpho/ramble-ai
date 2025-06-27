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
		client: {start:"_app/immutable/entry/start.C4SQ1KgK.js",app:"_app/immutable/entry/app.CCUsLWUG.js",imports:["_app/immutable/entry/start.C4SQ1KgK.js","_app/immutable/chunks/OZ36DSTA.js","_app/immutable/chunks/B3azi3pk.js","_app/immutable/chunks/SW6kwxMa.js","_app/immutable/entry/app.CCUsLWUG.js","_app/immutable/chunks/SW6kwxMa.js","_app/immutable/chunks/DA9ChAja.js","_app/immutable/chunks/B3p-91KC.js","_app/immutable/chunks/B3azi3pk.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
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
