package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"EchoBridge/db"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// --- CONFIGURATION ---
func getYouTubeOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")
	redirectURL := os.Getenv("YOUTUBE_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://127.0.0.1:8000/callback/youtube"
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{youtube.YoutubeReadonlyScope, youtube.YoutubeScope},
		Endpoint:     google.Endpoint,
	}
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
	youtubeOAuthConfig := getYouTubeOAuthConfig()
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
					Artist:    item.Snippet.VideoOwnerChannelTitle,
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
	fmt.Printf("Creating YouTube playlist: %s\n", playlist.Title)
	playlistID, err := CreateYouTubePlaylist(ctx, client, playlist.Title, playlist.Description)
	if err != nil {
		fmt.Printf("Failed to create YouTube playlist: %v\n", err)
		return "", fmt.Errorf("failed to create YouTube playlist: %w", err)
	}
	fmt.Printf("Created YouTube playlist ID: %s\n", playlistID)

	for _, track := range tracks {
		videoID, err := SearchYouTubeVideo(ctx, client, track.Title, track.Artist)
		if err != nil {
			if strings.Contains(err.Error(), "quotaExceeded") {
				return playlistID, fmt.Errorf("quota exceeded during search: %w", err)
			}
			fmt.Printf("Skipping track %s - %s (search failed): %v\n", track.Title, track.Artist, err)
			continue
		}
		if videoID == "" {
			fmt.Printf("Skipping track %s - %s (not found)\n", track.Title, track.Artist)
			continue
		}
		if err := AddYouTubePlaylistItem(ctx, client, playlistID, videoID); err != nil {
			if strings.Contains(err.Error(), "quotaExceeded") {
				return playlistID, fmt.Errorf("quota exceeded during add: %w", err)
			}
			fmt.Printf("Failed to add track to YouTube: %v\n", err)
		} else {
			fmt.Printf("Added track %s to YouTube playlist\n", track.Title)
		}
		time.Sleep(500 * time.Millisecond)
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

	var err error
	maxRetries := 5
	baseDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		call := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
		_, err = call.Do()
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "409") || strings.Contains(err.Error(), "500") || strings.Contains(err.Error(), "503") {
			delay := baseDelay * time.Duration(1<<i)
			fmt.Printf("   ⚠️ YouTube API Error (Attempt %d/%d): %v. Retrying in %v...\n", i+1, maxRetries, err, delay)
			time.Sleep(delay)
			continue
		}

		break
	}
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

// ImportAllYouTubePlaylists imports all playlists and tracks for a user
func ImportAllYouTubePlaylists(ctx context.Context, user db.User) error {
	client, err := GetYouTubeClient(ctx, user)
	if err != nil {
		return err
	}

	playlists, err := GetYouTubePlaylists(ctx, client)
	if err != nil {
		return err
	}

	for _, p := range playlists {
		p.OwnerID = user.ID
		var existing db.Playlist
		if err := db.DB.Where("owner_id = ? AND source_id = ? AND platform = 'youtube'", user.ID, p.SourceID).First(&existing).Error; err == nil {
			continue
		}

		if err := db.DB.Create(&p).Error; err != nil {
			log.Printf("Failed to create playlist %s: %v", p.Title, err)
			continue
		}

		tracks, err := GetYouTubePlaylistTracks(ctx, client, p.SourceID)
		if err != nil {
			log.Printf("Failed to get tracks for %s: %v", p.Title, err)
			continue
		}

		for _, t := range tracks {
			t.PlaylistID = p.ID
			db.DB.Create(&t)
		}
	}
	return nil
}

// ImportYouTubePlaylist imports a single YouTube playlist to DB
func ImportYouTubePlaylist(ctx context.Context, client *youtube.Service, user db.User, sourceID string) error {
	call := client.Playlists.List([]string{"id", "snippet"}).Id(sourceID)
	response, err := call.Do()
	if err != nil || len(response.Items) == 0 {
		return fmt.Errorf("failed to fetch YouTube playlist: %w", err)
	}
	ytPlaylist := response.Items[0]
	var coverImage string
	if ytPlaylist.Snippet.Thumbnails != nil && ytPlaylist.Snippet.Thumbnails.Default != nil {
		coverImage = ytPlaylist.Snippet.Thumbnails.Default.Url
	}

	playlist := db.Playlist{
		ID:          uuid.New(),
		OwnerID:     user.ID,
		Title:       ytPlaylist.Snippet.Title,
		Description: ytPlaylist.Snippet.Description,
		Platform:    "youtube",
		SourceID:    sourceID,
		IsPublic:    false,
		CoverImage:  coverImage,
		CreatedAt:   time.Now(),
	}

	var existing db.Playlist
	if err := db.DB.Where("owner_id = ? AND source_id = ? AND platform = 'youtube'", user.ID, sourceID).First(&existing).Error; err == nil {
		playlist = existing
	} else {
		if err := db.DB.Create(&playlist).Error; err != nil {
			return fmt.Errorf("failed to create playlist in DB: %w", err)
		}
	}

	tracks, err := GetYouTubePlaylistTracks(ctx, client, sourceID)
	if err != nil {
		return fmt.Errorf("failed to fetch tracks: %w", err)
	}

	for _, t := range tracks {
		t.PlaylistID = playlist.ID
		db.DB.Create(&t)
	}
	return nil
}
