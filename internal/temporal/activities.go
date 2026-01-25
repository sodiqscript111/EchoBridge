package temporal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/services"

	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
)

func FetchUserActivity(ctx context.Context, userID uuid.UUID) (*db.User, error) {
	var user db.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func FetchPlaylistActivity(ctx context.Context, playlistID, userID uuid.UUID) (*db.Playlist, error) {
	var playlist db.Playlist
	if err := db.DB.Where("id = ? AND owner_id = ?", playlistID, userID).First(&playlist).Error; err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}
	return &playlist, nil
}

func FetchPlaylistTracksActivity(ctx context.Context, playlistID uuid.UUID) ([]db.Track, error) {
	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch tracks: %w", err)
	}
	return tracks, nil
}

func CreateSpotifyPlaylistActivity(ctx context.Context, user db.User, playlist db.Playlist) (string, error) {
	client, err := services.GetSpotifyClient(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to get Spotify client: %w", err)
	}

	spPlaylist, err := client.CreatePlaylistForUser(ctx, user.SpotifyID, playlist.Title, playlist.Description, playlist.IsPublic, false)
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			time.Sleep(30 * time.Second)
			return "", fmt.Errorf("rate limited, will retry: %w", err)
		}
		return "", fmt.Errorf("failed to create Spotify playlist: %w", err)
	}
	return spPlaylist.ID.String(), nil
}

func SearchSpotifyTrackActivity(ctx context.Context, user db.User, title, artist string) (string, error) {
	client, err := services.GetSpotifyClient(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to get Spotify client: %w", err)
	}

	query := fmt.Sprintf("track:%s artist:%s", title, artist)
	results, err := client.Search(ctx, query, spotify.SearchTypeTrack, spotify.Limit(1))
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			time.Sleep(30 * time.Second)
			return "", fmt.Errorf("rate limited, will retry: %w", err)
		}
		return "", fmt.Errorf("search failed: %w", err)
	}

	if results.Tracks != nil && len(results.Tracks.Tracks) > 0 {
		return results.Tracks.Tracks[0].ID.String(), nil
	}
	return "", nil
}

func AddTrackToSpotifyActivity(ctx context.Context, user db.User, playlistID, trackID string) error {
	client, err := services.GetSpotifyClient(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to get Spotify client: %w", err)
	}

	_, err = client.AddTracksToPlaylist(ctx, spotify.ID(playlistID), spotify.ID(trackID))
	if err != nil {
		if strings.Contains(err.Error(), "429") {
			time.Sleep(30 * time.Second)
			return fmt.Errorf("rate limited, will retry: %w", err)
		}
		return fmt.Errorf("failed to add track: %w", err)
	}
	return nil
}

func CreateYouTubePlaylistActivity(ctx context.Context, user db.User, playlist db.Playlist) (string, error) {
	client, err := services.GetYouTubeClient(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to get YouTube client: %w", err)
	}

	playlistID, err := services.CreateYouTubePlaylist(ctx, client, playlist.Title, playlist.Description)
	if err != nil {
		if strings.Contains(err.Error(), "quotaExceeded") || strings.Contains(err.Error(), "429") {
			time.Sleep(60 * time.Second)
			return "", fmt.Errorf("quota exceeded, will retry: %w", err)
		}
		return "", fmt.Errorf("failed to create YouTube playlist: %w", err)
	}
	return playlistID, nil
}

func SearchYouTubeVideoActivity(ctx context.Context, user db.User, title, artist string) (string, error) {
	client, err := services.GetYouTubeClient(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to get YouTube client: %w", err)
	}

	videoID, err := services.SearchYouTubeVideo(ctx, client, title, artist)
	if err != nil {
		if strings.Contains(err.Error(), "quotaExceeded") || strings.Contains(err.Error(), "429") {
			time.Sleep(60 * time.Second)
			return "", fmt.Errorf("quota exceeded, will retry: %w", err)
		}
		return "", fmt.Errorf("search failed: %w", err)
	}
	return videoID, nil
}

func AddVideoToYouTubePlaylistActivity(ctx context.Context, user db.User, playlistID, videoID string) error {
	client, err := services.GetYouTubeClient(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to get YouTube client: %w", err)
	}

	err = services.AddYouTubePlaylistItem(ctx, client, playlistID, videoID)
	if err != nil {
		if strings.Contains(err.Error(), "quotaExceeded") || strings.Contains(err.Error(), "429") {
			time.Sleep(60 * time.Second)
			return fmt.Errorf("quota exceeded, will retry: %w", err)
		}
		return fmt.Errorf("failed to add video: %w", err)
	}
	return nil
}

func ImportSpotifyPlaylistActivity(ctx context.Context, user db.User, sourceID string) error {
	client, err := services.GetSpotifyClient(ctx, user)
	if err != nil {
		return err
	}
	return services.ImportSpotifyPlaylist(ctx, client, user, sourceID)
}

func ImportYouTubePlaylistActivity(ctx context.Context, user db.User, sourceID string) error {
	client, err := services.GetYouTubeClient(ctx, user)
	if err != nil {
		return err
	}
	return services.ImportYouTubePlaylist(ctx, client, user, sourceID)
}
