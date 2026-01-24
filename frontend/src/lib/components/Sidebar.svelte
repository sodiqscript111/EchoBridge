<script>
    import { Home, Search, Library, Disc, Settings } from "lucide-svelte";
    import { page } from "$app/stores";

    import { isSidebarOpen } from "$lib/stores/ui";

    const links = [
        { icon: Home, label: "Home", href: "/" },
        { icon: Search, label: "Search", href: "/search" },
        { icon: Library, label: "Library", href: "/library" },
        { icon: Settings, label: "Settings", href: "/settings" },
    ];

    const playlists = ["Chill Vibes", "Late Night", "Workout", "Focus Flow"];
</script>

<aside
    class="w-64 h-screen fixed left-0 top-0 border-r border-white/5 bg-black/95 backdrop-blur-xl flex flex-col z-50 transition-transform duration-300 md:translate-x-0 {$isSidebarOpen
        ? 'translate-x-0'
        : '-translate-x-full'}"
>
    <div class="p-6">
        <h1
            class="text-white font-semibold text-xl tracking-tight flex items-center gap-2"
        >
            <Disc class="w-6 h-6 text-white" />
            EchoBridge
        </h1>
    </div>

    <nav class="flex-1 px-4 space-y-1">
        {#each links as link}
            <a
                href={link.href}
                class="flex items-center gap-3 px-4 py-2 rounded-md transition-all duration-200 group
        {$page.url.pathname === link.href
                    ? 'bg-white/10 text-white'
                    : 'text-zinc-400 hover:text-white hover:bg-white/5'}"
            >
                <svelte:component
                    this={link.icon}
                    class="w-5 h-5 {$page.url.pathname === link.href
                        ? 'text-white'
                        : 'group-hover:text-white'}"
                />
                <span class="font-medium text-sm">{link.label}</span>
            </a>
        {/each}

        <div class="pt-8 pb-2 px-4">
            <p
                class="text-xs font-semibold text-zinc-500 uppercase tracking-wider"
            >
                Your Playlists
            </p>
        </div>

        <div class="space-y-1">
            {#each playlists as playlist}
                <a
                    href="#"
                    class="block px-4 py-2 text-sm text-zinc-400 hover:text-white transition-colors"
                >
                    {playlist}
                </a>
            {/each}
        </div>
    </nav>

    <div class="p-4 border-t border-white/5">
        <div class="flex items-center gap-3 px-4 py-2 text-zinc-600 text-xs">
            <span>v1.0.0</span>
        </div>
    </div>
</aside>
