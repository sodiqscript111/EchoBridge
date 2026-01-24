import { c as create_ssr_component, v as validate_component, a as subscribe, e as each, b as escape, d as add_attribute, m as missing_component } from "../../chunks/ssr.js";
import { p as page } from "../../chunks/stores.js";
import { w as writable } from "../../chunks/index.js";
import { I as Icon } from "../../chunks/Icon.js";
const Bell = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "path",
      {
        "d": "M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"
      }
    ],
    ["path", { "d": "M10.3 21a1.94 1.94 0 0 0 3.4 0" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "bell" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Disc = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    ["circle", { "cx": "12", "cy": "12", "r": "10" }],
    ["circle", { "cx": "12", "cy": "12", "r": "2" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "disc" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Home = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "path",
      {
        "d": "m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"
      }
    ],
    ["polyline", { "points": "9 22 9 12 15 12 15 22" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "home" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Library = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    ["path", { "d": "m16 6 4 14" }],
    ["path", { "d": "M12 6v14" }],
    ["path", { "d": "M8 8v12" }],
    ["path", { "d": "M4 4v16" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "library" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Menu = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "line",
      {
        "x1": "4",
        "x2": "20",
        "y1": "12",
        "y2": "12"
      }
    ],
    [
      "line",
      {
        "x1": "4",
        "x2": "20",
        "y1": "6",
        "y2": "6"
      }
    ],
    [
      "line",
      {
        "x1": "4",
        "x2": "20",
        "y1": "18",
        "y2": "18"
      }
    ]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "menu" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Search = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    ["circle", { "cx": "11", "cy": "11", "r": "8" }],
    ["path", { "d": "m21 21-4.3-4.3" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "search" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const Settings = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "path",
      {
        "d": "M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"
      }
    ],
    ["circle", { "cx": "12", "cy": "12", "r": "3" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "settings" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const User = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "path",
      {
        "d": "M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"
      }
    ],
    ["circle", { "cx": "12", "cy": "7", "r": "4" }]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "user" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const isSidebarOpen = writable(false);
const Sidebar = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $isSidebarOpen, $$unsubscribe_isSidebarOpen;
  let $page, $$unsubscribe_page;
  $$unsubscribe_isSidebarOpen = subscribe(isSidebarOpen, (value) => $isSidebarOpen = value);
  $$unsubscribe_page = subscribe(page, (value) => $page = value);
  const links = [
    { icon: Home, label: "Home", href: "/" },
    {
      icon: Search,
      label: "Search",
      href: "/search"
    },
    {
      icon: Library,
      label: "Library",
      href: "/library"
    },
    {
      icon: Settings,
      label: "Settings",
      href: "/settings"
    }
  ];
  const playlists = ["Chill Vibes", "Late Night", "Workout", "Focus Flow"];
  $$unsubscribe_isSidebarOpen();
  $$unsubscribe_page();
  return `<aside class="${"w-64 h-screen fixed left-0 top-0 border-r border-white/5 bg-black/95 backdrop-blur-xl flex flex-col z-50 transition-transform duration-300 md:translate-x-0 " + escape($isSidebarOpen ? "translate-x-0" : "-translate-x-full", true)}"><div class="p-6"><h1 class="text-white font-semibold text-xl tracking-tight flex items-center gap-2">${validate_component(Disc, "Disc").$$render($$result, { class: "w-6 h-6 text-white" }, {}, {})}
            EchoBridge</h1></div> <nav class="flex-1 px-4 space-y-1">${each(links, (link) => {
    return `<a${add_attribute("href", link.href, 0)} class="${"flex items-center gap-3 px-4 py-2 rounded-md transition-all duration-200 group " + escape(
      $page.url.pathname === link.href ? "bg-white/10 text-white" : "text-zinc-400 hover:text-white hover:bg-white/5",
      true
    )}">${validate_component(link.icon || missing_component, "svelte:component").$$render(
      $$result,
      {
        class: "w-5 h-5 " + ($page.url.pathname === link.href ? "text-white" : "group-hover:text-white")
      },
      {},
      {}
    )} <span class="font-medium text-sm">${escape(link.label)}</span> </a>`;
  })} <div class="pt-8 pb-2 px-4" data-svelte-h="svelte-ffttul"><p class="text-xs font-semibold text-zinc-500 uppercase tracking-wider">Your Playlists</p></div> <div class="space-y-1">${each(playlists, (playlist) => {
    return `<a href="#" class="block px-4 py-2 text-sm text-zinc-400 hover:text-white transition-colors">${escape(playlist)} </a>`;
  })}</div></nav> <div class="p-4 border-t border-white/5" data-svelte-h="svelte-fbnfwh"><div class="flex items-center gap-3 px-4 py-2 text-zinc-600 text-xs"><span>v1.0.0</span></div></div></aside>`;
});
const Navbar = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  return `<header class="sticky top-0 z-40 w-full glass"><div class="flex items-center justify-between px-8 py-4"> <div class="flex items-center gap-4"><button class="md:hidden text-zinc-400 hover:text-white">${validate_component(Menu, "Menu").$$render($$result, { class: "w-6 h-6" }, {}, {})}</button> <div class="flex items-center gap-2 text-zinc-400 bg-zinc-900/50 px-3 py-1.5 rounded-full border border-white/5 focus-within:border-white/20 focus-within:text-white transition-colors w-64">${validate_component(Search, "Search").$$render($$result, { class: "w-4 h-4" }, {}, {})} <input type="text" placeholder="Search" class="bg-transparent border-none outline-none text-sm w-full placeholder-zinc-500"></div></div>  <div class="flex items-center gap-4"><button class="text-zinc-400 hover:text-white transition-colors">${validate_component(Bell, "Bell").$$render($$result, { class: "w-5 h-5" }, {}, {})}</button> <button class="w-8 h-8 rounded-full bg-zinc-700 flex items-center justify-center text-white font-bold text-xs ring-2 ring-black">${validate_component(User, "User").$$render($$result, { class: "w-4 h-4" }, {}, {})}</button></div></div></header>`;
});
const Layout = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $page, $$unsubscribe_page;
  let $isSidebarOpen, $$unsubscribe_isSidebarOpen;
  $$unsubscribe_page = subscribe(page, (value) => $page = value);
  $$unsubscribe_isSidebarOpen = subscribe(isSidebarOpen, (value) => $isSidebarOpen = value);
  {
    if ($page.url.pathname) {
      isSidebarOpen.set(false);
    }
  }
  $$unsubscribe_page();
  $$unsubscribe_isSidebarOpen();
  return `<div class="min-h-screen bg-black text-zinc-400 selection:bg-white selection:text-black flex">${validate_component(Sidebar, "Sidebar").$$render($$result, {}, {}, {})}  ${$isSidebarOpen ? `<div class="fixed inset-0 bg-black/80 z-40 md:hidden backdrop-blur-sm" role="button" tabindex="0"></div>` : ``} <main class="flex-1 md:ml-64 min-h-screen flex flex-col w-full transition-all duration-300">${validate_component(Navbar, "Navbar").$$render($$result, {}, {}, {})} <div class="flex-1 p-4 md:p-8">${slots.default ? slots.default({}) : ``}</div></main></div>`;
});
export {
  Layout as default
};
