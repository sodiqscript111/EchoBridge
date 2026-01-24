package handlers

import (
	"net/http"

	"EchoBridge/db"
	"EchoBridge/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// YouTubePlaylists retrieves user YouTube playlists
func YouTubePlaylists(c *gin.Context) {
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

	client, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
		return
	}

	playlists, err := services.GetYouTubePlaylists(c.Request.Context(), client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve YouTube playlists", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, p := range playlists {
		result = append(result, gin.H{
			"platform":    "youtube",
			"id":          p.SourceID,
			"title":       p.Title,
			"description": p.Description,
			"cover_image": p.CoverImage,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "YouTube playlists retrieved", "playlists": result})
}

// ImportToYouTube imports a playlist to YouTube
func ImportToYouTube(c *gin.Context) {
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

	client, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
		return
	}

	youtubeID, err := services.ImportToYouTube(c.Request.Context(), client, playlist, tracks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import to YouTube", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Playlist imported to YouTube",
		"youtube_playlist_id": youtubeID,
	})
}
