package handlers

import (
	"fmt"
	"net/http"

	"EchoBridge/db"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DebugCategorizePlaylist manually triggers categorization for a playlist
func DebugCategorizePlaylist(c *gin.Context) {
	playlistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID"})
		return
	}

	if WorkerPool != nil {
		WorkerPool.Submit(worker.Job{
			Type:       "categorize",
			JobID:      uuid.New(),
			PlaylistID: playlistID,
		})
		c.JSON(http.StatusOK, gin.H{"message": "Categorization job submitted", "playlist_id": playlistID})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Worker pool not initialized"})
	}
}

// DebugCategorizeAll categorizes all PUBLIC playlists that don't have a category
func DebugCategorizeAll(c *gin.Context) {
	if WorkerPool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Worker pool not initialized"})
		return
	}

	var playlists []db.Playlist
	// Find all PUBLIC playlists where category is empty or null
	if err := db.DB.Where("is_public = ? AND (category = ? OR category IS NULL)", true, "").Find(&playlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlists", "details": err.Error()})
		return
	}

	if len(playlists) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "All public playlists are already categorized", "count": 0})
		return
	}

	// Note: Free tier Gemini has 15 requests/minute limit
	// Workers will process these sequentially with rate limiting
	count := 0
	for _, p := range playlists {
		WorkerPool.Submit(worker.Job{
			Type:       "categorize",
			JobID:      uuid.New(),
			PlaylistID: p.ID,
		})
		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Submitted %d public playlists for categorization", count),
		"count":   count,
		"note":    "Processing with rate limiting (15 requests/minute for free tier)",
	})
}
