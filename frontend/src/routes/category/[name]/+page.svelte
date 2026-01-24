<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import PlaylistCard from "$lib/components/PlaylistCard.svelte";
    import { api } from "$lib/api";

    export let data;

    let playlists = [];
    let loading = true;
    let error = null;
    let categoryName = decodeURIComponent(data.category);

    onMount(async () => {
        try {
            const res = await api.getPublicPlaylists();
            // Filter by category
            playlists = (res.playlists || []).filter(
                (p) => (p.category || "Uncategorized") === categoryName,
            );
        } catch (e) {
            console.error("Failed to load playlists", e);
            error = "Failed to load playlists.";
        } finally {
            loading = false;
        }
    });
</script>

<div class="space-y-8">
    <header>
        <div class="flex items-center gap-2 mb-2 text-sm text-zinc-400">
            <a href="/" class="hover:text-white transition-colors">Home</a>
            <span>/</span>
            <span class="text-white">{categoryName}</span>
        </div>
        <h1 class="text-3xl font-bold text-white tracking-tight">
            {categoryName}
        </h1>
        <p class="text-zinc-400">All public playlists in this category.</p>
    </header>

    {#if loading}
        <div class="text-zinc-500">Loading...</div>
    {:else if error}
        <div class="text-red-500">{error}</div>
    {:else if playlists.length === 0}
        <div class="text-zinc-500">No playlists found in this category.</div>
    {:else}
        <div
            class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6"
        >
            {#each playlists as playlist}
                <PlaylistCard
                    id={playlist.id}
                    title={playlist.title}
                    description={playlist.description || playlist.platform}
                    image={playlist.cover_image ||
                        "https://images.unsplash.com/photo-1614680376593-902f74cf0d41?w=800&auto=format&fit=crop&q=60"}
                />
            {/each}
        </div>
    {/if}
</div>
