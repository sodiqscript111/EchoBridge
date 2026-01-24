import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

import { cn } from '../lib/utils';
import type { Playlist, PublicPlaylistsResponse } from '../types';

const gradientPresets = [
    'from-pink-500 via-red-400 to-yellow-300',
    'from-green-400 via-lime-500 to-emerald-500',
    'from-purple-500 via-indigo-500 to-blue-400',
    'from-orange-400 via-pink-500 to-red-500',
    'from-gray-700 via-gray-900 to-black',
    'from-yellow-400 via-orange-500 to-rose-500',
];

export default function PublicPlaylists() {
    const [playlists, setPlaylists] = useState<Playlist[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const navigate = useNavigate();
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const fetchPlaylists = async () => {
            try {
                const response = await axios.get<PublicPlaylistsResponse>(
                    `${import.meta.env.VITE_API_BASE_URL}/playlists/public`
                );
                setPlaylists(response.data.playlists || []);
            } catch (err) {
                setError('Failed to load playlists');
            } finally {
                setLoading(false);
            }
        };
        fetchPlaylists();
    }, []);

    if (loading) return <div className="text-center py-16 text-gray-600">Loading...</div>;
    if (error) return <div className="text-center py-16 text-red-500">{error}</div>;

    return (
        <section ref={containerRef} className="w-full px-6 py-16 bg-white">
            <h2 className="text-3xl font-bold text-center text-gray-800 mb-10">Public Playlists</h2>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 max-w-6xl mx-auto">
                {playlists.map((playlist, index) => {
                    const gradient = gradientPresets[index % gradientPresets.length]; // fallback-safe
                    return (
                        <div
                            key={playlist.id}
                            className={cn(
                                'rounded-2xl overflow-hidden h-[240px] flex flex-col justify-between text-white shadow-xl bg-gradient-to-br animate-gradient-move',
                                gradient
                            )}
                        >
                            <div className="p-4">
                                <h3 className="text-lg sm:text-xl font-bold tracking-tight">{playlist.title}</h3>
                                <p className="text-sm sm:text-base opacity-90 mt-2">
                                    {playlist.tracks?.length || 14} handpicked tracks
                                </p>
                            </div>

                            <div className="p-4">
                                {/* Button wrapper to ensure visibility */}
                                <div className="w-full flex">
                                    <button
                                        onClick={() => navigate(`/playlist/${playlist.id}`)}
                                        className="bg-white/20 backdrop-blur-md border border-white/30 text-white px-5 py-2 rounded-full font-semibold text-sm hover:bg-white/30 transition duration-300 shadow-md"
                                    >
                                        Explore Playlist
                                    </button>
                                </div>

                            </div>
                        </div>
                    );
                })}
            </div>

            {/* Animation for gradient movement */}
            <style jsx>{`
        .animate-gradient-move {
          background-size: 200% 200%;
          animation: gradientShift 6s ease infinite;
        }

        @keyframes gradientShift {
          0% {
            background-position: 0% 50%;
          }
          50% {
            background-position: 100% 50%;
          }
          100% {
            background-position: 0% 50%;
          }
        }
      `}</style>
        </section>
    );
}
