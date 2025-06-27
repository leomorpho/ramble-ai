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
		client: {start:"_app/immutable/entry/start.bmG5DLWh.js",app:"_app/immutable/entry/app.CIwJUPuy.js",imports:["_app/immutable/entry/start.bmG5DLWh.js","_app/immutable/chunks/dRPt9nES.js","_app/immutable/chunks/DB07etuJ.js","_app/immutable/chunks/U--zayAP.js","_app/immutable/entry/app.CIwJUPuy.js","_app/immutable/chunks/U--zayAP.js","_app/immutable/chunks/DB07etuJ.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/Bxce8gUl.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js')),
			__memo(() => import('./nodes/3.js'))
		],
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/projects/[id]",
				pattern: /^\/projects\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
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
