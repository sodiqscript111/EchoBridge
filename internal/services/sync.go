package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"EchoBridge/db"

	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// --- CONFIGURATION (HARDCODED CREDENTIALS) ---
var (
	// SPOTIFY CONFIG
	spotifyAuth = spotifyauth.New(
		spotifyauth.WithClientID("6f55af09467d44319c84abf777042886"),
		spotifyauth.WithClientSecret("879b791b755c4a9cbb97254bee4d3917"),
		spotifyauth.WithRedirectURL("http://127.0.0.1:8000/callback/spotify"),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
		),
	)

	// YOUTUBE CONFIG
	youtubeOAuthConfig = &oauth2.Config{
		ClientID:     "1055231814715-mkfdshbi2l4rkmvn39l4f530o61o6c4u.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://127.0.0.1:8000/callback/youtube",
		Scopes:       []string{youtube.YoutubeReadonlyScope, youtube.YoutubeScope},
		Endpoint:     google.Endpoint,
	}
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

// --- YOUTUBE FUNCTIONS ---

// GetYouTubeClient returns a YouTube client for a user
func GetYouTubeClient(ctx context.Context, user db.User) (*youtube.Service, error) {
	if user.YouTubeToken == "" {
		return nil, fmt.Errorf("no YouTube token available")
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(user.YouTubeToken), &token); err != nil {
		return nil, fmt.Errorf("invalid YouTube token: %w", err)
	}
	// Use TokenSource to handle auto-refreshes
	tokenSource := youtubeOAuthConfig.TokenSource(ctx, &token)
	return youtube.NewService(ctx, option.WithTokenSource(tokenSource))
}

// GetYouTubePlaylists retrieves user playlists
func GetYouTubePlaylists(ctx context.Context, service *youtube.Service) ([]db.Playlist, error) {
	call := service.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve YouTube playlists: %w", err)
	}

	var playlists []db.Playlist
	for _, playlist := range response.Items {
		if playlist.Id != "" && playlist.Snippet != nil && playlist.Snippet.Title != "" {
			var coverImage string
			if playlist.Snippet.Thumbnails != nil && playlist.Snippet.Thumbnails.Default != nil {
				coverImage = playlist.Snippet.Thumbnails.Default.Url
			}
			playlists = append(playlists, db.Playlist{
				ID:          uuid.New(),
				Title:       playlist.Snippet.Title,
				Description: playlist.Snippet.Description,
				Platform:    "youtube",
				SourceID:    playlist.Id,
				IsPublic:    false,
				CoverImage:  coverImage,
				CreatedAt:   time.Now(),
			})
		}
	}
	return playlists, nil
}

// GetYouTubePlaylistTracks retrieves tracks from a YouTube playlist
func GetYouTubePlaylistTracks(ctx context.Context, service *youtube.Service, playlistID string) ([]db.Track, error) {
	var tracks []db.Track
	nextPageToken := ""

	for {
		call := service.PlaylistItems.List([]string{"snippet"}).PlaylistId(playlistID).MaxResults(50).PageToken(nextPageToken)
		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve YouTube playlist items: %w", err)
		}

		for _, item := range response.Items {
			if item.Snippet != nil && item.Snippet.ResourceId != nil && item.Snippet.ResourceId.Kind == "youtube#video" {
				tracks = append(tracks, db.Track{
					ID:        uuid.New(),
					Title:     item.Snippet.Title,
					Artist:    item.Snippet.VideoOwnerChannelTitle, // Best guess for artist
					YouTubeID: item.Snippet.ResourceId.VideoId,
					CreatedAt: time.Now(),
				})
			}
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
	return tracks, nil
}

// ImportToYouTube creates a YouTube playlist and adds tracks
func ImportToYouTube(ctx context.Context, client *youtube.Service, playlist db.Playlist, tracks []db.Track) (string, error) {
	// 1. Create Playlist
	playlistID, err := CreateYouTubePlaylist(ctx, client, playlist.Title, playlist.Description)
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube playlist: %w", err)
	}

	// 2. Add Tracks
	for _, track := range tracks {
		// Search
		videoID, err := SearchYouTubeVideo(ctx, client, track.Title, track.Artist)
		if err != nil || videoID == "" {
			continue // Skip if not found
		}
		// Add
		if err := AddYouTubePlaylistItem(ctx, client, playlistID, videoID); err != nil {
			// Log error but continue
			fmt.Printf("Failed to add track to YouTube: %v\n", err)
		}
		// Rate limiting to avoid API quota issues
		time.Sleep(200 * time.Millisecond)
	}

	return playlistID, nil
}

// CreateYouTubePlaylist creates a new playlist
func CreateYouTubePlaylist(ctx context.Context, service *youtube.Service, title, description string) (string, error) {
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{Title: title, Description: description},
		Status:  &youtube.PlaylistStatus{PrivacyStatus: "private"},
	}
	call := service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube playlist: %w", err)
	}
	return response.Id, nil
}

// AddYouTubePlaylistItem adds a video to a playlist
func AddYouTubePlaylistItem(ctx context.Context, service *youtube.Service, playlistID, videoID string) error {
	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{Kind: "youtube#video", VideoId: videoID},
		},
	}
	call := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err := call.Do()
	return err
}

// SearchYouTubeVideo searches for a video by title and artist
func SearchYouTubeVideo(ctx context.Context, service *youtube.Service, title, artist string) (string, error) {
	searchQuery := fmt.Sprintf("%s %s", title, artist)
	call := service.Search.List([]string{"id"}).Q(searchQuery).MaxResults(1).Type("video")
	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("failed to search YouTube video: %w", err)
	}
	if len(response.Items) > 0 {
		return response.Items[0].Id.VideoId, nil
	}
	return "", nil
}

// --- SPOTIFY FUNCTIONS ---

// GetSpotifyClient returns a Spotify client for a user
func GetSpotifyClient(ctx context.Context, user db.User) (*spotify.Client, error) {
	if user.SpotifyToken == "" {
		return nil, fmt.Errorf("no Spotify token available")
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(user.SpotifyToken), &token); err != nil {
		return nil, fmt.Errorf("invalid Spotify token: %w", err)
	}
	return spotify.New(spotifyAuth.Client(ctx, &token)), nil
}

// GetSpotifyPlaylistTracks retrieves tracks from a Spotify playlist
func GetSpotifyPlaylistTracks(ctx context.Context, client *spotify.Client, playlistID string) ([]db.Track, error) {
	tracksPage, err := client.GetPlaylistTracks(ctx, spotify.ID(playlistID))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Spotify playlist tracks: %w", err)
	}

	var tracks []db.Track
	for _, item := range tracksPage.Tracks {
		if item.Track.ID != "" {
			tracks = append(tracks, db.Track{
				ID:        uuid.New(),
				Title:     item.Track.Name,
				Artist:    strings.Join(GetArtists(item.Track.Artists), ", "),
				Album:     item.Track.Album.Name,
				SpotifyID: item.Track.ID.String(),
				CreatedAt: time.Now(),
			})
		}
	}
	return tracks, nil
}

// ImportToSpotify creates a Spotify playlist and adds tracks
func ImportToSpotify(ctx context.Context, client *spotify.Client, user db.User, playlist db.Playlist, tracks []db.Track) (string, error) {
	spPlaylist, err := client.CreatePlaylistForUser(ctx, user.SpotifyID, playlist.Title, playlist.Description, playlist.IsPublic, false)
	if err != nil {
		return "", fmt.Errorf("failed to create Spotify playlist: %w", err)
	}

	var trackIDs []spotify.ID
	for _, t := range tracks {
		if t.SpotifyID != "" {
			trackIDs = append(trackIDs, spotify.ID(t.SpotifyID))
		}
	}

	// Fix: Chunking (Spotify limit is 100 tracks per request)
	chunkSize := 100
	for i := 0; i < len(trackIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(trackIDs) {
			end = len(trackIDs)
		}
		_, err = client.AddTracksToPlaylist(ctx, spPlaylist.ID, trackIDs[i:end]...)
		if err != nil {
			return "", fmt.Errorf("failed to add tracks to Spotify playlist: %w", err)
		}
	}
	return spPlaylist.ID.String(), nil
}

func GetArtists(artists []spotify.SimpleArtist) []string {
	var names []string
	for _, a := range artists {
		names = append(names, a.Name)
	}
	return names
}
