<script>
  import { Search, Bell, User, Menu } from "lucide-svelte";
  import { isSidebarOpen } from "$lib/stores/ui";
</script>

<header class="sticky top-0 z-40 w-full glass">
  <div class="flex items-center justify-between px-8 py-4">
    <!-- Left: Breadcrumbs or Page Title (Placeholder) -->
    <div class="flex items-center gap-4">
      <button
        class="md:hidden text-zinc-400 hover:text-white"
        on:click={() => isSidebarOpen.update((v) => !v)}
      >
        <Menu class="w-6 h-6" />
      </button>

      <!-- Search Bar Removed -->
    </div>

    <!-- Right: User Actions -->
    <div class="flex items-center gap-4">
      <button class="text-zinc-400 hover:text-white transition-colors">
        <Bell class="w-5 h-5" />
      </button>
      <button
        class="w-8 h-8 rounded-full bg-zinc-700 flex items-center justify-center text-white font-bold text-xs ring-2 ring-black"
      >
        <User class="w-4 h-4" />
      </button>
    </div>
  </div>
</header>
