import { c as create_ssr_component, a as subscribe, v as validate_component } from "../../../chunks/ssr.js";
import { a as auth } from "../../../chunks/auth.js";
/* empty css                                                                */
import { P as Plus } from "../../../chunks/plus.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $$unsubscribe_auth;
  $$unsubscribe_auth = subscribe(auth, (value) => value);
  $$unsubscribe_auth();
  return `<div class="space-y-8"><header class="flex items-center justify-between"><div data-svelte-h="svelte-1qi5p6h"><h1 class="text-3xl font-bold text-white tracking-tight mb-2">Your Library</h1> <p class="text-zinc-400">All your imported playlists from Spotify and YouTube.</p></div> <button class="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-full font-bold hover:scale-105 transition-transform">${validate_component(Plus, "Plus").$$render($$result, { size: 20 }, {}, {})}
            Import Playlist</button></header> ${`<div class="text-zinc-500" data-svelte-h="svelte-1mte21q">Loading library...</div>`}</div> ${``}`;
});
export {
  Page as default
};
