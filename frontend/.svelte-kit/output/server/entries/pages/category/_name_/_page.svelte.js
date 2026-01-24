import { c as create_ssr_component, b as escape } from "../../../../chunks/ssr.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/state.svelte.js";
import "../../../../chunks/auth.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let { data } = $$props;
  let categoryName = decodeURIComponent(data.category);
  if ($$props.data === void 0 && $$bindings.data && data !== void 0) $$bindings.data(data);
  return `<div class="space-y-8"><header><div class="flex items-center gap-2 mb-2 text-sm text-zinc-400"><a href="/" class="hover:text-white transition-colors" data-svelte-h="svelte-9u9mv4">Home</a> <span data-svelte-h="svelte-ocknwt">/</span> <span class="text-white">${escape(categoryName)}</span></div> <h1 class="text-3xl font-bold text-white tracking-tight">${escape(categoryName)}</h1> <p class="text-zinc-400" data-svelte-h="svelte-15wcw02">All public playlists in this category.</p></header> ${`<div class="text-zinc-500" data-svelte-h="svelte-16o4mal">Loading...</div>`}</div>`;
});
export {
  Page as default
};
