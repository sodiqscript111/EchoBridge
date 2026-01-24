<script>
  import { goto } from "$app/navigation";
  import { api } from "$lib/api";
  import { auth } from "$lib/stores/auth";
  import { LogIn } from "lucide-svelte";

  let email = "";
  let username = "";
  let password = "";
  let loading = false;
  let error = "";

  async function handleSignup() {
    error = "";
    loading = true;

    try {
      const response = await api.post("/register", {
        email,
        username,
        password,
      });

      auth.login(response.jwt_token, response.user);
      goto("/dashboard");
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="flex items-center justify-center min-h-[80vh] px-4">
  <div class="w-full max-w-md">
    <div class="glass rounded-3xl p-8">
      <div class="text-center mb-8">
        <h2 class="text-3xl font-bold mb-2">Create Account</h2>
        <p class="text-white/60">Join EchoBridge and sync your music</p>
      </div>

      {#if error}
        <div
          class="bg-red-500/10 border border-red-500/50 rounded-2xl p-4 mb-6 text-red-400 text-sm"
        >
          {error}
        </div>
      {/if}

      <form on:submit|preventDefault={handleSignup} class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-2 text-white/70"
            >Username</label
          >
          <input
            type="text"
            bind:value={username}
            required
            class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30"
            placeholder="johndoe"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2 text-white/70"
            >Email</label
          >
          <input
            type="email"
            bind:value={email}
            required
            class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30"
            placeholder="you@example.com"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2 text-white/70"
            >Password</label
          >
          <input
            type="password"
            bind:value={password}
            required
            minlength="8"
            class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30"
            placeholder="••••••••"
          />
          <p class="text-xs text-white/40 mt-1">Minimum 8 characters</p>
        </div>

        <button
          type="submit"
          disabled={loading}
          class="w-full py-3 bg-apple-blue rounded-xl font-bold hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {#if loading}
            <div
              class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
            ></div>
            Creating Account...
          {:else}
            <LogIn size={20} />
            Sign Up
          {/if}
        </button>
      </form>

      <div class="mt-6 text-center text-sm text-white/60">
        Already have an account?
        <a href="/login" class="text-apple-blue hover:underline font-medium"
          >Log in</a
        >
      </div>
    </div>
  </div>
</div>
