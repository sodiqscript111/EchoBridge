package handlers

import (
	"EchoBridge/internal/auth"
	"fmt"
	"net/http"
	"time"

	"EchoBridge/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterShareRoutes sets up song sharing routes
func RegisterShareRoutes(r *gin.Engine) {
	protected := r.Group("/api").Use(auth.AuthMiddleware())
	protected.POST("/share/track/:id", ShareTrack)
}

// ShareTrack generates platform-specific and custom links for a track
func ShareTrack(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	trackID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var track db.Track
	if err := db.DB.Where("id = ?", trackID).First(&track).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found"})
		return
	}

	links := make(map[string]string)
	if track.SpotifyID != "" && dbUser.SpotifyToken != "" {
		links["spotify"] = fmt.Sprintf("https://open.spotify.com/track/%s", track.SpotifyID)
	}
	if track.YouTubeID != "" && dbUser.YouTubeToken != "" {
		links["youtube"] = fmt.Sprintf("https://music.youtube.com/watch?v=%s", track.YouTubeID)
	}
	if track.AppleMusicID != "" && dbUser.AppleMusicToken != "" {
		links["applemusic"] = fmt.Sprintf("https://music.apple.com/us/song/%s", track.AppleMusicID)
	}

	customLink := fmt.Sprintf("https://yourdomain.com/share/track/%s", track.ID)
	links["custom"] = customLink

	share := db.Share{
		ID:        uuid.New(),
		TrackID:   trackID,
		UserID:    userID,
		CustomURL: customLink,
		CreatedAt: time.Now(),
	}
	if err := db.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save share link", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Track links generated",
		"track":   track.Title,
		"artist":  track.Artist,
		"links":   links,
	})
}
