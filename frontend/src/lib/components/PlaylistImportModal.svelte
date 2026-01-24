<script>
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { auth } from "$lib/stores/auth";
    import { Music, X, Check } from "lucide-svelte";

    export let platform = "spotify"; // "spotify" or "youtube"
    export let onClose = () => {};
    export let onImportComplete = () => {};

    let playlists = [];
    let loading = true;
    let selectedPlaylists = new Set();
    let importing = false;
    let existingPlaylistIds = new Set();

    onMount(async () => {
        await Promise.all([loadPlaylists(), loadExistingLibrary()]);
    });

    async function loadExistingLibrary() {
        try {
            const res = await api.getPlaylists();
            const library = res.playlists || [];
            existingPlaylistIds = new Set(library.map((p) => p.source_id));
        } catch (e) {
            console.error("Failed to load library for duplicate check", e);
        }
    }

    async function loadPlaylists() {
        loading = true;
        try {
            const endpoint =
                platform === "spotify"
                    ? "/api/spotify/playlists"
                    : "/api/youtube/playlists";
            const response = await api.get(endpoint, $auth.token);
            playlists = response.playlists || [];
        } catch (err) {
            console.error(`Failed to load ${platform} playlists:`, err);
        }
        loading = false;
    }

    function togglePlaylist(id) {
        if (selectedPlaylists.has(id)) {
            selectedPlaylists.delete(id);
        } else {
            if (selectedPlaylists.size >= 3) {
                alert("You can only import up to 3 playlists at a time.");
                return;
            }
            selectedPlaylists.add(id);
        }
        selectedPlaylists = selectedPlaylists; // Trigger reactivity
    }

    function selectAll() {
        selectedPlaylists = new Set(playlists.map((p) => p.id));
    }

    async function importSelected() {
        if (selectedPlaylists.size === 0) {
            alert("Please select at least one playlist");
            return;
        }

        importing = true;
        try {
            const playlistsToImport = Array.from(selectedPlaylists).map(
                (id) => ({
                    platform,
                    source_id: id,
                }),
            );

            await api.post(
                "/api/playlists/batch-import",
                {
                    playlists: playlistsToImport,
                },
                $auth.token,
            );

            onImportComplete();
            onClose();
        } catch (err) {
            console.error("Import failed:", err);
            alert("Failed to import playlists. Check console for details.");
        }
        importing = false;
    }

    async function importAll() {
        selectedPlaylists = new Set(playlists.map((p) => p.id));
        await importSelected();
    }
</script>

<div class="modal-overlay" on:click={onClose}>
    <div class="modal-content glass" on:click|stopPropagation>
        <div class="modal-header">
            <div>
                <h2 class="text-2xl font-bold capitalize">
                    Import from {platform}
                </h2>
                <p class="text-sm text-white/60 mt-1">
                    Select playlists to import into EchoBridge
                </p>
            </div>
            <button on:click={onClose} class="close-btn">
                <X size={24} />
            </button>
        </div>

        <div class="modal-body">
            {#if loading}
                <div class="loader">Loading your playlists...</div>
            {:else if playlists.length === 0}
                <div class="empty-state">
                    <Music size={48} />
                    <p>No playlists found on {platform}</p>
                </div>
            {:else}
                <div class="playlists-grid">
                    {#each playlists as playlist}
                        {@const isImported = existingPlaylistIds.has(
                            playlist.id,
                        )}
                        <div
                            class="playlist-item {selectedPlaylists.has(
                                playlist.id,
                            )
                                ? 'selected'
                                : ''} {isImported
                                ? 'opacity-50 cursor-not-allowed'
                                : ''}"
                            on:click={() =>
                                !isImported && togglePlaylist(playlist.id)}
                        >
                            <div class="playlist-checkbox">
                                {#if isImported}
                                    <div
                                        class="w-5 h-5 flex items-center justify-center bg-white/20 rounded-full"
                                    >
                                        <Check size={12} />
                                    </div>
                                {:else}
                                    <input
                                        type="checkbox"
                                        checked={selectedPlaylists.has(
                                            playlist.id,
                                        )}
                                        on:click|stopPropagation={() =>
                                            togglePlaylist(playlist.id)}
                                    />
                                {/if}
                            </div>
                            {#if playlist.cover_image}
                                <img
                                    src={playlist.cover_image}
                                    alt={playlist.title}
                                    class="playlist-cover"
                                />
                            {:else}
                                <div class="playlist-cover-placeholder">
                                    <Music size={24} />
                                </div>
                            {/if}
                            <div class="playlist-info">
                                <div class="flex items-center gap-2">
                                    <h3 class="playlist-title">
                                        {playlist.title}
                                    </h3>
                                    {#if isImported}
                                        <span
                                            class="text-[10px] font-bold uppercase tracking-wider bg-white/20 px-1.5 py-0.5 rounded text-white/80"
                                            >Imported</span
                                        >
                                    {/if}
                                </div>
                                {#if playlist.description}
                                    <p class="playlist-description">
                                        {playlist.description}
                                    </p>
                                {/if}
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>

        <div class="modal-footer">
            <div class="selection-info">
                {selectedPlaylists.size} of {playlists.length} selected
            </div>
            <div class="button-group">
                <!-- Select All removed due to import limit -->
                <!-- Import All removed due to import limit -->
                <button
                    on:click={importSelected}
                    class="btn-primary"
                    disabled={importing ||
                        loading ||
                        selectedPlaylists.size === 0}
                >
                    {importing ? "Importing..." : "Import Selected"}
                </button>
            </div>
        </div>
    </div>
</div>

<style>
    .modal-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.75);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
        backdrop-filter: blur(4px);
    }

    .modal-content {
        width: 90%;
        max-width: 800px;
        max-height: 85vh;
        border-radius: 16px;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    .modal-header {
        padding: 24px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
    }

    .close-btn {
        background: transparent;
        border: none;
        color: white;
        cursor: pointer;
        padding: 4px;
        border-radius: 8px;
        transition: background 0.2s;
    }

    .close-btn:hover {
        background: rgba(255, 255, 255, 0.1);
    }

    .modal-body {
        flex: 1;
        overflow-y: auto;
        padding: 24px;
    }

    .loader,
    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 48px;
        color: rgba(255, 255, 255, 0.6);
        gap: 16px;
    }

    .playlists-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
        gap: 16px;
    }

    .playlist-item {
        background: rgba(255, 255, 255, 0.05);
        border: 2px solid transparent;
        border-radius: 12px;
        padding: 12px;
        cursor: pointer;
        transition: all 0.2s;
        display: flex;
        gap: 12px;
    }

    .playlist-item:hover {
        background: rgba(255, 255, 255, 0.1);
    }

    .playlist-item.selected {
        border-color: #1db954;
        background: rgba(29, 185, 84, 0.1);
    }

    .playlist-checkbox {
        flex-shrink: 0;
        display: flex;
        align-items: center;
    }

    .playlist-checkbox input[type="checkbox"] {
        width: 20px;
        height: 20px;
        cursor: pointer;
        accent-color: #1db954;
    }

    .playlist-cover,
    .playlist-cover-placeholder {
        width: 60px;
        height: 60px;
        border-radius: 8px;
        object-fit: cover;
        flex-shrink: 0;
    }

    .playlist-cover-placeholder {
        background: rgba(255, 255, 255, 0.1);
        display: flex;
        align-items: center;
        justify-content: center;
        color: rgba(255, 255, 255, 0.4);
    }

    .playlist-info {
        flex: 1;
        min-width: 0;
    }

    .playlist-title {
        font-weight: 600;
        margin-bottom: 4px;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .playlist-description {
        font-size: 0.875rem;
        color: rgba(255, 255, 255, 0.6);
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .modal-footer {
        padding: 24px;
        border-top: 1px solid rgba(255, 255, 255, 0.1);
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 16px;
        flex-wrap: wrap;
    }

    .selection-info {
        color: rgba(255, 255, 255, 0.7);
        font-size: 0.875rem;
    }

    .button-group {
        display: flex;
        gap: 12px;
    }

    .btn-primary,
    .btn-secondary {
        padding: 10px 20px;
        border-radius: 8px;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.2s;
        border: none;
    }

    .btn-primary {
        background: #1db954;
        color: white;
    }

    .btn-primary:hover:not(:disabled) {
        background: #1ed760;
        transform: scale(1.05);
    }

    .btn-secondary {
        background: rgba(255, 255, 255, 0.1);
        color: white;
    }

    .btn-secondary:hover:not(:disabled) {
        background: rgba(255, 255, 255, 0.2);
    }

    .btn-primary:disabled,
    .btn-secondary:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
</style>
