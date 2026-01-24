import * as universal from '../entries/pages/category/_name_/_page.js';

export const index = 3;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/category/_name_/_page.svelte.js')).default;
export { universal };
export const universal_id = "src/routes/category/[name]/+page.js";
export const imports = ["_app/immutable/nodes/3.D5OB1tAs.js","_app/immutable/chunks/DF-Dz73P.js","_app/immutable/chunks/BdLCvuL5.js","_app/immutable/chunks/BeuEQZSN.js","_app/immutable/chunks/Cq7TlG0e.js","_app/immutable/chunks/DS_D3Oba.js","_app/immutable/chunks/B5nKSrpE.js","_app/immutable/chunks/CEpfNtLG.js","_app/immutable/chunks/DIVqPN1U.js","_app/immutable/chunks/DOzjk_c7.js"];
export const stylesheets = [];
export const fonts = [];
