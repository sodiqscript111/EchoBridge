package handlers

import (
	"net/http"
	"strings"

	"EchoBridge/db"
	"EchoBridge/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
)

// SpotifyPlaylists retrieves user Spotify playlists (Direct API call)
func SpotifyPlaylists(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	playlists, err := services.GetSpotifyPlaylists(c.Request.Context(), client, dbUser.SpotifyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlists", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, p := range playlists {
		result = append(result, gin.H{
			"platform":    "spotify",
			"id":          p.SourceID,
			"title":       p.Title,
			"description": p.Description,
			"cover_image": p.CoverImage,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Spotify playlists retrieved", "playlists": result})
}

// SpotifyPlaylistTracks retrieves tracks from a Spotify playlist (Direct API call, not DB)
func SpotifyPlaylistTracks(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	tracks, err := services.GetSpotifyPlaylistTracks(c.Request.Context(), client, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify tracks", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, t := range tracks {
		result = append(result, gin.H{
			"name":    t.Title,
			"artists": t.Artist,
			"album":   t.Album,
			"uri":     t.SpotifyID,
			"id":      t.SpotifyID,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"platform":    "spotify",
		"playlist_id": c.Param("id"),
		"tracks":      result,
	})
}

// ImportToSpotify imports a playlist to Spotify
func ImportToSpotify(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	playlistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var playlist db.Playlist
	if err := db.DB.Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check if user has access (Owner or Public)
	if playlist.OwnerID != userID && !playlist.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this playlist"})
		return
	}

	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
		return
	}

	client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	spotifyID, err := services.ImportToSpotify(c.Request.Context(), client, dbUser, playlist, tracks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import to Spotify", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Playlist imported to Spotify",
		"spotify_playlist_id": spotifyID,
	})
}

// ExportSpotifyToYouTube exports a Spotify playlist to YouTube
func ExportSpotifyToYouTube(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	spotifyPlaylistID := c.Param("spotifyPlaylistID")
	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	spotifyClient, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	youtubeClient, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
		return
	}

	spotifyPlaylist, err := spotifyClient.GetPlaylist(c.Request.Context(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlist", "details": err.Error()})
		return
	}

	youtubePlaylistID, err := services.CreateYouTubePlaylist(c.Request.Context(), youtubeClient, spotifyPlaylist.Name, spotifyPlaylist.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create YouTube playlist", "details": err.Error()})
		return
	}

	tracksPage, err := spotifyClient.GetPlaylistTracks(c.Request.Context(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify tracks", "details": err.Error()})
		return
	}

	var exported, failed, skipped int
	for _, item := range tracksPage.Tracks {
		if item.Track.ID == "" {
			skipped++
			continue
		}
		videoID, err := services.SearchYouTubeVideo(c.Request.Context(), youtubeClient, item.Track.Name, strings.Join(services.GetArtists(item.Track.Artists), ", "))
		if err != nil || videoID == "" {
			failed++
			continue
		}
		if err := services.AddYouTubePlaylistItem(c.Request.Context(), youtubeClient, youtubePlaylistID, videoID); err != nil {
			failed++
		} else {
			exported++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                    "Playlist exported to YouTube",
		"spotify_playlist_name":      spotifyPlaylist.Name,
		"spotify_playlist_id":        spotifyPlaylistID,
		"youtube_playlist_id":        youtubePlaylistID,
		"total_spotify_tracks":       spotifyPlaylist.Tracks.Total,
		"tracks_exported_to_youtube": exported,
		"tracks_failed_to_export":    failed,
		"tracks_skipped_no_youtube":  skipped,
	})
}
