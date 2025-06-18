export interface Track {
    title: string;
    artist: string;
    album: string;
    spotify_id?: string;
    youtube_id?: string;
}

export interface Playlist {
    id: string;
    title: string;
    description: string;
    platform: 'spotify' | 'youtube';
    owner_id: string;
    tracks: Track[];
    cover_image: string;
}

export interface PublicPlaylistsResponse {
    message: string;
    playlists: Playlist[];
}

export interface AuthState {
    token: string | null;
    setToken: (token: string) => void;
    clearToken: () => void;
}
