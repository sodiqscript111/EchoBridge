<script>
    import { createEventDispatcher } from "svelte";
    import { Play, Globe, Lock } from "lucide-svelte";

    export let title;
    export let description;
    export let image;
    export let id;
    export let isPublic = false;
    export let showToggle = false;

    const dispatch = createEventDispatcher();

    function handleToggle(e) {
        e.preventDefault();
        e.stopPropagation();
        dispatch("toggle", { id, isPublic });
    }
</script>

<a
    href="/playlist/{id}"
    class="group block bg-white/5 rounded-md p-4 transition-all duration-300 ease-out hover:-translate-y-1 hover:shadow-xl border border-white/5"
>
    <div
        class="relative aspect-square mb-4 rounded-md overflow-hidden shadow-lg"
    >
        <img
            src={image}
            alt={title}
            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
        />
        <div
            class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center backdrop-blur-[2px]"
        >
            <button
                class="w-12 h-12 rounded-full bg-white flex items-center justify-center text-black transform translate-y-4 group-hover:translate-y-0 transition-all duration-300 shadow-lg shadow-white/20 hover:scale-105"
            >
                <Play class="w-5 h-5 fill-current ml-1" />
            </button>
        </div>

        {#if showToggle}
            <button
                on:click={handleToggle}
                class="absolute top-2 right-2 p-2 rounded-full backdrop-blur-md transition-colors z-20
                       {isPublic
                    ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                    : 'bg-black/60 text-white/60 hover:bg-black/80'}"
                title={isPublic ? "Public" : "Private"}
            >
                {#if isPublic}
                    <Globe size={16} />
                {:else}
                    <Lock size={16} />
                {/if}
            </button>
        {/if}
    </div>

    <div class="space-y-1">
        <h3
            class="font-bold text-white truncate tracking-tight transition-colors"
        >
            {title}
        </h3>
        <p class="text-sm text-zinc-400 line-clamp-2 leading-relaxed">
            {description}
        </p>
    </div>
</a>
