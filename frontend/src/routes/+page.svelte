<script>
  import { onMount } from "svelte";
  import PlaylistCard from "$lib/components/PlaylistCard.svelte";
  import { api } from "$lib/api";

  let categories = {};
  let loading = true;
  let error = null;

  onMount(async () => {
    try {
      const res = await api.getPublicPlaylists();
      const playlists = res.playlists || [];

      // Group by category
      categories = playlists.reduce((acc, playlist) => {
        const category = playlist.category || "Uncategorized";
        if (!acc[category]) {
          acc[category] = [];
        }
        acc[category].push(playlist);
        return acc;
      }, {});
    } catch (e) {
      console.error("Failed to load public playlists", e);
      error = "Failed to load playlists. Please try again.";
    } finally {
      loading = false;
    }
  });
</script>

<div class="space-y-12">
  <header>
    <h1 class="text-3xl font-bold text-white tracking-tight mb-2">Discover</h1>
    <p class="text-zinc-400">Explore public playlists from the community.</p>
  </header>

  {#if loading}
    <div class="text-zinc-500">Loading music...</div>
  {:else if error}
    <div class="text-red-500">{error}</div>
  {:else if Object.keys(categories).length === 0}
    <div class="text-zinc-500">No public playlists found.</div>
  {:else}
    {#each Object.entries(categories) as [category, playlists]}
      <section>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-bold text-white tracking-tight">
            {category}
          </h2>
          <a
            href="/category/{encodeURIComponent(category)}"
            class="text-sm font-medium text-apple-blue hover:text-blue-400 transition-colors"
          >
            See more
          </a>
        </div>

        <div
          class="flex overflow-x-auto snap-x snap-mandatory gap-6 pb-4 -mx-8 px-8 scrollbar-hide"
        >
          {#each playlists.slice(0, 6) as playlist}
            <div class="snap-start shrink-0 w-48">
              <PlaylistCard
                id={playlist.id}
                title={playlist.title}
                description={playlist.description || playlist.platform}
                image={playlist.cover_image ||
                  "https://images.unsplash.com/photo-1614680376593-902f74cf0d41?w=800&auto=format&fit=crop&q=60"}
              />
            </div>
          {/each}
        </div>
      </section>
    {/each}
  {/if}
</div>
