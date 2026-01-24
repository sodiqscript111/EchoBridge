import { c as create_ssr_component } from "../../chunks/ssr.js";
import "../../chunks/auth.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  return `<div class="space-y-12"><header data-svelte-h="svelte-cyni1w"><h1 class="text-3xl font-bold text-white tracking-tight mb-2">Discover</h1> <p class="text-zinc-400">Explore public playlists from the community.</p></header> ${`<div class="text-zinc-500" data-svelte-h="svelte-1yugvm">Loading music...</div>`}</div>`;
});
export {
  Page as default
};
