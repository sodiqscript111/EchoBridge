package scheduler

import (
	"log"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/worker"

	"github.com/google/uuid"
)


func StartCategorizationScheduler(pool *worker.WorkerPool) {
	// Run once immediately
	go func() {
		log.Println("Scheduler: Running initial categorization check...")
		enqueueUncategorized(pool)
	}()

	// Start ticker for every 1 hour
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			enqueueUncategorized(pool)
		}
	}()
}

func enqueueUncategorized(pool *worker.WorkerPool) {
	var playlists []db.Playlist
	// Fetch playlists where category is empty
	if err := db.DB.Where("category = ? OR category IS NULL", "").Find(&playlists).Error; err != nil {
		log.Printf("Scheduler: Failed to fetch playlists: %v", err)
		return
	}

	if len(playlists) == 0 {
		log.Println("Scheduler: No uncategorized playlists found.")
		return
	}

	log.Printf("Scheduler: Found %d uncategorized playlists. Enqueueing...", len(playlists))

	for _, p := range playlists {
		pool.Submit(worker.Job{
			Type:       "categorize",
			JobID:      uuid.New(),
			UserID:     p.OwnerID,
			PlaylistID: p.ID,
		})
	}
}