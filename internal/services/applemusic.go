package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"EchoBridge/db"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AppleMusicClient struct {
	httpClient *http.Client
}

var appleMusicOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("APPLE_MUSIC_CLIENT_ID"),
	ClientSecret: os.Getenv("APPLE_MUSIC_SECRET"),
	RedirectURL:  "http://127.0.0.1:8000/callback/applemusic",
	Scopes:       []string{"https://api.music.apple.com/v1/me/library"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://appleid.apple.com/auth/authorize",
		TokenURL: "https://appleid.apple.com/auth/token",
	},
}

// GetAppleMusicClient returns an Apple Music client
func GetAppleMusicClient(ctx context.Context, user db.User) (*AppleMusicClient, error) {
	if user.AppleMusicToken == "" {
		return nil, fmt.Errorf("no Apple Music token available")
	}
	var token oauth2.Token
	if err := json.Unmarshal([]byte(user.AppleMusicToken), &token); err != nil {
		return nil, fmt.Errorf("invalid Apple Music token: %w", err)
	}
	return &AppleMusicClient{httpClient: appleMusicOAuthConfig.Client(ctx, &token)}, nil
}

// GetAppleMusicPlaylists retrieves user playlists
func GetAppleMusicPlaylists(ctx context.Context, client *AppleMusicClient) ([]db.Playlist, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.music.apple.com/v1/me/library/playlists", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Apple Music request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple Music playlists: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Apple Music response: %w", err)
	}

	var playlists []db.Playlist
	for _, p := range result.Data {
		playlists = append(playlists, db.Playlist{
			ID:          uuid.New(),
			Title:       p.Attributes.Name,
			Description: p.Attributes.Description,
			Platform:    "applemusic",
			SourceID:    p.ID,
			IsPublic:    false,
			CreatedAt:   time.Now(),
		})
	}
	return playlists, nil
}

// SearchAppleMusicTrack searches for a track
func SearchAppleMusicTrack(ctx context.Context, client *AppleMusicClient, title, artist string) (string, error) {
	url := fmt.Sprintf("https://api.music.apple.com/v1/catalog/us/songs?filter[term]=%s+%s", url.QueryEscape(title), url.QueryEscape(artist))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Apple Music search request: %w", err)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to search Apple Music track: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Apple Music search response: %w", err)
	}

	if len(result.Data) > 0 {
		return result.Data[0].ID, nil
	}
	return "", nil
}
