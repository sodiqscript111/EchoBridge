import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import useAuthStore from '../context/AuthContext';

export default function Callback() {
    const navigate = useNavigate();
    const location = useLocation();
    const setToken = useAuthStore((state) => state.setToken);

    useEffect(() => {
        const params = new URLSearchParams(location.search);
        const token = params.get('jwt_token');
        if (token) {
            setToken(token);
            navigate('/playlists');
        } else {
            navigate('/login');
        }
    }, [location, setToken, navigate]);

    return <div className="text-center py-16 text-gray-600">Loading...</div>;
}
