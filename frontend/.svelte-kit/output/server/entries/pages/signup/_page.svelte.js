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
  let username = "";
  let password = "";
  return `<div class="flex items-center justify-center min-h-[80vh] px-4"><div class="w-full max-w-md"><div class="glass rounded-3xl p-8"><div class="text-center mb-8" data-svelte-h="svelte-1q4w5b"><h2 class="text-3xl font-bold mb-2">Create Account</h2> <p class="text-white/60">Join EchoBridge and sync your music</p></div> ${``} <form class="space-y-4"><div><label class="block text-sm font-medium mb-2 text-white/70" data-svelte-h="svelte-n1d9vo">Username</label> <input type="text" required class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30" placeholder="johndoe"${add_attribute("value", username, 0)}></div> <div><label class="block text-sm font-medium mb-2 text-white/70" data-svelte-h="svelte-pqa7z4">Email</label> <input type="email" required class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30" placeholder="you@example.com"${add_attribute("value", email, 0)}></div> <div><label class="block text-sm font-medium mb-2 text-white/70" data-svelte-h="svelte-1rs24bv">Password</label> <input type="password" required minlength="8" class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30" placeholder="••••••••"${add_attribute("value", password, 0)}> <p class="text-xs text-white/40 mt-1" data-svelte-h="svelte-1j80x3">Minimum 8 characters</p></div> <button type="submit" ${""} class="w-full py-3 bg-apple-blue rounded-xl font-bold hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2">${`${validate_component(Log_in, "LogIn").$$render($$result, { size: 20 }, {}, {})}
            Sign Up`}</button></form> <div class="mt-6 text-center text-sm text-white/60" data-svelte-h="svelte-1u0sj36">Already have an account?
        <a href="/login" class="text-apple-blue hover:underline font-medium">Log in</a></div></div></div></div>`;
});
export {
  Page as default
};
