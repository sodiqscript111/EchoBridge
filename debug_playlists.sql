-- Query to check all playlists and their public status
SELECT id, title, is_public, platform FROM playlists;

-- Update all playlists to be public (for testing)
UPDATE playlists SET is_public = true;

-- Update specific playlist to be public
-- UPDATE playlists SET is_public = true WHERE id = 'your-playlist-id-here';

-- Count public vs private playlists
SELECT 
    is_public,
    COUNT(*) as count
FROM playlists
GROUP BY is_public;
