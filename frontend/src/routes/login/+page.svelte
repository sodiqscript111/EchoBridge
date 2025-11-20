<script>
  import { goto } from '$app/navigation';
  import { api } from '$lib/api';
  import { auth } from '$lib/stores/auth';
  import { LogIn } from 'lucide-svelte';

  let email = '';
  let password = '';
  let loading = false;
  let error = '';

  async function handleLogin() {
    error = '';
    loading = true;

    try {
      const response = await api.post('/login/email', {
        email,
        password,
      });

      auth.login(response.jwt_token, response.user);
      goto('/dashboard');
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
        <h2 class="text-3xl font-bold mb-2">Welcome Back</h2>
        <p class="text-white/60">Log in to continue syncing</p>
      </div>

      {#if error}
        <div class="bg-red-500/10 border border-red-500/50 rounded-2xl p-4 mb-6 text-red-400 text-sm">
          {error}
        </div>
      {/if}

      <form on:submit|preventDefault={handleLogin} class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-2 text-white/70">Email</label>
          <input
            type="email"
            bind:value={email}
            required
            class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30"
            placeholder="you@example.com"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2 text-white/70">Password</label>
          <input
            type="password"
            bind:value={password}
            required
            class="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-xl focus:outline-none focus:border-apple-blue transition-colors text-white placeholder-white/30"
            placeholder="••••••••"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          class="w-full py-3 bg-apple-blue rounded-xl font-bold hover:bg-blue-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {#if loading}
            <div class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
            Logging in...
          {:else}
            <LogIn size={20} />
            Log In
          {/if}
        </button>
      </form>

      <div class="mt-6 text-center text-sm text-white/60">
        Don't have an account?
        <a href="/signup" class="text-apple-blue hover:underline font-medium">Sign up</a>
      </div>
    </div>
  </div>
</div>
