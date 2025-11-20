<script>
  import { Home, Globe, User, LogOut } from "lucide-svelte";
  import { auth } from "$lib/stores/auth";
  import { goto } from "$app/navigation";

  function handleLogout() {
    auth.logout();
    goto("/");
  }
</script>

<nav class="fixed top-0 left-0 right-0 z-50 px-6 py-4">
  <div class="max-w-7xl mx-auto">
    <div class="glass rounded-full px-6 py-3 flex justify-between items-center">
      <!-- Logo -->
      <a
        href="/"
        class="text-xl font-bold tracking-tighter flex items-center gap-2"
      >
        <div class="w-3 h-3 rounded-full bg-apple-blue animate-pulse"></div>
        EchoBridge
      </a>

      <!-- Right Side: Links & Auth -->
      <div class="flex items-center gap-8">
        <!-- Links -->
        {#if $auth.token}
          <a
            href="/dashboard"
            class="text-sm font-medium text-white/70 hover:text-white transition-colors flex items-center gap-2"
          >
            <Home size={16} /> Dashboard
          </a>
        {/if}
        <a
          href="/public"
          class="text-sm font-medium text-white/70 hover:text-white transition-colors flex items-center gap-2"
        >
          <Globe size={16} /> Public
        </a>

        <!-- Divider -->
        <div class="h-4 w-px bg-white/10"></div>

        <!-- Profile / Auth -->
        {#if $auth.token}
          <button
            on:click={handleLogout}
            class="flex items-center gap-2 px-4 py-2 rounded-full bg-white/10 hover:bg-white/20 transition-colors text-sm font-medium"
          >
            <User size={16} />
            {$auth.user?.username || "User"}
            <LogOut size={14} class="ml-2" />
          </button>
        {:else}
          <div class="flex items-center gap-3">
            <a
              href="/login"
              class="text-sm font-medium text-white/70 hover:text-white transition-colors"
            >
              Log In
            </a>
            <a
              href="/signup"
              class="px-4 py-2 bg-apple-blue rounded-full text-sm font-bold hover:bg-blue-600 transition-colors shadow-lg shadow-blue-500/20"
            >
              Sign Up
            </a>
          </div>
        {/if}
      </div>
    </div>
  </div>
</nav>
