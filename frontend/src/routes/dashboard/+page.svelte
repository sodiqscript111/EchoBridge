<script>
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { api } from '$lib/api';
  import { goto } from '$app/navigation';
  import { Music, Globe, GlobeLock, Trash2 } from 'lucide-svelte';

  let playlists = [];
  let loading = true;
  let error = '';

  // Reactive statement to check auth
  $: if (!$auth.token) {
    if (typeof window !== 'undefined') goto('/login');
  }

  onMount(async () => {
    // Wait for auth to be ready
    if (!$auth.token) return;

    try {
      const response = await api.get('/api/my/playlists', $auth.token);
      playlists = response.playlists || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  async function togglePublic(playlistId, currentStatus) {
    try {
      await api.patch(
        `/api/playlists/${playlistId}/public`,
        { is_public: !currentStatus },
        $auth.token
      );

      // Update local state
      playlists = playlists.map(p =>
        p.id === playlistId ? { ...p, is_public: !currentStatus } : p
      );
    } catch (e) {
      alert('Failed to update playlist: ' + e.message);
    }
  }
</script>

<div class="p-8 pt-24 max-w-7xl mx-auto">
  <div class="mb-12">
    <h1 class="text-5xl font-bold mb-2">My Playlists</h1>
    <p class="text-xl text-white/60">Manage your synced collections</p>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="w-12 h-12 border-4 border-white/20 border-t-apple-blue rounded-full animate-spin"></div>
    </div>
  {:else if error}
    <div class="glass rounded-2xl p-8 text-center text-red-400">
      {error}
    </div>
  {:else if playlists.length === 0}
    <div class="glass rounded-3xl p-12 text-center">
      <Music class="mx-auto mb-4 opacity-30" size={48} />
      <h3 class="text-xl font-bold mb-2">No Playlists Yet</h3>
      <p class="text-white/60 mb-6">Import your first playlist to get started!</p>
      <a href="/public" class="inline-block px-6 py-3 bg-apple-blue rounded-full font-bold hover:bg-blue-600 transition-colors">
        Browse Public Playlists
      </a>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#each playlists as playlist}
        <div class="glass rounded-2xl p-6 hover:bg-white/10 transition-colors">
          <div class="flex items-start justify-between mb-4">
            <div class="flex-1">
              <h3 class="font-bold text-lg mb-1">{playlist.title}</h3>
              <p class="text-sm text-white/50 capitalize">{playlist.platform}</p>
            </div>
            {#if playlist.cover_image}
              <img src={playlist.cover_image} alt="" class="w-12 h-12 rounded-lg object-cover" />
            {/if}
          </div>

          <p class="text-sm text-white/60 mb-4 line-clamp-2">
            {playlist.description || 'No description'}
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
                <GlobeLock size={14} />
                Private
              {/if}
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
