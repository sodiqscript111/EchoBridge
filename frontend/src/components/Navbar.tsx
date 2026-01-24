import { Link, useNavigate } from 'react-router-dom';
import useAuthStore from '../context/AuthContext';

export default function Navbar() {
    const { token, clearToken } = useAuthStore();
    const navigate = useNavigate();

    const handleLogout = () => {
        clearToken();
        navigate('/login');
    };

    return (
        <nav className="w-full bg-gray-800 text-white p-4 shadow-md">
            <div className="max-w-6xl mx-auto flex justify-between items-center">
                <Link to="/" className="text-xl font-bold">
                    Echobridge
                </Link>
                <div className="space-x-4">
                    <Link to="/" className="hover:text-gray-300">
                        Public Playlists
                    </Link>
                    {token ? (
                        <>
                            <Link to="/playlists" className="hover:text-gray-300">
                                My Playlists
                            </Link>
                            <button
                                onClick={handleLogout}
                                className="hover:text-gray-300"
                            >
                                Logout
                            </button>
                        </>
                    ) : (
                        <Link to="/login" className="hover:text-gray-300">
                            Login
                        </Link>
                    )}
                </div>
            </div>
        </nav>
    );
}
