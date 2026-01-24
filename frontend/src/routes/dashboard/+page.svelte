<script>
  import { onMount } from "svelte";
  import { auth } from "$lib/stores/auth";
  import { api } from "$lib/api";
  import { goto } from "$app/navigation";
  import { Music, Globe, Lock } from "lucide-svelte";
  import PlatformCard from "$lib/components/PlatformCard.svelte";
  import PlaylistImport from "$lib/components/PlaylistImport.svelte";
  import PlaylistImportModal from "$lib/components/PlaylistImportModal.svelte";

  let connectionStatus = {
    spotify_connected: false,
    youtube_connected: false,
    username: "",
  };
  let savedPlaylists = [];
  let loading = true;
  let error = "";
  let showImportModal = false;
  let importPlatform = null;

  // Reactive statement to check auth
  $: if (!$auth.token) {
    if (typeof window !== "undefined") goto("/login");
  }

  onMount(async () => {
    if (!$auth.token) return;

    // Check if redirected from OAuth
    const params = new URLSearchParams(window.location.search);
    const connectedPlatform = params.get("connected");
    if (connectedPlatform) {
      // Show import modal after OAuth redirect
      importPlatform = connectedPlatform;
      showImportModal = true;
      // Clean up URL
      window.history.replaceState({}, "", "/dashboard");
    }

    await Promise.all([loadConnectionStatus(), loadSavedPlaylists()]);
    loading = false;
  });

  async function loadConnectionStatus() {
    try {
      const response = await api.get("/api/connection/status", $auth.token);
      connectionStatus = response;
    } catch (e) {
      console.error("Failed to load connection status:", e);
      error = "Connection status error: " + e.message;
    }
  }

  async function loadSavedPlaylists() {
    try {
      const response = await api.get("/api/my/playlists", $auth.token);
      savedPlaylists = response.playlists || [];
    } catch (e) {
      console.error("Failed to load saved playlists:", e);
      error = "Saved playlists error: " + e.message;
    }
  }

  function handleConnectionSuccess(platform) {
    // Show import modal after successful connection
    importPlatform = platform;
    showImportModal = true;
    // Reload connection status
    loadConnectionStatus();
  }

  function onImportComplete() {
    // Reload playlists after import
    loadSavedPlaylists();
  }

  async function togglePublic(playlistId, currentStatus) {
    try {
      await api.patch(
        `/api/playlists/${playlistId}/public`,
        { is_public: !currentStatus },
        $auth.token,
      );

      savedPlaylists = savedPlaylists.map((p) =>
        p.id === playlistId ? { ...p, is_public: !currentStatus } : p,
      );
    } catch (e) {
      alert("Failed to update playlist: " + e.message);
    }
  }
</script>

<div class="p-8 pt-24 max-w-7xl mx-auto">
  <!-- Header -->
  <div class="mb-12">
    <h1 class="text-5xl font-bold mb-2">Dashboard</h1>
    <p class="text-xl text-white/60">
      Welcome back{connectionStatus.username
        ? `, ${connectionStatus.username}`
        : ""}!
    </p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div
        class="w-12 h-12 border-4 border-white/20 border-t-apple-blue rounded-full animate-spin"
      ></div>
    </div>
  {:else}
    <!-- Platform Connections Section -->
    <section class="mb-12">
      <h2 class="text-2xl font-bold mb-6">Platform Connections</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <PlatformCard
          platform="spotify"
          connected={connectionStatus.spotify_connected}
          color="#1DB954"
          onConnectionSuccess={() => handleConnectionSuccess("spotify")}
        />
        <PlatformCard
          platform="youtube"
          connected={connectionStatus.youtube_connected}
          color="#FF0000"
          onConnectionSuccess={() => handleConnectionSuccess("youtube")}
        />
      </div>
    </section>

    <!-- Available Playlists Section -->
    {#if connectionStatus.spotify_connected || connectionStatus.youtube_connected}
      <section class="mb-12">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h2 class="text-2xl font-bold">Import Playlists</h2>
            <p class="text-sm text-white/60 mt-1">
              Import playlists from your connected platforms
            </p>
          </div>
        </div>

        <PlaylistImport on:import={loadSavedPlaylists} />
      </section>
    {/if}

    <!-- Saved Playlists Section -->
    <section>
      <h2 class="text-2xl font-bold mb-6">My Saved Playlists</h2>

      {#if error}
        <div class="glass rounded-2xl p-8 text-center text-red-400">
          {error}
        </div>
      {:else if savedPlaylists.length === 0}
        <div class="glass rounded-3xl p-12 text-center">
          <Music class="mx-auto mb-4 opacity-30" size={48} />
          <h3 class="text-xl font-bold mb-2">No Saved Playlists Yet</h3>
          <p class="text-white/60 mb-6">
            Import your first playlist to get started!
          </p>
          <a
            href="/public"
            class="inline-block px-6 py-3 bg-apple-blue rounded-full font-bold hover:bg-blue-600 transition-colors"
          >
            Browse Public Playlists
          </a>
        </div>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {#each savedPlaylists as playlist}
            <div
              class="glass rounded-2xl p-6 hover:bg-white/10 transition-colors"
            >
              <div class="flex items-start justify-between mb-4">
                <div class="flex-1">
                  <h3 class="font-bold text-lg mb-1">{playlist.title}</h3>
                  <p class="text-sm text-white/50 capitalize">
                    {playlist.platform}
                  </p>
                </div>
                {#if playlist.cover_image}
                  <img
                    src={playlist.cover_image}
                    alt=""
                    class="w-12 h-12 rounded-lg object-cover"
                  />
                {/if}
              </div>

              <p class="text-sm text-white/60 mb-4 line-clamp-2">
                {playlist.description || "No description"}
              </p>

              <div class="flex items-center justify-between">
                <button
                  on:click={() => togglePublic(playlist.id, playlist.is_public)}
                  class="flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium transition-colors {playlist.is_public
                    ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                    : 'bg-white/5 text-white/60 hover:bg-white/10'}"
                >
                  {#if playlist.is_public}
                    <Globe size={14} />
                    Public
                  {:else}
                    <Lock size={14} />
                    Private
                  {/if}
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  {/if}
</div>

<!-- Import Modal -->
{#if showImportModal}
  <PlaylistImportModal
    platform={importPlatform}
    onClose={() => (showImportModal = false)}
    {onImportComplete}
  />
{/if}

<style>
  .glass {
    background: rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
