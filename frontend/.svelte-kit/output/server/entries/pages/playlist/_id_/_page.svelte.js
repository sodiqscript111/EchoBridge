import { c as create_ssr_component, a as subscribe } from "../../../../chunks/ssr.js";
import { p as page } from "../../../../chunks/stores.js";
import { a as auth } from "../../../../chunks/auth.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $$unsubscribe_auth;
  let $$unsubscribe_page;
  $$unsubscribe_auth = subscribe(auth, (value) => value);
  $$unsubscribe_page = subscribe(page, (value) => value);
  $$unsubscribe_auth();
  $$unsubscribe_page();
  return `<div class="space-y-8 pb-20 relative">${`<div class="text-zinc-500" data-svelte-h="svelte-knwo9p">Loading playlist...</div>`}  ${``}  ${``}</div>`;
});
export {
  Page as default
};
