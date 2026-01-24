<script>
    import { auth } from "$lib/stores/auth";
    import { api } from "$lib/api";

    export let platform = "spotify";
    export let connected = false;
    export let color = "#1DB954";

    function handleConnect() {
        try {
            const url =
                platform === "spotify"
                    ? api.getSpotifyLink()
                    : api.getYouTubeLink();
            window.location.href = url;
        } catch (e) {
            console.error(`Failed to connect ${platform}:`, e);
            alert(`Failed to connect ${platform}: ` + e.message);
        }
    }
</script>

<div
    class="glass rounded-2xl p-6 hover:bg-white/10 transition-all duration-300 flex flex-col justify-between h-full"
    style="border-left: 4px solid {color}"
>
    <div>
        <h3 class="text-xl font-bold capitalize mb-2">{platform}</h3>
        <p class="text-sm text-white/60 mb-6">
            {#if connected}
                You are connected to {platform}. Your playlists will be synced
                automatically.
            {:else}
                Connect to {platform} to start syncing your playlists and sharing
                music.
            {/if}
        </p>
    </div>

    <div>
        {#if connected}
            <div
                class="px-4 py-2 rounded-full font-medium text-sm inline-flex"
                style="background-color: {color}20; color: {color}"
            >
                <span class="flex items-center gap-2">
                    <span
                        class="w-2 h-2 rounded-full"
                        style="background-color: {color}"
                    ></span>
                    Connected
                </span>
            </div>
        {:else}
            <button
                on:click={handleConnect}
                class="px-6 py-2 rounded-full font-medium text-sm transition-all duration-200 hover:scale-105 w-full"
                style="background-color: {color}; color: white"
            >
                Connect {platform}
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
</style>
