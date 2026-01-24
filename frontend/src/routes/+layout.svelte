<script>
  import "../app.css";
  import Sidebar from "$lib/components/Sidebar.svelte";
  import Navbar from "$lib/components/Navbar.svelte";
  import { isSidebarOpen } from "$lib/stores/ui";
  import { page } from "$app/stores";

  // Close sidebar on page navigation
  $: if ($page.url.pathname) {
    isSidebarOpen.set(false);
  }
</script>

<div
  class="min-h-screen bg-black text-zinc-400 selection:bg-white selection:text-black flex"
>
  <Sidebar />

  <!-- Mobile Overlay -->
  {#if $isSidebarOpen}
    <div
      class="fixed inset-0 bg-black/80 z-40 md:hidden backdrop-blur-sm"
      on:click={() => isSidebarOpen.set(false)}
      role="button"
      tabindex="0"
      on:keydown={(e) => e.key === "Escape" && isSidebarOpen.set(false)}
    ></div>
  {/if}

  <main
    class="flex-1 md:ml-64 min-h-screen flex flex-col w-full transition-all duration-300"
  >
    <Navbar />
    <div class="flex-1 p-4 md:p-8">
      <slot />
    </div>
  </main>
</div>
