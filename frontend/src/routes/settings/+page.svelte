<script>
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { auth } from "$lib/stores/auth";
    import { Check, X, ExternalLink, Plus, LogIn } from "lucide-svelte";
    import PlaylistImportModal from "$lib/components/PlaylistImportModal.svelte";

    let status = {
        spotify_connected: false,
        youtube_connected: false,
        username: "",
    };
    let loading = true;
    let showImportModal = false;
    let showLoginPopup = false;

    $: if ($auth.token) {
        fetchStatus();
    } else {
        loading = false;
    }

    onMount(() => {
        if (!$auth.token) {
            setTimeout(() => {
                showLoginPopup = true;
            }, 3000);
        }
    });

    async function fetchStatus() {
        try {
            status = await api.getConnectionStatus();
        } catch (e) {
            console.error("Failed to fetch status", e);
        } finally {
            loading = false;
        }
    }

    function connectSpotify() {
        window.location.href = api.getSpotifyLink();
    }

    function connectYouTube() {
        window.location.href = api.getYouTubeLink();
    }

    function onImportComplete() {
        // No-op for settings page
    }
</script>

<div class="max-w-4xl mx-auto space-y-8">
    <header class="flex items-center justify-between">
        <div>
            <h1 class="text-3xl font-bold text-white tracking-tight mb-2">
                Settings
            </h1>
            <p class="text-zinc-400">
                Manage your connected accounts and preferences.
            </p>
        </div>
        {#if $auth.token}
            <button
                on:click={() => (showImportModal = true)}
                class="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-full font-bold hover:scale-105 transition-transform"
            >
                <Plus size={20} />
                Import Playlist
            </button>
        {/if}
    </header>

    {#if showImportModal}
        <PlaylistImportModal
            platform="spotify"
            onClose={() => (showImportModal = false)}
            {onImportComplete}
        />
    {/if}

    <!-- Login Popup -->
    {#if showLoginPopup && !$auth.token}
        <div class="fixed inset-0 z-50 flex items-center justify-center px-4">
            <div
                class="absolute inset-0 bg-black/60 backdrop-blur-sm"
                on:click={() => (showLoginPopup = false)}
            ></div>
            <div
                class="relative bg-zinc-900 border border-white/10 p-8 rounded-2xl max-w-md w-full shadow-2xl space-y-6 text-center"
            >
                <div
                    class="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center mx-auto mb-4"
                >
                    <LogIn class="w-8 h-8 text-white" />
                </div>
                <h3 class="text-2xl font-bold text-white">
                    Connect Your Music
                </h3>
                <p class="text-zinc-400">
                    Please sign up or log in to connect your Spotify and YouTube
                    accounts and start syncing playlists.
                </p>
                <div class="flex gap-4 pt-2">
                    <a
                        href="/login"
                        class="flex-1 py-3 rounded-lg font-semibold bg-white/10 hover:bg-white/20 text-white transition-colors"
                    >
                        Log In
                    </a>
                    <a
                        href="/signup"
                        class="flex-1 py-3 rounded-lg font-semibold bg-white text-black hover:scale-105 transition-transform"
                    >
                        Sign Up
                    </a>
                </div>
                <button
                    on:click={() => (showLoginPopup = false)}
                    class="absolute top-4 right-4 text-zinc-500 hover:text-white"
                >
                    <X size={20} />
                </button>
            </div>
        </div>
    {/if}

    <section class="space-y-6">
        <h2 class="text-xl font-semibold text-white">Connected Platforms</h2>

        {#if !$auth.token}
            <!-- Unauthenticated State -->
            <div
                class="p-12 rounded-xl border border-white/5 bg-white/5 backdrop-blur-sm text-center space-y-6"
            >
                <div class="max-w-md mx-auto space-y-6">
                    <h3 class="text-2xl font-bold text-white">
                        Sign in to Manage Connections
                    </h3>
                    <p class="text-zinc-400">
                        You need to be logged in to connect your Spotify and
                        YouTube accounts.
                    </p>
                    <div class="flex gap-4 justify-center">
                        <a
                            href="/login"
                            class="px-8 py-3 rounded-full font-semibold bg-white/10 hover:bg-white/20 text-white transition-colors"
                        >
                            Log In
                        </a>
                        <a
                            href="/signup"
                            class="px-8 py-3 rounded-full font-semibold bg-white text-black hover:scale-105 transition-transform"
                        >
                            Sign Up
                        </a>
                    </div>
                </div>
            </div>
        {:else}
            <!-- Authenticated State -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Spotify Card -->
                <div
                    class="p-6 rounded-xl border border-white/5 bg-white/5 backdrop-blur-sm flex flex-col justify-between h-48"
                >
                    <div class="flex justify-between items-start">
                        <div>
                            <h3 class="text-lg font-bold text-white mb-1">
                                Spotify
                            </h3>
                            <p class="text-sm text-zinc-400">
                                Import and sync playlists.
                            </p>
                        </div>
                        {#if status.spotify_connected}
                            <div
                                class="bg-green-500/20 text-green-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1"
                            >
                                <Check class="w-3 h-3" /> Connected
                            </div>
                        {:else}
                            <div
                                class="bg-zinc-500/20 text-zinc-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1"
                            >
                                <X class="w-3 h-3" /> Not Connected
                            </div>
                        {/if}
                    </div>

                    <button
                        on:click={connectSpotify}
                        disabled={status.spotify_connected}
                        class="w-full py-3 rounded-lg font-semibold text-sm transition-all duration-200 flex items-center justify-center gap-2
              {status.spotify_connected
                            ? 'bg-white/5 text-zinc-500 cursor-default'
                            : 'bg-[#1DB954] text-black hover:scale-[1.02] shadow-lg shadow-green-900/20'}"
                    >
                        {#if status.spotify_connected}
                            Account Linked
                        {:else}
                            Connect Spotify <ExternalLink class="w-4 h-4" />
                        {/if}
                    </button>
                </div>

                <!-- YouTube Music Card -->
                <div
                    class="p-6 rounded-xl border border-white/5 bg-white/5 backdrop-blur-sm flex flex-col justify-between h-48"
                >
                    <div class="flex justify-between items-start">
                        <div>
                            <h3 class="text-lg font-bold text-white mb-1">
                                YouTube Music
                            </h3>
                            <p class="text-sm text-zinc-400">
                                Sync to your Google account.
                            </p>
                        </div>
                        {#if status.youtube_connected}
                            <div
                                class="bg-green-500/20 text-green-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1"
                            >
                                <Check class="w-3 h-3" /> Connected
                            </div>
                        {:else}
                            <div
                                class="bg-zinc-500/20 text-zinc-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1"
                            >
                                <X class="w-3 h-3" /> Not Connected
                            </div>
                        {/if}
                    </div>

                    <button
                        on:click={connectYouTube}
                        disabled={status.youtube_connected}
                        class="w-full py-3 rounded-lg font-semibold text-sm transition-all duration-200 flex items-center justify-center gap-2
              {status.youtube_connected
                            ? 'bg-white/5 text-zinc-500 cursor-default'
                            : 'bg-[#FF0000] text-white hover:scale-[1.02] shadow-lg shadow-red-900/20'}"
                    >
                        {#if status.youtube_connected}
                            Account Linked
                        {:else}
                            Connect YouTube <ExternalLink class="w-4 h-4" />
                        {/if}
                    </button>
                </div>
            </div>
        {/if}
    </section>
</div>
