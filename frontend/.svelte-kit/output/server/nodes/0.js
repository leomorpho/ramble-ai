

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/fallbacks/layout.svelte.js')).default;
export const imports = ["_app/immutable/nodes/0.CaLlLyVE.js","_app/immutable/chunks/B3p-91KC.js","_app/immutable/chunks/SW6kwxMa.js"];
export const stylesheets = [];
export const fonts = [];
