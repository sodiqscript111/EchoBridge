import { c as create_ssr_component, d as add_attribute, v as validate_component } from "../../../chunks/ssr.js";
import "@sveltejs/kit/internal";
import "../../../chunks/exports.js";
import "../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../chunks/state.svelte.js";
import "../../../chunks/auth.js";
import { L as Log_in } from "../../../chunks/log-in.js";
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let email = "";
  let password = "";
  return `<div class="flex items-center justify-center min-h-[80vh] px-4"><div class="w-full max-w-md"><div class="glass rounded-3xl p-8"><div class="text-center mb-8" data-svelte-h="svelte-4cr1pp"><h2 class="text-3xl font-bold mb-2">Welcome Back</h2> <p class="text-white/60">Log in to continue syncing</p></div> ${``} <form class="space-y-4"><div><label class="block text-sm font-medium mb-2 text-white/70" data-svelte-h="svelte-kv3ka2">Email</label> <input type="email" required class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30" placeholder="you@example.com"${add_attribute("value", email, 0)}></div> <div><label class="block text-sm font-medium mb-2 text-white/70" data-svelte-h="svelte-1euysjn">Password</label> <input type="password" required class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30" placeholder="••••••••"${add_attribute("value", password, 0)}></div> <button type="submit" ${""} class="w-full py-3 bg-apple-blue rounded-xl font-bold hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2">${`${validate_component(Log_in, "LogIn").$$render($$result, { size: 20 }, {}, {})}
            Log In`}</button></form> <div class="mt-6 text-center text-sm text-white/60" data-svelte-h="svelte-5hro96">Don&#39;t have an account?
        <a href="/signup" class="text-apple-blue hover:underline font-medium">Sign up</a></div></div></div></div>`;
});
export {
  Page as default
};
