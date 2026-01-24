<script>
    import { onMount, createEventDispatcher } from "svelte";
    import { api } from "$lib/api";
    import { auth } from "$lib/stores/auth";
    import { Loader2, Music, Youtube, Plus, Check } from "lucide-svelte";

    const dispatch = createEventDispatcher();

    let spotifyPlaylists = [];
    let youtubePlaylists = [];
    let loadingSpotify = false;
    let loadingYouTube = false;
    let importing = {}; // Map of sourceID -> boolean

    async function fetchSpotifyPlaylists() {
        loadingSpotify = true;
        try {
            const res = await api.get("/api/spotify/playlists", $auth.token);
            spotifyPlaylists = res.playlists || [];
        } catch (e) {
            console.error("Failed to fetch Spotify playlists", e);
        } finally {
            loadingSpotify = false;
        }
    }

    async function fetchYouTubePlaylists() {
        loadingYouTube = true;
        try {
            const res = await api.get(
                "/api/youtube/playlists/fresh",
                $auth.token,
            );
            youtubePlaylists = res.playlists || [];
        } catch (e) {
            console.error("Failed to fetch YouTube playlists", e);
        } finally {
            loadingYouTube = false;
        }
    }

    async function importPlaylist(playlist) {
        importing[playlist.source_id] = true;
        try {
            await api.post(
                "/api/playlists/import",
                {
                    platform: playlist.platform,
                    source_id: playlist.source_id,
                    is_public: false,
                },
                $auth.token,
            );
            // Emit event to refresh main list
            dispatch("import");
            playlist.imported = true;
        } catch (e) {
            console.error("Failed to import playlist", e);
            alert("Failed to import playlist: " + e.message);
        } finally {
            importing[playlist.source_id] = false;
        }
    }

    onMount(() => {
        fetchSpotifyPlaylists();
        fetchYouTubePlaylists();
    });
</script>

<div class="space-y-8">
    <!-- Spotify Section -->
    <div>
        <div class="flex items-center gap-2 mb-4">
            <Music class="text-green-500" />
            <h3 class="text-xl font-bold">Spotify Playlists</h3>
        </div>

        {#if loadingSpotify}
            <div class="flex justify-center p-8">
                <Loader2 class="animate-spin text-white/50" />
            </div>
        {:else if spotifyPlaylists.length === 0}
            <div
                class="p-8 text-center border border-white/10 rounded-xl bg-white/5 text-white/50"
            >
                No Spotify playlists found or not connected.
            </div>
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {#each spotifyPlaylists as playlist}
                    <div
                        class="glass p-4 rounded-xl flex items-center gap-4 group"
                    >
                        <img
                            src={playlist.cover_image || "/placeholder.png"}
                            alt={playlist.title}
                            class="w-16 h-16 rounded-lg object-cover bg-white/10"
                        />
                        <div class="flex-1 min-w-0">
                            <h4 class="font-bold truncate">{playlist.title}</h4>
                            <p class="text-sm text-white/50 truncate">
                                {playlist.track_count} tracks
                            </p>
                        </div>
                        <button
                            class="p-2 rounded-full bg-white/10 hover:bg-white/20 transition-colors disabled:opacity-50"
                            disabled={importing[playlist.source_id] ||
                                playlist.imported}
                            on:click={() => importPlaylist(playlist)}
                        >
                            {#if importing[playlist.source_id]}
                                <Loader2 size={20} class="animate-spin" />
                            {:else if playlist.imported}
                                <Check size={20} class="text-green-400" />
                            {:else}
                                <Plus size={20} />
                            {/if}
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>

    <!-- YouTube Section -->
    <div>
        <div class="flex items-center gap-2 mb-4">
            <Youtube class="text-red-500" />
            <h3 class="text-xl font-bold">YouTube Playlists</h3>
        </div>

        {#if loadingYouTube}
            <div class="flex justify-center p-8">
                <Loader2 class="animate-spin text-white/50" />
            </div>
        {:else if youtubePlaylists.length === 0}
            <div
                class="p-8 text-center border border-white/10 rounded-xl bg-white/5 text-white/50"
            >
                No YouTube playlists found or not connected.
            </div>
        {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {#each youtubePlaylists as playlist}
                    <div
                        class="glass p-4 rounded-xl flex items-center gap-4 group"
                    >
                        <img
                            src={playlist.cover_image || "/placeholder.png"}
                            alt={playlist.title}
                            class="w-16 h-16 rounded-lg object-cover bg-white/10"
                        />
                        <div class="flex-1 min-w-0">
                            <h4 class="font-bold truncate">{playlist.title}</h4>
                            <p class="text-sm text-white/50 truncate">
                                {playlist.track_count} tracks
                            </p>
                        </div>
                        <button
                            class="p-2 rounded-full bg-white/10 hover:bg-white/20 transition-colors disabled:opacity-50"
                            disabled={importing[playlist.source_id] ||
                                playlist.imported}
                            on:click={() => importPlaylist(playlist)}
                        >
                            {#if importing[playlist.source_id]}
                                <Loader2 size={20} class="animate-spin" />
                            {:else if playlist.imported}
                                <Check size={20} class="text-green-400" />
                            {:else}
                                <Plus size={20} />
                            {/if}
                        </button>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>

<style>
    .glass {
        background: rgba(255, 255, 255, 0.05);
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }
</style>
