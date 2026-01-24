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
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// --- CONFIGURATION ---
func getSpotifyAuth() *spotifyauth.Authenticator {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	redirectURL := os.Getenv("SPOTIFY_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://127.0.0.1:8000/callback/spotify"
	}

	return spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(redirectURL),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
		),
	)
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
	spotifyAuth := getSpotifyAuth()
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
				ID:         uuid.New(),
				Title:      item.Track.Name,
				Artist:     strings.Join(GetArtists(item.Track.Artists), ", "),
				Album:      item.Track.Album.Name,
				SpotifyID:  item.Track.ID.String(),
				PreviewURL: item.Track.PreviewURL,
				CreatedAt:  time.Now(),
			})
			log.Printf("Track: %s, PreviewURL: %s", item.Track.Name, item.Track.PreviewURL)
		}
	}
	return tracks, nil
}

// GetSpotifyPlaylists retrieves user playlists
func GetSpotifyPlaylists(ctx context.Context, client *spotify.Client, userID string) ([]db.Playlist, error) {
	playlistsPage, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get spotify playlists: %w", err)
	}

	var playlists []db.Playlist
	for _, p := range playlistsPage.Playlists {
		var coverImage string
		if len(p.Images) > 0 {
			coverImage = p.Images[0].URL
		}
		playlists = append(playlists, db.Playlist{
			ID:          uuid.New(),
			Title:       p.Name,
			Description: p.Description,
			Platform:    "spotify",
			SourceID:    p.ID.String(),
			IsPublic:    false,
			CoverImage:  coverImage,
			CreatedAt:   time.Now(),
		})
	}
	return playlists, nil
}

// ImportAllSpotifyPlaylists imports all playlists and tracks for a user
func ImportAllSpotifyPlaylists(ctx context.Context, user db.User) error {
	client, err := GetSpotifyClient(ctx, user)
	if err != nil {
		return err
	}

	playlists, err := GetSpotifyPlaylists(ctx, client, user.SpotifyID)
	if err != nil {
		return err
	}

	for _, p := range playlists {
		p.OwnerID = user.ID
		var existing db.Playlist
		if err := db.DB.Where("owner_id = ? AND source_id = ? AND platform = 'spotify'", user.ID, p.SourceID).First(&existing).Error; err == nil {
			continue
		}

		if err := db.DB.Create(&p).Error; err != nil {
			log.Printf("Failed to create playlist %s: %v", p.Title, err)
			continue
		}

		tracks, err := GetSpotifyPlaylistTracks(ctx, client, p.SourceID)
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

// ImportSpotifyPlaylist imports a single Spotify playlist to DB
func ImportSpotifyPlaylist(ctx context.Context, client *spotify.Client, user db.User, sourceID string) error {
	spPlaylist, err := client.GetPlaylist(ctx, spotify.ID(sourceID))
	if err != nil {
		return fmt.Errorf("failed to fetch Spotify playlist: %w", err)
	}
	var coverImage string
	if len(spPlaylist.Images) > 0 {
		coverImage = spPlaylist.Images[0].URL
	}

	playlist := db.Playlist{
		ID:          uuid.New(),
		OwnerID:     user.ID,
		Title:       spPlaylist.Name,
		Description: spPlaylist.Description,
		Platform:    "spotify",
		SourceID:    sourceID,
		IsPublic:    false,
		CoverImage:  coverImage,
		CreatedAt:   time.Now(),
	}

	var existing db.Playlist
	if err := db.DB.Where("owner_id = ? AND source_id = ? AND platform = 'spotify'", user.ID, sourceID).First(&existing).Error; err == nil {
		playlist = existing
	} else {
		if err := db.DB.Create(&playlist).Error; err != nil {
			return fmt.Errorf("failed to create playlist in DB: %w", err)
		}
	}

	tracks, err := GetSpotifyPlaylistTracks(ctx, client, sourceID)
	if err != nil {
		return fmt.Errorf("failed to fetch tracks: %w", err)
	}

	for _, t := range tracks {
		t.PlaylistID = playlist.ID
		db.DB.Create(&t)
	}
	return nil
}

func GetArtists(artists []spotify.SimpleArtist) []string {
	var names []string
	for _, a := range artists {
		names = append(names, a.Name)
	}
	return names
}
