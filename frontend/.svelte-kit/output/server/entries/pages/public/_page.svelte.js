import { c as create_ssr_component } from "../../../chunks/ssr.js";
import "../../../chunks/auth.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  return `<div class="p-8 pt-24 max-w-7xl mx-auto pb-24"><div class="mb-16 text-center" data-svelte-h="svelte-up1if6"><h1 class="text-6xl font-black mb-6 tracking-tight text-white">Public Mixes</h1> <p class="text-xl text-white/60 font-medium max-w-2xl mx-auto leading-relaxed">Discover what the world is listening to, curated by AI and our community.</p></div> ${`<div class="flex items-center justify-center py-32" data-svelte-h="svelte-eao3vd"><div class="w-16 h-16 border-4 border-white/10 border-t-apple-blue rounded-full animate-spin"></div></div>`}</div>`;
});
export {
  Page as default
};
