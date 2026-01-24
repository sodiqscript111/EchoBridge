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
		} else if job.Type == "import_all" {
			wp.handleImportAllJob(job)
		} else if job.Type == "import_playlist" {
			wp.handleImportPlaylistJob(job)
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

	// Rate limiting: Free tier Gemini API allows 15 requests/minute
	// That's 1 request every 4 seconds. Let's add a 5-second delay to be safe.
	time.Sleep(5 * time.Second)

	err := services.CategorizePlaylist(context.Background(), job.PlaylistID)
	if err != nil {
		log.Printf("Categorization failed: %v", err)
	} else {
		log.Printf("Categorization completed for playlist %s", job.PlaylistID)
	}
}

func (wp *WorkerPool) handleImportAllJob(job Job) {
	log.Printf("Starting Import All for user %s, platforms: %v", job.UserID, job.Platforms)

	var user db.User
	if err := db.DB.Where("id = ?", job.UserID).First(&user).Error; err != nil {
		log.Printf("Failed to fetch user for import job: %v", err)
		return
	}

	for _, platform := range job.Platforms {
		if platform == "spotify" {
			if err := services.ImportAllSpotifyPlaylists(context.Background(), user); err != nil {
				log.Printf("Failed to import Spotify playlists: %v", err)
			} else {
				log.Printf("Successfully imported Spotify playlists for user %s", user.Username)
			}
		} else if platform == "youtube" {
			if err := services.ImportAllYouTubePlaylists(context.Background(), user); err != nil {
				log.Printf("Failed to import YouTube playlists: %v", err)
			} else {
				log.Printf("Successfully imported YouTube playlists for user %s", user.Username)
			}
		}
	}
}

func (wp *WorkerPool) handleImportPlaylistJob(job Job) {
	if len(job.Platforms) < 2 {
		log.Printf("Invalid job data for import_playlist")
		return
	}
	platform := job.Platforms[0]
	sourceID := job.Platforms[1]

	log.Printf("Importing playlist %s from %s for user %s", sourceID, platform, job.UserID)

	var user db.User
	if err := db.DB.Where("id = ?", job.UserID).First(&user).Error; err != nil {
		log.Printf("Failed to fetch user: %v", err)
		return
	}

	if platform == "spotify" {
		client, err := services.GetSpotifyClient(context.Background(), user)
		if err != nil {
			log.Printf("Failed to get Spotify client: %v", err)
			return
		}
		if err := services.ImportSpotifyPlaylist(context.Background(), client, user, sourceID); err != nil {
			log.Printf("Failed to import Spotify playlist: %v", err)
		}
	} else if platform == "youtube" {
		client, err := services.GetYouTubeClient(context.Background(), user)
		if err != nil {
			log.Printf("Failed to get YouTube client: %v", err)
			return
		}
		if err := services.ImportYouTubePlaylist(context.Background(), client, user, sourceID); err != nil {
			log.Printf("Failed to import YouTube playlist: %v", err)
		}
	}
}

// Submit submits a job to the pool
func (wp *WorkerPool) Submit(job Job) {
	wp.JobQueue <- job
}
