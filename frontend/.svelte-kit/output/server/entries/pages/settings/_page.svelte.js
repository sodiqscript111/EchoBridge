import { c as create_ssr_component, v as validate_component, f as get_store_value, a as subscribe, b as escape } from "../../../chunks/ssr.js";
import { a as auth } from "../../../chunks/auth.js";
/* empty css                                                                */
import { P as Plus } from "../../../chunks/plus.js";
import { I as Icon } from "../../../chunks/Icon.js";
const Check = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [["path", { "d": "M20 6 9 17l-5-5" }]];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "check" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const External_link = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [
    [
      "path",
      {
        "d": "M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
      }
    ],
    ["polyline", { "points": "15 3 21 3 21 9" }],
    [
      "line",
      {
        "x1": "10",
        "x2": "21",
        "y1": "14",
        "y2": "3"
      }
    ]
  ];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "external-link" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const X = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  const iconNode = [["path", { "d": "M18 6 6 18" }], ["path", { "d": "m6 6 12 12" }]];
  return `${validate_component(Icon, "Icon").$$render($$result, Object.assign({}, { name: "x" }, $$props, { iconNode }), {}, {
    default: () => {
      return `${slots.default ? slots.default({}) : ``}`;
    }
  })}`;
});
const API_BASE = "http://localhost:8000";
async function request(endpoint, method = "GET", data = null) {
  const token = get_store_value(auth).token;
  const headers = {
    "Content-Type": "application/json"
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  const config = {
    method,
    headers
  };
  if (data) {
    config.body = JSON.stringify(data);
  }
  try {
    const response = await fetch(`${API_BASE}${endpoint}`, config);
    let json;
    const text = await response.text();
    try {
      json = JSON.parse(text);
    } catch (e) {
      console.warn(`[API] Response was not JSON: ${text.substring(0, 100)}...`);
      json = { error: text || response.statusText };
    }
    if (!response.ok) {
      const error = new Error(json.error || "Request failed");
      error.code = json.code;
      error.details = json;
      throw error;
    }
    return json;
  } catch (error) {
    console.error(`[API] ${method} ${endpoint} failed:`, error);
    throw error;
  }
}
const api = {
  get: (endpoint) => request(endpoint, "GET"),
  post: (endpoint, data) => request(endpoint, "POST", data),
  patch: (endpoint, data) => request(endpoint, "PATCH", data),
  // Specific methods
  getPlaylists: () => request("/api/my/playlists"),
  getPublicPlaylists: () => request("/playlists/public"),
  getPlaylist: (id) => request(`/playlists/${id}`),
  importPlaylist: (id, platform) => request(`/api/playlists/${id}/import`, "POST", { platform }),
  getConnectionStatus: () => request("/api/connection/status"),
  // Auth links (these return URLs, not API calls)
  getSpotifyLink: () => `${API_BASE}/api/link/spotify?token=${get_store_value(auth).token}`,
  getYouTubeLink: () => `${API_BASE}/api/link/youtube?token=${get_store_value(auth).token}`
};
const Page = create_ssr_component(($$result, $$props, $$bindings, slots) => {
  let $auth, $$unsubscribe_auth;
  $$unsubscribe_auth = subscribe(auth, (value) => $auth = value);
  let status = {
    spotify_connected: false,
    youtube_connected: false,
    username: ""
  };
  async function fetchStatus() {
    try {
      status = await api.getConnectionStatus();
    } catch (e) {
      console.error("Failed to fetch status", e);
    } finally {
    }
  }
  {
    if ($auth.token) {
      fetchStatus();
    }
  }
  $$unsubscribe_auth();
  return `<div class="max-w-4xl mx-auto space-y-8"><header class="flex items-center justify-between"><div data-svelte-h="svelte-jue2fr"><h1 class="text-3xl font-bold text-white tracking-tight mb-2">Settings</h1> <p class="text-zinc-400">Manage your connected accounts and preferences.</p></div> <button class="flex items-center gap-2 px-4 py-2 bg-white text-black rounded-full font-bold hover:scale-105 transition-transform">${validate_component(Plus, "Plus").$$render($$result, { size: 20 }, {}, {})}
            Import Playlist</button></header> ${``} <section class="space-y-6"><h2 class="text-xl font-semibold text-white" data-svelte-h="svelte-1tdyeqw">Connected Platforms</h2> <div class="grid grid-cols-1 md:grid-cols-2 gap-6"> <div class="p-6 rounded-xl border border-white/5 bg-white/5 backdrop-blur-sm flex flex-col justify-between h-48"><div class="flex justify-between items-start"><div data-svelte-h="svelte-nw587n"><h3 class="text-lg font-bold text-white mb-1">Spotify</h3> <p class="text-sm text-zinc-400">Import and sync playlists.</p></div> ${status.spotify_connected ? `<div class="bg-green-500/20 text-green-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1">${validate_component(Check, "Check").$$render($$result, { class: "w-3 h-3" }, {}, {})} Connected</div>` : `<div class="bg-zinc-500/20 text-zinc-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1">${validate_component(X, "X").$$render($$result, { class: "w-3 h-3" }, {}, {})} Not Connected</div>`}</div> <button ${status.spotify_connected ? "disabled" : ""} class="${"w-full py-3 rounded-lg font-semibold text-sm transition-all duration-200 flex items-center justify-center gap-2 " + escape(
    status.spotify_connected ? "bg-white/5 text-zinc-500 cursor-default" : "bg-[#1DB954] text-black hover:scale-[1.02] shadow-lg shadow-green-900/20",
    true
  )}">${status.spotify_connected ? `Account Linked` : `Connect Spotify ${validate_component(External_link, "ExternalLink").$$render($$result, { class: "w-4 h-4" }, {}, {})}`}</button></div>  <div class="p-6 rounded-xl border border-white/5 bg-white/5 backdrop-blur-sm flex flex-col justify-between h-48"><div class="flex justify-between items-start"><div data-svelte-h="svelte-eispci"><h3 class="text-lg font-bold text-white mb-1">YouTube Music</h3> <p class="text-sm text-zinc-400">Sync to your Google account.</p></div> ${status.youtube_connected ? `<div class="bg-green-500/20 text-green-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1">${validate_component(Check, "Check").$$render($$result, { class: "w-3 h-3" }, {}, {})} Connected</div>` : `<div class="bg-zinc-500/20 text-zinc-400 px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1">${validate_component(X, "X").$$render($$result, { class: "w-3 h-3" }, {}, {})} Not Connected</div>`}</div> <button ${status.youtube_connected ? "disabled" : ""} class="${"w-full py-3 rounded-lg font-semibold text-sm transition-all duration-200 flex items-center justify-center gap-2 " + escape(
    status.youtube_connected ? "bg-white/5 text-zinc-500 cursor-default" : "bg-[#FF0000] text-white hover:scale-[1.02] shadow-lg shadow-red-900/20",
    true
  )}">${status.youtube_connected ? `Account Linked` : `Connect YouTube ${validate_component(External_link, "ExternalLink").$$render($$result, { class: "w-4 h-4" }, {}, {})}`}</button></div></div></section></div>`;
});
export {
  Page as default
};
