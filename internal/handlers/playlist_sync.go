package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SyncPlaylist syncs a playlist to specified platforms
func SyncPlaylist(c *gin.Context) {
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

	var input struct {
		Platforms []string `json:"platforms" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	if WorkerPool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Worker pool not initialized"})
		return
	}

	// Create SyncJob record
	jobID := uuid.New()
	platformsJSON, _ := json.Marshal(input.Platforms)
	syncJob := db.SyncJob{
		ID:         jobID,
		UserID:     userID,
		PlaylistID: playlistID,
		Platforms:  string(platformsJSON),
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
	if err := db.DB.Create(&syncJob).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sync job", "details": err.Error()})
		return
	}

	job := worker.Job{
		JobID:      jobID,
		UserID:     userID,
		PlaylistID: playlistID,
		Platforms:  input.Platforms,
	}
	WorkerPool.Submit(job)

	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Sync started in background",
		"job_id":      jobID,
		"playlist_id": playlistID,
		"status":      "pending",
	})
}

// GetSyncStatus retrieves the status of a sync job
func GetSyncStatus(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("jobID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var job db.SyncJob
	if err := db.DB.Where("id = ?", jobID).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	response := gin.H{
		"job_id":       job.ID,
		"status":       job.Status,
		"created_at":   job.CreatedAt,
		"completed_at": job.CompletedAt,
	}

	if job.Result != "" {
		var result map[string]interface{}
		json.Unmarshal([]byte(job.Result), &result)
		response["result"] = result
	}
	if job.ErrorMsg != "" {
		response["error"] = job.ErrorMsg
	}

	c.JSON(http.StatusOK, response)
}
