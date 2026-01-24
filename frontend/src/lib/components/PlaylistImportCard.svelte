<script>
    import { createEventDispatcher } from "svelte";
    import { Music } from "lucide-svelte";

    export let playlist = {};
    export let importing = false;
    export let alreadyImported = false;

    const dispatch = createEventDispatcher();

    let isPublic = false;

    function handleImport() {
        dispatch("import", {
            platform: playlist.platform,
            source_id: playlist.source_id,
            is_public: isPublic,
        });
    }

    function getPlatformColor(platform) {
        return platform === "spotify" ? "#1DB954" : "#FF0000";
    }
</script>

<div
    class="glass rounded-2xl overflow-hidden hover:scale-[1.02] transition-all duration-300"
>
    <!-- Cover Image -->
    <div class="relative h-48 bg-zinc-800">
        {#if playlist.cover_image}
            <img
                src={playlist.cover_image}
                alt={playlist.title}
                class="w-full h-full object-cover"
            />
        {:else}
            <div class="w-full h-full flex items-center justify-center">
                <Music size={64} class="opacity-20" />
            </div>
        {/if}

        <!-- Platform Badge -->
        <div
            class="absolute top-3 right-3 px-3 py-1 rounded-full text-xs font-bold uppercase backdrop-blur-md"
            style="background-color: {getPlatformColor(
                playlist.platform,
            )}40; color: {getPlatformColor(
                playlist.platform,
            )}; border: 1px solid {getPlatformColor(playlist.platform)}60"
        >
            {playlist.platform}
        </div>
    </div>

    <!-- Content -->
    <div class="p-5">
        <h3 class="font-bold text-lg mb-1 line-clamp-1">{playlist.title}</h3>
        <p class="text-sm text-white/60 mb-3 line-clamp-2">
            {playlist.description || "No description"}
        </p>

        <div class="flex items-center gap-2 text-sm text-white/50 mb-4">
            <Music size={14} />
            <span>{playlist.track_count || 0} tracks</span>
        </div>

        {#if alreadyImported}
            <div
                class="px-4 py-2 rounded-full bg-green-500/20 text-green-400 text-sm font-medium text-center"
            >
                Already Imported
            </div>
        {:else}
            <!-- Public/Private Toggle -->
            <div class="mb-3">
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="checkbox"
                        bind:checked={isPublic}
                        class="w-4 h-4 rounded accent-apple-blue"
                    />
                    <span class="text-sm text-white/70">Make Public</span>
                </label>
            </div>

            <!-- Import Button -->
            <button
                on:click={handleImport}
                disabled={importing}
                class="w-full py-3 rounded-xl font-bold transition-all duration-200 {importing
                    ? 'opacity-50 cursor-not-allowed'
                    : 'hover:scale-105'}"
                style="background-color: {getPlatformColor(playlist.platform)}"
            >
                {#if importing}
                    <span class="flex items-center justify-center gap-2">
                        <div
                            class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
                        ></div>
                        Importing...
                    </span>
                {:else}
                    Import Playlist
                {/if}
            </button>
        {/if}
    </div>
</div>

<style>
    .glass {
        background: rgba(255, 255, 255, 0.05);
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    .line-clamp-1 {
        display: -webkit-box;
        -webkit-line-clamp: 1;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .line-clamp-2 {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
</style>
