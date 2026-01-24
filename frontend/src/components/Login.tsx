import useAuthStore from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';

export default function Login() {
    const token = useAuthStore((state) => state.token);
    const navigate = useNavigate();

    const handleLogin = (platform: 'spotify' | 'youtube') => {
        window.location.href = `${import.meta.env.VITE_API_BASE_URL}/login/${platform}`;
    };

    if (token) {
        navigate('/playlists');
    }

    return (
        <div className="flex justify-center items-center h-screen bg-gray-100">
            <div className="p-6 bg-white rounded shadow">
                <h1 className="text-2xl mb-4">Echobridge</h1>
                <button
                    onClick={() => handleLogin('spotify')}
                    className="bg-green-500 text-white px-4 py-2 rounded mr-2 hover:bg-green-600"
                >
                    Login with Spotify
                </button>
                <button
                    onClick={() => handleLogin('youtube')}
                    className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
                >
                    Login with YouTube
                </button>
            </div>
        </div>
    );
}
