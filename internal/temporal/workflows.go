package temporal

import (
	"fmt"
	"time"

	"EchoBridge/db"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	PlaylistSyncTaskQueue = "playlist-sync"
	// TEST MODE: Set to true to simulate rate limiting
	TestModeEnabled       = true
	TestRateLimitAfter    = 10               // Pause after every N tracks
	TestRateLimitDuration = 30 * time.Second // Pause duration
)

type PlaylistSyncInput struct {
	UserID     uuid.UUID
	PlaylistID uuid.UUID
	Platforms  []string
	TestMode   bool // Enable test mode for this specific workflow
}

type PlaylistSyncResult struct {
	SpotifyPlaylistID string
	YouTubePlaylistID string
	TracksProcessed   int
	TracksFailed      int
}

type TrackSyncProgress struct {
	CurrentTrack int
	TotalTracks  int
	TrackTitle   string
	Platform     string
}

func PlaylistSyncWorkflow(ctx workflow.Context, input PlaylistSyncInput) (*PlaylistSyncResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting PlaylistSyncWorkflow", "playlistID", input.PlaylistID, "testMode", input.TestMode || TestModeEnabled)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var user db.User
	err := workflow.ExecuteActivity(ctx, FetchUserActivity, input.UserID).Get(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	var playlist db.Playlist
	err = workflow.ExecuteActivity(ctx, FetchPlaylistActivity, input.PlaylistID, input.UserID).Get(ctx, &playlist)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist: %w", err)
	}

	var tracks []db.Track
	err = workflow.ExecuteActivity(ctx, FetchPlaylistTracksActivity, input.PlaylistID).Get(ctx, &tracks)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tracks: %w", err)
	}

	if len(tracks) == 0 {
		logger.Warn("No tracks found in playlist")
		return &PlaylistSyncResult{TracksProcessed: 0}, nil
	}

	logger.Info("Fetched tracks", "count", len(tracks))

	result := &PlaylistSyncResult{
		TracksProcessed: 0,
		TracksFailed:    0,
	}

	// Check if test mode is enabled
	testMode := input.TestMode || TestModeEnabled

	for _, platform := range input.Platforms {
		switch platform {
		case "spotify":
			if user.SpotifyToken == "" {
				logger.Warn("No Spotify token, skipping")
				continue
			}

			var spotifyPlaylistID string
			err = workflow.ExecuteActivity(ctx, CreateSpotifyPlaylistActivity, user, playlist).Get(ctx, &spotifyPlaylistID)
			if err != nil {
				logger.Error("Failed to create Spotify playlist", "error", err)
				continue
			}
			result.SpotifyPlaylistID = spotifyPlaylistID
			logger.Info("Created Spotify playlist", "playlistID", spotifyPlaylistID)

			for i, track := range tracks {
				logger.Info("Processing track", "index", i+1, "total", len(tracks), "title", track.Title)

				if track.SpotifyID != "" {
					err = workflow.ExecuteActivity(ctx, AddTrackToSpotifyActivity, user, spotifyPlaylistID, track.SpotifyID).Get(ctx, nil)
					if err != nil {
						logger.Warn("Failed to add track to Spotify", "track", track.Title, "error", err)
						result.TracksFailed++
					} else {
						logger.Info("Added track to Spotify", "track", track.Title)
						result.TracksProcessed++
					}
				} else {
					var foundTrackID string
					err = workflow.ExecuteActivity(ctx, SearchSpotifyTrackActivity, user, track.Title, track.Artist).Get(ctx, &foundTrackID)
					if err != nil || foundTrackID == "" {
						logger.Warn("Track not found on Spotify", "track", track.Title)
						result.TracksFailed++
						continue
					}
					err = workflow.ExecuteActivity(ctx, AddTrackToSpotifyActivity, user, spotifyPlaylistID, foundTrackID).Get(ctx, nil)
					if err != nil {
						result.TracksFailed++
					} else {
						logger.Info("Added track to Spotify", "track", track.Title)
						result.TracksProcessed++
					}
				}

				// TEST MODE: Simulate rate limit after every N tracks
				if testMode && (i+1)%TestRateLimitAfter == 0 && i+1 < len(tracks) {
					logger.Warn("🚦 TEST MODE: Simulated rate limit hit! Pausing workflow...", "tracksProcessed", i+1, "pauseDuration", TestRateLimitDuration)
					workflow.Sleep(ctx, TestRateLimitDuration)
					logger.Info("🟢 TEST MODE: Resuming after simulated rate limit pause")
				}
			}

		case "youtube":
			if user.YouTubeToken == "" {
				logger.Warn("No YouTube token, skipping")
				continue
			}

			var ytPlaylistID string
			err = workflow.ExecuteActivity(ctx, CreateYouTubePlaylistActivity, user, playlist).Get(ctx, &ytPlaylistID)
			if err != nil {
				logger.Error("Failed to create YouTube playlist", "error", err)
				continue
			}
			result.YouTubePlaylistID = ytPlaylistID
			logger.Info("Created YouTube playlist", "playlistID", ytPlaylistID)

			for i, track := range tracks {
				logger.Info("Processing track", "index", i+1, "total", len(tracks), "title", track.Title)

				var videoID string
				err = workflow.ExecuteActivity(ctx, SearchYouTubeVideoActivity, user, track.Title, track.Artist).Get(ctx, &videoID)
				if err != nil || videoID == "" {
					logger.Warn("Video not found on YouTube", "track", track.Title)
					result.TracksFailed++
					continue
				}

				err = workflow.ExecuteActivity(ctx, AddVideoToYouTubePlaylistActivity, user, ytPlaylistID, videoID).Get(ctx, nil)
				if err != nil {
					logger.Warn("Failed to add video to YouTube", "track", track.Title, "error", err)
					result.TracksFailed++
				} else {
					logger.Info("Added video to YouTube", "track", track.Title)
					result.TracksProcessed++
				}

				// TEST MODE: Simulate rate limit after every N tracks
				if testMode && (i+1)%TestRateLimitAfter == 0 && i+1 < len(tracks) {
					logger.Warn("🚦 TEST MODE: Simulated rate limit hit! Pausing workflow...", "tracksProcessed", i+1, "pauseDuration", TestRateLimitDuration)
					workflow.Sleep(ctx, TestRateLimitDuration)
					logger.Info("🟢 TEST MODE: Resuming after simulated rate limit pause")
				}
			}
		}
	}

	logger.Info("PlaylistSyncWorkflow completed", "processed", result.TracksProcessed, "failed", result.TracksFailed)
	return result, nil
}

type ImportPlaylistInput struct {
	UserID   uuid.UUID
	Platform string
	SourceID string
}

func ImportPlaylistWorkflow(ctx workflow.Context, input ImportPlaylistInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ImportPlaylistWorkflow", "platform", input.Platform, "sourceID", input.SourceID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Minute,
			MaximumAttempts:    10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var user db.User
	err := workflow.ExecuteActivity(ctx, FetchUserActivity, input.UserID).Get(ctx, &user)
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	switch input.Platform {
	case "spotify":
		err = workflow.ExecuteActivity(ctx, ImportSpotifyPlaylistActivity, user, input.SourceID).Get(ctx, nil)
	case "youtube":
		err = workflow.ExecuteActivity(ctx, ImportYouTubePlaylistActivity, user, input.SourceID).Get(ctx, nil)
	default:
		return fmt.Errorf("unknown platform: %s", input.Platform)
	}

	if err != nil {
		return fmt.Errorf("failed to import playlist: %w", err)
	}

	logger.Info("ImportPlaylistWorkflow completed")
	return nil
}
