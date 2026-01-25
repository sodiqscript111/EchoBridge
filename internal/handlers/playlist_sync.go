package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/temporal"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
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
		Platforms   []string `json:"platforms" binding:"required"`
		UseTemporal bool     `json:"use_temporal"` // Optional: explicitly use Temporal
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
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

	// Try to use Temporal if available
	temporalClient := temporal.GetClient()
	if temporalClient != nil {
		workflowOptions := client.StartWorkflowOptions{
			ID:        "playlist-sync-" + jobID.String(),
			TaskQueue: temporal.PlaylistSyncTaskQueue,
		}

		workflowInput := temporal.PlaylistSyncInput{
			UserID:     userID,
			PlaylistID: playlistID,
			Platforms:  input.Platforms,
			TestMode:   true, // Enable test mode for demo
		}

		we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, temporal.PlaylistSyncWorkflow, workflowInput)
		if err != nil {
			// Fall back to worker pool
			fallbackToWorkerPool(c, jobID, userID, playlistID, input.Platforms)
			return
		}

		// Update job with workflow ID
		db.DB.Model(&syncJob).Updates(map[string]interface{}{
			"status": "processing",
		})

		c.JSON(http.StatusAccepted, gin.H{
			"message":     "Sync started via Temporal workflow",
			"job_id":      jobID,
			"workflow_id": we.GetID(),
			"run_id":      we.GetRunID(),
			"playlist_id": playlistID,
			"status":      "processing",
			"temporal_ui": "http://localhost:8233",
		})
		return
	}

	// Fall back to worker pool
	fallbackToWorkerPool(c, jobID, userID, playlistID, input.Platforms)
}

func fallbackToWorkerPool(c *gin.Context, jobID, userID, playlistID uuid.UUID, platforms []string) {
	if WorkerPool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No worker available"})
		return
	}

	job := worker.Job{
		Type:       "sync",
		JobID:      jobID,
		UserID:     userID,
		PlaylistID: playlistID,
		Platforms:  platforms,
	}
	WorkerPool.Submit(job)

	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Sync started via worker pool (Temporal unavailable)",
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
