

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/dashboard/_page.svelte.js')).default;
export const universal = {
  "ssr": false
};
export const universal_id = "src/routes/dashboard/+page.js";
export const imports = ["_app/immutable/nodes/4.AB_QUz0X.js","_app/immutable/chunks/DF-Dz73P.js","_app/immutable/chunks/BdLCvuL5.js","_app/immutable/chunks/BeuEQZSN.js","_app/immutable/chunks/DOzjk_c7.js","_app/immutable/chunks/DS_D3Oba.js","_app/immutable/chunks/BNgdRHux.js","_app/immutable/chunks/Cq7TlG0e.js","_app/immutable/chunks/mF8Y1LZz.js","_app/immutable/chunks/ju0NKtQh.js","_app/immutable/chunks/BzGWfxMs.js","_app/immutable/chunks/DIVqPN1U.js"];
export const stylesheets = ["_app/immutable/assets/PlaylistImportModal.BgR5RrPa.css","_app/immutable/assets/4.CyEIygQM.css"];
export const fonts = [];
