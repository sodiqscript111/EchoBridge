<script>
  import { onMount } from "svelte";
  import GradientCard from "$lib/components/GradientCard.svelte";
  import { api } from "$lib/api";

  let playlists = [];
  let groupedPlaylists = {};
  let loading = true;
  let categories = [];

  const colorPairs = [
    { c1: "#4f46e5", c2: "#ec4899" },
    { c1: "#ef4444", c2: "#f59e0b" },
    { c1: "#10b981", c2: "#3b82f6" },
    { c1: "#8b5cf6", c2: "#ec4899" },
    { c1: "#06b6d4", c2: "#8b5cf6" },
  ];

  onMount(async () => {
    try {
      const response = await api.get("/playlists/public");
      playlists = (response.playlists || []).map((p, i) => ({
        ...p,
        colors: colorPairs[i % colorPairs.length],
        artistsPreview:
          p.tracks
            ?.slice(0, 3)
            .map((t) => t.artist)
            .join(", ") || "Various Artists",
        songCount: p.tracks?.length || 0,
        category: p.category || "Uncategorized",
      }));

      // Group by category
      groupedPlaylists = playlists.reduce((acc, p) => {
        const cat = p.category;
        if (!acc[cat]) acc[cat] = [];
        acc[cat].push(p);
        return acc;
      }, {});

      // Sort categories (Uncategorized last)
      categories = Object.keys(groupedPlaylists).sort((a, b) => {
        if (a === "Uncategorized") return 1;
        if (b === "Uncategorized") return -1;
        return a.localeCompare(b);
      });
    } catch (e) {
      console.error("Failed to fetch:", e);
    } finally {
      loading = false;
    }
  });
</script>

<div class="p-8 pt-24 max-w-7xl mx-auto pb-24">
  <div class="mb-16 text-center">
    <h1 class="text-6xl font-black mb-6 tracking-tight text-white">
      Public Mixes
    </h1>
    <p
      class="text-xl text-white/60 font-medium max-w-2xl mx-auto leading-relaxed"
    >
      Discover what the world is listening to, curated by AI and our community.
    </p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-32">
      <div
        class="w-16 h-16 border-4 border-white/10 border-t-apple-blue rounded-full animate-spin"
      ></div>
    </div>
  {:else if playlists.length === 0}
    <div class="glass rounded-3xl p-16 text-center max-w-2xl mx-auto">
      <p class="text-2xl text-white/60 font-medium">No public playlists yet.</p>
      <p class="text-white/40 mt-2">
        Be the first to share your vibe with the world!
      </p>
    </div>
  {:else}
    <div class="space-y-16">
      {#each categories as category}
        <section>
          <div class="flex items-center gap-4 mb-8">
            <h2 class="text-3xl font-bold tracking-tight text-white">
              {category}
            </h2>
            <div class="h-px flex-1 bg-white/10"></div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {#each groupedPlaylists[category] as p}
              <a href="/playlist/{p.id}" class="block">
                <GradientCard
                  title={p.title}
                  songCount={p.songCount}
                  artists={p.artistsPreview}
                  coverImage={p.cover_image}
                  category={p.category === "Uncategorized" ? "" : p.category}
                  color1={p.colors.c1}
                  color2={p.colors.c2}
                />
              </a>
            {/each}
          </div>
        </section>
      {/each}
    </div>
  {/if}
</div>
