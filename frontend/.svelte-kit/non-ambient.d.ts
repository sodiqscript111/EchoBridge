
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	export interface AppTypes {
		RouteId(): "/" | "/category" | "/category/[name]" | "/dashboard" | "/library" | "/login" | "/playlist" | "/playlist/[id]" | "/public" | "/settings" | "/signup";
		RouteParams(): {
			"/category/[name]": { name: string };
			"/playlist/[id]": { id: string }
		};
		LayoutParams(): {
			"/": { name?: string; id?: string };
			"/category": { name?: string };
			"/category/[name]": { name: string };
			"/dashboard": Record<string, never>;
			"/library": Record<string, never>;
			"/login": Record<string, never>;
			"/playlist": { id?: string };
			"/playlist/[id]": { id: string };
			"/public": Record<string, never>;
			"/settings": Record<string, never>;
			"/signup": Record<string, never>
		};
		Pathname(): "/" | "/category" | "/category/" | `/category/${string}` & {} | `/category/${string}/` & {} | "/dashboard" | "/dashboard/" | "/library" | "/library/" | "/login" | "/login/" | "/playlist" | "/playlist/" | `/playlist/${string}` & {} | `/playlist/${string}/` & {} | "/public" | "/public/" | "/settings" | "/settings/" | "/signup" | "/signup/";
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): string & {};
	}
}