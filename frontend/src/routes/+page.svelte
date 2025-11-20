<script>
  import { ArrowRight } from 'lucide-svelte';
</script>

<div class="flex flex-col items-center justify-center min-h-[80vh] text-center px-4">
  <h1 class="text-6xl md:text-8xl font-bold tracking-tighter mb-6 bg-clip-text text-transparent bg-gradient-to-b from-white to-white/40">
    Sync Your World.
  </h1>
  <p class="text-xl text-white/60 max-w-2xl mb-10 font-light">
    Seamlessly transfer and sync your playlists between Spotify and YouTube. 
    Share your taste with the world.
  </p>
  
  <div class="flex gap-4">
    <a href="/public" class="px-8 py-4 bg-white text-black rounded-full font-bold hover:bg-gray-200 transition-colors flex items-center gap-2">
      Explore Public Mixes <ArrowRight size={20} />
    </a>
    <button class="px-8 py-4 glass rounded-full font-bold hover:bg-white/10 transition-colors">
      Sign In
    </button>
  </div>
</div>
