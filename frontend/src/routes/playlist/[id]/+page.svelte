<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { api } from "$lib/api";
    import { auth } from "$lib/stores/auth";
    import {
        Play,
        Clock,
        Music2,
        Calendar,
        Import,
        X,
        LogIn,
        Globe,
        Lock,
    } from "lucide-svelte";

    let activeEmbedId = null;

    function toggleEmbed(track) {
        if (activeEmbedId === track.id) {
            activeEmbedId = null;
        } else {
            activeEmbedId = track.id;
        }
    }

    let playlist = null;
    let loading = true;
    let error = null;
    let showImportModal = false;
    let importLoading = false;
    let importError = null;
    let importSuccess = null;
    let showConnectPrompt = null; // { platform: 'spotify' | 'youtube' }
    let showLoginPrompt = false;
    let isPublic = false;

    onMount(async () => {
        console.log("Playlist page mounted");
        const id = $page.params.id;
        console.log("Playlist ID:", id);
        try {
            console.log("Fetching playlist...");
            playlist = await api.getPlaylist(id);
            isPublic = playlist.is_public;
            console.log("Playlist fetched:", playlist);
        } catch (e) {
            console.error("Failed to load playlist", e);
            error = "Failed to load playlist.";
        } finally {
            loading = false;
            console.log("Loading set to false");
        }
    });

    async function handleToggle() {
        if (!$auth.token) {
            showLoginPrompt = true;
            return;
        }

        const newStatus = !isPublic;
        try {
            await api.patch(
                `/api/playlists/${playlist.id}/public`,
                { is_public: newStatus },
                $auth.token,
            );
            isPublic = newStatus;
            playlist.is_public = newStatus;
        } catch (e) {
            console.error("Failed to toggle playlist visibility", e);
            alert("Failed to update playlist visibility");
        }
    }

    async function handleImport(platform) {
        importLoading = true;
        importError = null;
        importSuccess = null;
        showConnectPrompt = null;

        try {
            await api.importPlaylist(playlist.id, platform);
            importSuccess = `Successfully imported to ${platform}!`;
            setTimeout(() => {
                showImportModal = false;
                importSuccess = null;
            }, 2000);
        } catch (e) {
            console.error("Import failed", e);
            if (e.code === "PLATFORM_NOT_CONNECTED") {
                showConnectPrompt = { platform };
            } else {
                importError =
                    e.message ||
                    "Import failed. Please check your connection settings.";
            }
        } finally {
            importLoading = false;
        }
    }

    function connectPlatform(platform) {
        if (platform === "spotify") {
            window.location.href = api.getSpotifyLink();
        } else if (platform === "youtube") {
            window.location.href = api.getYouTubeLink();
        }
    }
</script>

<div class="space-y-8 pb-20 relative">
    {#if loading}
        <div class="text-zinc-500">Loading playlist...</div>
    {:else if error}
        <div class="text-red-500">{error}</div>
    {:else if playlist}
        <!-- Header -->
        <div
            class="flex flex-col md:flex-row gap-8 items-center md:items-end text-center md:text-left"
        >
            <div
                class="w-52 h-52 shadow-2xl rounded-md overflow-hidden bg-zinc-800 flex-shrink-0 mx-auto md:mx-0"
            >
                <img
                    src={playlist.cover_image ||
                        "https://images.unsplash.com/photo-1614680376593-902f74cf0d41?w=800&auto=format&fit=crop&q=60"}
                    alt={playlist.title}
                    class="w-full h-full object-cover"
                />
            </div>
            <div class="flex-1 space-y-4">
                <span
                    class="text-xs font-bold uppercase tracking-wider text-zinc-400"
                    >Playlist</span
                >
                <h1
                    class="text-3xl md:text-5xl font-bold text-white tracking-tighter"
                >
                    {playlist.title}
                </h1>
                <!-- Description removed as requested -->
                <div
                    class="flex items-center justify-center md:justify-start gap-2 text-sm text-zinc-400 font-medium"
                >
                    <Music2 class="w-4 h-4" />
                    <span>{playlist.tracks.length} songs</span>
                </div>
            </div>
        </div>

        <!-- Actions -->
        <div
            class="flex items-center justify-center md:justify-start gap-4 py-6"
        >
            <button
                on:click={() => {
                    if (!$auth.token) {
                        showLoginPrompt = true;
                    } else {
                        showImportModal = true;
                    }
                }}
                class="px-6 py-3 rounded-full border border-white/20 hover:border-white text-white font-semibold text-sm transition-colors flex items-center gap-2"
            >
                <Import class="w-4 h-4" />
                Import to...
            </button>

            {#if $auth.user && playlist.owner_id === $auth.user.id}
                <button
                    on:click={handleToggle}
                    class="w-12 h-12 rounded-full border border-white/20 hover:border-white flex items-center justify-center text-white transition-colors"
                    title={isPublic ? "Public" : "Private"}
                >
                    {#if isPublic}
                        <Globe class="w-5 h-5" />
                    {:else}
                        <Lock class="w-5 h-5" />
                    {/if}
                </button>
            {/if}
        </div>

        <!-- Track List -->
        <div class="w-full">
            <!-- Header Row -->
            <div
                class="grid grid-cols-[1fr_1fr_40px] gap-4 px-4 py-2 border-b border-white/5 text-sm font-medium text-zinc-400 uppercase tracking-wider mb-2"
            >
                <span>Title</span>
                <span class="hidden md:block">Album</span>
                <span class="hidden md:block"><Clock class="w-4 h-4" /></span>
            </div>

            <!-- Tracks -->
            <div class="space-y-1">
                {#each playlist.tracks as track, i}
                    {#if activeEmbedId === track.id && track.spotify_id}
                        <div class="px-4 py-2 bg-white/5 rounded-md">
                            <iframe
                                src="https://open.spotify.com/embed/track/{track.spotify_id}?utm_source=generator&theme=0"
                                width="100%"
                                height="80"
                                frameBorder="0"
                                allowfullscreen=""
                                allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
                                loading="lazy"
                                title="Spotify Embed"
                                class="rounded-md shadow-lg"
                            ></iframe>
                            <button
                                on:click={() => (activeEmbedId = null)}
                                class="text-xs text-zinc-500 hover:text-white mt-2 underline w-full text-center"
                            >
                                Close Player
                            </button>
                        </div>
                    {:else}
                        <div
                            class="group grid grid-cols-[1fr_1fr_40px] gap-4 px-4 py-3 rounded-md hover:bg-white/5 transition-colors items-center text-sm text-zinc-400"
                        >
                            <div class="flex flex-col relative pl-8">
                                <div
                                    class="absolute left-0 top-1/2 -translate-y-1/2 w-4 h-4 flex items-center justify-center"
                                >
                                    {#if track.spotify_id}
                                        <button
                                            on:click={() => toggleEmbed(track)}
                                            class="text-white opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <Play
                                                class="w-4 h-4 fill-current"
                                            />
                                        </button>
                                    {/if}
                                </div>
                                <span class="text-white font-medium truncate"
                                    >{track.title}</span
                                >
                                <span
                                    class="group-hover:text-white transition-colors"
                                    >{track.artist}</span
                                >
                            </div>
                            <span
                                class="hidden md:block truncate group-hover:text-white transition-colors"
                                >{track.album}</span
                            >
                            <span class="hidden md:block font-mono text-xs"
                                >3:45</span
                            >
                        </div>
                    {/if}
                {/each}
            </div>
        </div>
    {/if}

    <!-- Import Modal -->
    {#if showImportModal}
        <div
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
        >
            <div
                class="bg-zinc-900 border border-white/10 rounded-xl p-6 max-w-md w-full shadow-2xl space-y-6 relative"
            >
                <button
                    on:click={() => {
                        showImportModal = false;
                        showConnectPrompt = null;
                    }}
                    class="absolute top-4 right-4 text-zinc-400 hover:text-white"
                >
                    <X class="w-5 h-5" />
                </button>

                {#if showConnectPrompt}
                    <div class="text-center space-y-4">
                        <div
                            class="w-12 h-12 rounded-full bg-zinc-800 flex items-center justify-center mx-auto text-zinc-400"
                        >
                            <Import class="w-6 h-6" />
                        </div>
                        <div>
                            <h2 class="text-xl font-bold text-white">
                                Connect {showConnectPrompt.platform ===
                                "spotify"
                                    ? "Spotify"
                                    : "YouTube"}?
                            </h2>
                            <p class="text-zinc-400 text-sm mt-2">
                                You need to connect your account to import
                                playlists.
                            </p>
                        </div>

                        <div class="flex gap-3 pt-2">
                            <button
                                on:click={() => {
                                    showConnectPrompt = null;
                                }}
                                class="flex-1 py-2 rounded-md bg-white/5 hover:bg-white/10 text-white font-medium text-sm transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                on:click={() =>
                                    connectPlatform(showConnectPrompt.platform)}
                                class="flex-1 py-2 rounded-md bg-white hover:bg-zinc-200 text-black font-bold text-sm transition-colors shadow-lg shadow-white/20"
                            >
                                Connect Now
                            </button>
                        </div>
                    </div>
                {:else}
                    <div class="text-center">
                        <h2 class="text-xl font-bold text-white mb-2">
                            Import Playlist
                        </h2>
                        <p class="text-zinc-400 text-sm">
                            Choose a platform to add this playlist to your
                            library.
                        </p>
                    </div>

                    {#if importError}
                        <div
                            class="bg-red-500/10 text-red-400 p-3 rounded-md text-sm text-center"
                        >
                            {importError}
                        </div>
                    {/if}

                    {#if importSuccess}
                        <div
                            class="bg-green-500/10 text-green-400 p-3 rounded-md text-sm text-center"
                        >
                            {importSuccess}
                        </div>
                    {/if}

                    <div class="grid grid-cols-2 gap-4">
                        <button
                            on:click={() => handleImport("spotify")}
                            disabled={importLoading}
                            class="flex flex-col items-center justify-center gap-3 p-4 rounded-lg bg-[#1DB954]/10 border border-[#1DB954]/20 hover:bg-[#1DB954]/20 transition-colors group disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            <div
                                class="w-10 h-10 rounded-full bg-[#1DB954] text-black flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform"
                            >
                                <Music2 class="w-5 h-5" />
                            </div>
                            <span class="text-white font-medium text-sm"
                                >Spotify</span
                            >
                        </button>

                        <button
                            on:click={() => handleImport("youtube")}
                            disabled={importLoading}
                            class="flex flex-col items-center justify-center gap-3 p-4 rounded-lg bg-[#FF0000]/10 border border-[#FF0000]/20 hover:bg-[#FF0000]/20 transition-colors group disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            <div
                                class="w-10 h-10 rounded-full bg-[#FF0000] text-white flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform"
                            >
                                <Play class="w-5 h-5 fill-current" />
                            </div>
                            <span class="text-white font-medium text-sm"
                                >YouTube</span
                            >
                        </button>
                    </div>
                {/if}
            </div>
        </div>
    {/if}

    <!-- Login Prompt Modal -->
    {#if showLoginPrompt}
        <div
            class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
        >
            <div
                class="bg-zinc-900 border border-white/10 rounded-xl p-6 max-w-md w-full shadow-2xl space-y-6 relative"
            >
                <button
                    on:click={() => (showLoginPrompt = false)}
                    class="absolute top-4 right-4 text-zinc-400 hover:text-white"
                >
                    <X class="w-5 h-5" />
                </button>

                <div class="text-center space-y-4">
                    <div
                        class="w-12 h-12 rounded-full bg-zinc-800 flex items-center justify-center mx-auto text-zinc-400"
                    >
                        <LogIn class="w-6 h-6" />
                    </div>
                    <div>
                        <h2 class="text-xl font-bold text-white">
                            Login Required
                        </h2>
                        <p class="text-zinc-400 text-sm mt-2">
                            You need to be logged in to import playlists to your
                            library.
                        </p>
                    </div>

                    <div class="flex gap-3 pt-2">
                        <button
                            on:click={() => (showLoginPrompt = false)}
                            class="flex-1 py-2 rounded-md bg-white/5 hover:bg-white/10 text-white font-medium text-sm transition-colors"
                        >
                            Cancel
                        </button>
                        <a
                            href="/login"
                            class="flex-1 py-2 rounded-md bg-white hover:bg-zinc-200 text-black font-bold text-sm transition-colors shadow-lg shadow-white/20 flex items-center justify-center"
                        >
                            Log In
                        </a>
                    </div>
                </div>
            </div>
        </div>
    {/if}
</div>
