<script>
  import { Music, User } from "lucide-svelte";

  export let title = "Playlist Title";
  export let songCount = 0;
  export let artists = "Various Artists";
  export let coverImage = null;
  export let category = "";
  export let color1 = "#ff0080"; // Fallback color
  export let color2 = "#7928ca"; // Fallback color
</script>

<div
  class="relative w-full h-64 rounded-3xl overflow-hidden cursor-pointer group transform transition-all duration-500 hover:scale-[1.02] shadow-2xl shadow-black/50"
>
  <!-- Background: Image or Gradient Fallback -->
  {#if coverImage}
    <img
      src={coverImage}
      alt={title}
      class="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-110"
    />
    <div
      class="absolute inset-0 bg-black/20 group-hover:bg-black/40 transition-colors duration-500"
    ></div>
  {:else}
    <div
      class="absolute inset-0 w-full h-full bg-gradient-to-br from-[var(--c1)] to-[var(--c2)]"
      style="--c1: {color1}; --c2: {color2};"
    ></div>
  {/if}

  <!-- Content Overlay (Glass) -->
  <div
    class="absolute inset-0 p-6 flex flex-col justify-between bg-gradient-to-t from-black/90 via-black/20 to-transparent"
  >
    <!-- Top: Icon -->
    <div class="flex justify-between items-start">
      <div
        class="p-3 bg-white/10 backdrop-blur-md rounded-full border border-white/10"
      >
        <Music size={24} color="white" />
      </div>
      <div class="flex gap-2">
        {#if category}
          <span
            class="px-3 py-1 text-xs font-medium bg-apple-blue/80 backdrop-blur-md rounded-full text-white border border-white/10 uppercase tracking-wider shadow-lg"
          >
            {category}
          </span>
        {/if}
        <span
          class="px-3 py-1 text-xs font-medium bg-black/40 backdrop-blur-md rounded-full text-white/90 border border-white/5"
        >
          {songCount} Songs
        </span>
      </div>
    </div>

    <!-- Bottom: Info -->
    <div>
      <h3
        class="text-2xl font-bold text-white mb-1 tracking-tight group-hover:translate-x-1 transition-transform duration-300 drop-shadow-lg"
      >
        {title}
      </h3>
      <div
        class="flex items-center gap-2 text-white/80 text-sm font-medium drop-shadow-md"
      >
        <User size={14} />
        <p class="truncate max-w-[200px]">{artists}</p>
      </div>
    </div>
  </div>
</div>
