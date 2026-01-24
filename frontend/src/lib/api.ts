import { get } from 'svelte/store';
import { auth } from '$lib/stores/auth';

// Use environment variable for API base, or empty string for relative path (same origin)
const API_BASE = import.meta.env.VITE_API_BASE_URL || '';

async function request(endpoint: string, method: string = 'GET', data: any = null) {
    const token = get(auth).token;
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const config: RequestInit = {
        method,
        headers,
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
            // If response is not JSON, use text as error or empty object
            console.warn(`[API] Response was not JSON: ${text.substring(0, 100)}...`);
            json = { error: text || response.statusText };
        }

        if (!response.ok) {
            const error = new Error(json.error || 'Request failed');
            (error as any).code = json.code;
            (error as any).details = json;
            throw error;
        }

        return json;
    } catch (error) {
        console.error(`[API] ${method} ${endpoint} failed:`, error);
        throw error;
    }
}

export const api = {
    get: (endpoint: string) => request(endpoint, 'GET'),
    post: (endpoint: string, data: any) => request(endpoint, 'POST', data),
    patch: (endpoint: string, data: any) => request(endpoint, 'PATCH', data),

    // Specific methods
    getPlaylists: () => request('/api/my/playlists'),
    getPublicPlaylists: () => request('/playlists/public'),
    getPlaylist: (id: string) => request(`/playlists/${id}`),
    importPlaylist: (id: string, platform: 'spotify' | 'youtube') => request(`/api/playlists/${id}/import`, 'POST', { platform }),
    getConnectionStatus: () => request('/api/connection/status'),

    // Auth links (these return URLs, not API calls)
    getSpotifyLink: () => `${API_BASE}/api/link/spotify?token=${get(auth).token}`,
    getYouTubeLink: () => `${API_BASE}/api/link/youtube?token=${get(auth).token}`,
};
