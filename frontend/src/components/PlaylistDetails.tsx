import { useParams } from 'react-router-dom';

export default function PlaylistDetails() {
    const { id } = useParams<{ id: string }>();

    return (
        <section className="w-full px-6 py-16 bg-white">
            <h2 className="text-3xl font-bold text-center text-gray-800 mb-8">
                Playlist Details
            </h2>
            <div className="max-w-4xl mx-auto">
                <p className="text-gray-600">Playlist ID: {id}</p>
                <p className="text-gray-600">Details page coming soon...</p>
            </div>
        </section>
    );
}
