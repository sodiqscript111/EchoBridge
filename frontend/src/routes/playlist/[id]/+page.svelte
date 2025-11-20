<script>
    import { page } from "$app/stores";
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import {
        Play,
        Clock,
        Music,
        Share2,
        Heart,
        MoreVertical,
    } from "lucide-svelte";

    let playlist = null;
    let loading = true;
    let error = null;
    const playlistId = $page.params.id;

    onMount(async () => {
        try {
            playlist = await api.get(`/playlists/${playlistId}`);
        } catch (e) {
            error = e.message;
        } finally {
            loading = false;
        }
    });
</script>

<div class="min-h-screen pb-20">
    {#if loading}
        <div class="flex items-center justify-center h-[80vh]">
            <div
                class="w-12 h-12 border-4 border-white/20 border-t-apple-blue rounded-full animate-spin"
            ></div>
        </div>
    {:else if error}
        <div class="flex items-center justify-center h-[80vh]">
            <div class="text-center">
                <h2 class="text-2xl font-bold text-red-400 mb-2">Error</h2>
                <p class="text-white/60">{error}</p>
                <a
                    href="/public"
                    class="inline-block mt-6 px-6 py-2 bg-white/10 rounded-full hover:bg-white/20 transition-colors"
                >
                    Back to Public Playlists
                </a>
            </div>
        </div>
    {:else if playlist}
        <!-- Header -->
        <div class="relative h-[45vh] w-full overflow-hidden">
            <div
                class="absolute inset-0 bg-gradient-to-b from-transparent via-black/50 to-black z-0"
            ></div>
            {#if playlist.cover_image}
                <img
                    src={playlist.cover_image}
                    alt={playlist.title}
                    class="absolute inset-0 w-full h-full object-cover opacity-40 blur-2xl scale-110"
                />
            {/if}

            <div
                class="relative z-10 h-full max-w-7xl mx-auto px-8 flex items-end pb-8"
            >
                <div class="flex items-end gap-6">
                    <div
                        class="w-56 h-56 rounded-2xl shadow-2xl overflow-hidden bg-gradient-to-br from-white/10 to-white/5 border border-white/10 flex-shrink-0 backdrop-blur-sm"
                    >
                        {#if playlist.cover_image}
                            <img
                                src={playlist.cover_image}
                                alt={playlist.title}
                                class="w-full h-full object-cover"
                            />
                        {:else}
                            <div
                                class="w-full h-full flex items-center justify-center bg-gradient-to-br from-gray-800 to-black"
                            >
                                <Music size={64} class="text-white/20" />
                            </div>
                        {/if}
                    </div>

                    <div class="mb-2 flex-1">
                        <div class="flex items-center gap-3 mb-2">
                            <span
                                class="text-xs font-bold text-apple-blue uppercase tracking-widest"
                            >
                                {playlist.platform} Playlist
                            </span>
                            {#if playlist.category}
                                <span
                                    class="px-2 py-0.5 text-[10px] font-bold bg-white/10 rounded-full text-white/80 uppercase tracking-wider border border-white/10"
                                >
                                    {playlist.category}
                                </span>
                            {/if}
                        </div>
                        <h1
                            class="text-5xl font-black mt-2 mb-3 tracking-tight drop-shadow-lg"
                        >
                            {playlist.title}
                        </h1>
                        <p
                            class="text-lg text-white/70 mb-4 max-w-2xl line-clamp-2 font-medium"
                        >
                            {playlist.description ||
                                "No description available."}
                        </p>

                        <div class="flex items-center gap-3 mt-6">
                            <button
                                class="px-6 py-3 bg-apple-blue hover:bg-blue-500 text-white rounded-full font-bold flex items-center gap-2 transition-all transform hover:scale-105 shadow-xl shadow-blue-500/30"
                            >
                                <Play size={18} fill="currentColor" />
                                Play All
                            </button>
                            <button
                                class="p-3 bg-white/10 hover:bg-white/20 rounded-full transition-all backdrop-blur-md border border-white/10"
                            >
                                <Heart size={20} />
                            </button>
                            <button
                                class="p-3 bg-white/10 hover:bg-white/20 rounded-full transition-all backdrop-blur-md border border-white/10"
                            >
                                <Share2 size={20} />
                            </button>
                            <button
                                class="p-3 bg-white/10 hover:bg-white/20 rounded-full transition-all backdrop-blur-md border border-white/10"
                            >
                                <MoreVertical size={20} />
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Tracks List -->
        <div class="max-w-7xl mx-auto px-8 mt-6">
            <div class="mb-4 flex items-center justify-between">
                <h2 class="text-2xl font-bold">Tracks</h2>
                <span class="text-sm text-white/50"
                    >{playlist.tracks?.length || 0} songs</span
                >
            </div>

            <div
                class="glass rounded-3xl overflow-hidden border border-white/5"
            >
                <table class="w-full text-left">
                    <thead class="bg-white/5 border-b border-white/10">
                        <tr>
                            <th
                                class="px-6 py-3 text-xs font-semibold text-white/40 uppercase tracking-wider w-16"
                                >#</th
                            >
                            <th
                                class="px-6 py-3 text-xs font-semibold text-white/40 uppercase tracking-wider"
                                >Title</th
                            >
                            <th
                                class="px-6 py-3 text-xs font-semibold text-white/40 uppercase tracking-wider"
                                >Album</th
                            >
                            <th
                                class="px-6 py-3 text-xs font-semibold text-white/40 uppercase tracking-wider text-right"
                            >
                                <Clock size={14} class="inline" />
                            </th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-white/5">
                        {#each playlist.tracks as track, i}
                            <tr
                                class="group hover:bg-white/5 transition-all duration-200"
                            >
                                <td
                                    class="px-6 py-4 text-sm text-white/30 group-hover:text-apple-blue transition-colors font-medium"
                                >
                                    <span class="group-hover:hidden"
                                        >{i + 1}</span
                                    >
                                    <Play
                                        size={14}
                                        class="hidden group-hover:inline"
                                        fill="currentColor"
                                    />
                                </td>
                                <td class="px-6 py-4">
                                    <div
                                        class="font-semibold text-white group-hover:text-apple-blue transition-colors"
                                    >
                                        {track.title}
                                    </div>
                                    <div class="text-sm text-white/50 mt-0.5">
                                        {track.artist}
                                    </div>
                                </td>
                                <td
                                    class="px-6 py-4 text-sm text-white/50 font-medium"
                                >
                                    {track.album || "—"}
                                </td>
                                <td
                                    class="px-6 py-4 text-sm text-white/30 text-right font-mono"
                                >
                                    3:45
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>

                {#if playlist.tracks?.length === 0}
                    <div class="p-16 text-center">
                        <Music size={48} class="mx-auto text-white/10 mb-4" />
                        <p class="text-white/40 text-lg">
                            No tracks found in this playlist.
                        </p>
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>
