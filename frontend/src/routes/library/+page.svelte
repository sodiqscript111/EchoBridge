<script>
    import { onMount } from "svelte";
    import PlaylistCard from "$lib/components/PlaylistCard.svelte";
    import PlaylistImportModal from "$lib/components/PlaylistImportModal.svelte";
    import { api } from "$lib/api";
    import { auth } from "$lib/stores/auth";
    import { Plus } from "lucide-svelte";

    let playlists = [];
    let loading = true;
    let error = null;
    let showImportModal = false;

    onMount(async () => {
        await loadLibrary();
    });

    async function loadLibrary() {
        if (!$auth.token) {
            loading = false;
            return;
        }

        try {
            const res = await api.getPlaylists();
            playlists = res.playlists || [];
        } catch (e) {
            console.error("Failed to load library", e);
            error = "Failed to load your library.";
        } finally {
            loading = false;
        }
    }

    async function handleToggle(event) {
        const { id, isPublic } = event.detail;
        try {
            await api.patch(
                `/api/playlists/${id}/public`,
                { is_public: !isPublic },
                $auth.token,
            );
            // Update local state
            playlists = playlists.map((p) =>
                p.id === id ? { ...p, is_public: !isPublic } : p,
            );
        } catch (e) {
            console.error("Failed to toggle playlist visibility", e);
            alert("Failed to update playlist visibility");
        }
    }

    function onImportComplete() {
        loadLibrary();
    }
</script>

<div class="space-y-8">
    <header class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold text-white tracking-tight mb-2">
                Your Library
            </h1>
            <p class="text-zinc-400">
                All your imported playlists from Spotify and YouTube.
            </p>
        </div>
        <button
            on:click={() => (showImportModal = true)}
            class="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-full font-bold hover:scale-105 transition-transform"
        >
            <Plus size={20} />
            Import Playlist
        </button>
    </header>

    {#if loading}
        <div class="text-zinc-500">Loading library...</div>
    {:else if !$auth.token}
        <div class="text-zinc-500">Please log in to view your library.</div>
    {:else if error}
        <div class="text-red-500">{error}</div>
    {:else if playlists.length === 0}
        <div class="text-zinc-500">
            You haven't imported any playlists yet.
            <a href="/settings" class="text-white hover:underline"
                >Connect your accounts</a
            > to get started.
        </div>
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
                    isPublic={playlist.is_public}
                    showToggle={true}
                    on:toggle={handleToggle}
                />
            {/each}
        </div>
    {/if}
</div>

{#if showImportModal}
    <PlaylistImportModal
        platform="spotify"
        onClose={() => (showImportModal = false)}
        {onImportComplete}
    />
{/if}
