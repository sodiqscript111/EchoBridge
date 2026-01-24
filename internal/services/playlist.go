package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"EchoBridge/db"

	"github.com/google/uuid"
)

// --- MAIN SYNC LOGIC ---

// SyncPlaylist syncs a playlist to specified platforms
func SyncPlaylist(ctx context.Context, user db.User, playlistID uuid.UUID, platforms []string) (map[string]string, error) {
	log.Printf("🚀 Starting Sync for Playlist ID: %s", playlistID)

	var playlist db.Playlist
	if err := db.DB.Where("id = ? AND owner_id = ?", playlistID, user.ID).First(&playlist).Error; err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}

	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch tracks: %w", err)
	}

	log.Printf("📊 Database Check: Found %d tracks for playlist '%s'", len(tracks), playlist.Title)
	if len(tracks) == 0 {
		log.Println("⚠️ WARNING: No tracks found in DB! Ensure the Import step worked correctly.")
		return nil, nil
	}

	result := make(map[string]string)
	for _, platform := range platforms {
		switch platform {
		case "spotify":
			if user.SpotifyToken != "" {
				log.Println("🎵 Starting Spotify Sync...")
				client, err := GetSpotifyClient(ctx, user)
				if err != nil {
					return nil, fmt.Errorf("failed to get Spotify client: %w", err)
				}
				spotifyID, err := ImportToSpotify(ctx, client, user, playlist, tracks)
				if err != nil {
					return nil, fmt.Errorf("failed to sync to Spotify: %w", err)
				}
				result["spotify"] = spotifyID
			}
		case "youtube":
			if user.YouTubeToken != "" {
				log.Println("🎥 Starting YouTube Sync...")
				client, err := GetYouTubeClient(ctx, user)
				if err != nil {
					return nil, fmt.Errorf("failed to get YouTube client: %w", err)
				}

				log.Println("🔨 Creating YouTube Playlist...")
				ytPlaylistID, err := CreateYouTubePlaylist(ctx, client, playlist.Title, playlist.Description)
				if err != nil {
					return nil, fmt.Errorf("failed to create YouTube playlist: %w", err)
				}

				successCount := 0
				for _, track := range tracks {
					videoID, err := SearchYouTubeVideo(ctx, client, track.Title, track.Artist)
					if err != nil || videoID == "" {
						log.Printf("   ⚠️ Failed to find track: %s - %s", track.Title, track.Artist)
						continue
					}
					if err := AddYouTubePlaylistItem(ctx, client, ytPlaylistID, videoID); err != nil {
						log.Printf("   ❌ Failed to add track: %s", track.Title)
					} else {
						log.Printf("   ✅ Added: %s", track.Title)
						successCount++
					}
					// Sleep to respect API rate limits
					time.Sleep(200 * time.Millisecond)
				}
				log.Printf("🏁 YouTube Sync Finished. Added %d/%d songs.", successCount, len(tracks))
				result["youtube"] = ytPlaylistID
			}
		}
	}
	return result, nil
}
