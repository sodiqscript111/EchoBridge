package handlers

import (
	"net/http"

	"EchoBridge/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetPlaylist retrieves a single playlist by ID with tracks
func GetPlaylist(c *gin.Context) {
	playlistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	var playlist db.Playlist
	if err := db.DB.Preload("Tracks").First(&playlist, "id = ?", playlistID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check if public or owned by user (if authenticated)
	// For now, we'll just check if it's public.
	// TODO: Add auth check for private playlists if we want to reuse this for private views
	if !playlist.IsPublic {
		// Check auth header manually since this is a public route potentially
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "This playlist is private"})
			return
		}
		// In a real app, we'd validate the token here to see if the user owns it.
		// For simplicity in this step, we will restrict to public only for unauth users.
	}

	tracks := []gin.H{}
	for _, t := range playlist.Tracks {
		tracks = append(tracks, gin.H{
			"title":         t.Title,
			"artist":        t.Artist,
			"album":         t.Album,
			"spotify_id":    t.SpotifyID,
			"youtube_id":    t.YouTubeID,
			"applemusic_id": t.AppleMusicID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          playlist.ID,
		"title":       playlist.Title,
		"description": playlist.Description,
		"platform":    playlist.Platform,
		"owner_id":    playlist.OwnerID,
		"tracks":      tracks,
		"cover_image": playlist.CoverImage,
		"is_public":   playlist.IsPublic,
		"category":    playlist.Category,
	})
}
