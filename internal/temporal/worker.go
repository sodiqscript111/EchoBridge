package temporal

import (
	"log"

	"go.temporal.io/sdk/worker"
)

func StartWorker() error {
	c := GetClient()
	if c == nil {
		return nil
	}

	w := worker.New(c, PlaylistSyncTaskQueue, worker.Options{})

	w.RegisterWorkflow(PlaylistSyncWorkflow)
	w.RegisterWorkflow(ImportPlaylistWorkflow)

	w.RegisterActivity(FetchUserActivity)
	w.RegisterActivity(FetchPlaylistActivity)
	w.RegisterActivity(FetchPlaylistTracksActivity)
	w.RegisterActivity(CreateSpotifyPlaylistActivity)
	w.RegisterActivity(SearchSpotifyTrackActivity)
	w.RegisterActivity(AddTrackToSpotifyActivity)
	w.RegisterActivity(CreateYouTubePlaylistActivity)
	w.RegisterActivity(SearchYouTubeVideoActivity)
	w.RegisterActivity(AddVideoToYouTubePlaylistActivity)
	w.RegisterActivity(ImportSpotifyPlaylistActivity)
	w.RegisterActivity(ImportYouTubePlaylistActivity)

	log.Println("🚀 Temporal worker started on queue:", PlaylistSyncTaskQueue)
	return w.Run(worker.InterruptCh())
}
