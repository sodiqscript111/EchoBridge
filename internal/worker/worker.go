package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/services"

	"github.com/google/uuid"
)

// Job represents a background job
type Job struct {
	Type       string // "sync" or "categorize"
	JobID      uuid.UUID
	UserID     uuid.UUID
	PlaylistID uuid.UUID
	Platforms  []string // Used for sync
}

// WorkerPool handles background jobs
type WorkerPool struct {
	JobQueue chan Job
	wg       sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(bufferSize int) *WorkerPool {
	return &WorkerPool{
		JobQueue: make(chan Job, bufferSize),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for job := range wp.JobQueue {
		log.Printf("Processing %s job %s for playlist %s", job.Type, job.JobID, job.PlaylistID)

		if job.Type == "sync" {
			wp.handleSyncJob(job)
		} else if job.Type == "categorize" {
			wp.handleCategorizeJob(job)
		}
	}
}

func (wp *WorkerPool) handleSyncJob(job Job) {
	// Update status to "processing"
	db.DB.Model(&db.SyncJob{}).Where("id = ?", job.JobID).Updates(map[string]interface{}{
		"status": "processing",
	})

	// Fetch user to get tokens
	var user db.User
	if err := db.DB.Where("id = ?", job.UserID).First(&user).Error; err != nil {
		log.Printf("Failed to fetch user for job: %v", err)
		// Update status to "failed"
		db.DB.Model(&db.SyncJob{}).Where("id = ?", job.JobID).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": err.Error(),
		})
		return
	}

	// Perform sync
	result, err := services.SyncPlaylist(context.Background(), user, job.PlaylistID, job.Platforms)
	if err != nil {
		log.Printf("Sync failed for playlist %s: %v", job.PlaylistID, err)
		// Update status to "failed"
		db.DB.Model(&db.SyncJob{}).Where("id = ?", job.JobID).Updates(map[string]interface{}{
			"status":    "failed",
			"error_msg": err.Error(),
		})
	} else {
		log.Printf("Sync completed for playlist %s", job.PlaylistID)
		// Marshal result to JSON
		resultJSON, _ := json.Marshal(result)
		now := time.Now()
		// Update status to "completed"
		db.DB.Model(&db.SyncJob{}).Where("id = ?", job.JobID).Updates(map[string]interface{}{
			"status":       "completed",
			"result":       string(resultJSON),
			"completed_at": &now,
		})
	}
}

func (wp *WorkerPool) handleCategorizeJob(job Job) {
	log.Printf("Categorizing playlist %s...", job.PlaylistID)
	err := services.CategorizePlaylist(context.Background(), job.PlaylistID)
	if err != nil {
		log.Printf("Categorization failed: %v", err)
	} else {
		log.Printf("Categorization completed for playlist %s", job.PlaylistID)
	}
}

// Submit submits a job to the pool
func (wp *WorkerPool) Submit(job Job) {
	wp.JobQueue <- job
}
