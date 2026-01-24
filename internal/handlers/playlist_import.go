package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/services"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
)

// PostPlaylist creates a new playlist AND imports tracks
func PostPlaylist(c *gin.Context) {
	var input struct {
		Platform string `json:"platform" binding:"required"`
		SourceID string `json:"source_id" binding:"required"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

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

	var playlist db.Playlist
	var importedTracksCount int

	switch input.Platform {
	case "spotify":
		client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
			return
		}

		// 1. Fetch Playlist Metadata
		spPlaylist, err := client.GetPlaylist(c.Request.Context(), spotify.ID(input.SourceID))
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to fetch Spotify playlist", "details": err.Error()})
			return
		}
		var coverImage string
		if len(spPlaylist.Images) > 0 {
			coverImage = spPlaylist.Images[0].URL
		}

		// 2. Create Playlist in DB
		playlist = db.Playlist{
			ID:          uuid.New(),
			OwnerID:     userID,
			Title:       spPlaylist.Name,
			Description: spPlaylist.Description,
			Platform:    "spotify",
			SourceID:    input.SourceID,
			IsPublic:    input.IsPublic,
			CoverImage:  coverImage,
			CreatedAt:   time.Now(),
		}
		if err := db.DB.Create(&playlist).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist", "details": err.Error()})
			return
		}

		// --- CRITICAL FIX: Fetch and Save Tracks ---
		tracks, err := services.GetSpotifyPlaylistTracks(c.Request.Context(), client, input.SourceID)
		if err != nil {
			fmt.Printf("Error fetching tracks: %v\n", err)
		} else {
			for _, t := range tracks {
				t.PlaylistID = playlist.ID // Link track to the new playlist
				if err := db.DB.Create(&t).Error; err == nil {
					importedTracksCount++
				}
			}
		}

	case "youtube":
		client, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
			return
		}
		call := client.Playlists.List([]string{"id", "snippet"}).Id(input.SourceID)
		response, err := call.Do()
		if err != nil || len(response.Items) == 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to fetch YouTube playlist"})
			return
		}
		ytPlaylist := response.Items[0]
		var coverImage string
		if ytPlaylist.Snippet.Thumbnails != nil && ytPlaylist.Snippet.Thumbnails.Default != nil {
			coverImage = ytPlaylist.Snippet.Thumbnails.Default.Url
		}
		playlist = db.Playlist{
			ID:          uuid.New(),
			OwnerID:     userID,
			Title:       ytPlaylist.Snippet.Title,
			Description: ytPlaylist.Snippet.Description,
			Platform:    "youtube",
			SourceID:    input.SourceID,
			IsPublic:    input.IsPublic,
			CoverImage:  coverImage,
			CreatedAt:   time.Now(),
		}
		if err := db.DB.Create(&playlist).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist", "details": err.Error()})
			return
		}

		// Import Tracks
		tracks, err := services.GetYouTubePlaylistTracks(c.Request.Context(), client, input.SourceID)
		if err != nil {
			fmt.Printf("Error fetching YouTube tracks: %v\n", err)
		} else {
			for _, t := range tracks {
				t.PlaylistID = playlist.ID
				if err := db.DB.Create(&t).Error; err == nil {
					importedTracksCount++
				}
			}
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported platform"})
		return
	}

	// Submit Categorization Job
	if WorkerPool != nil {
		WorkerPool.Submit(worker.Job{
			Type:       "categorize",
			JobID:      uuid.New(),
			UserID:     userID,
			PlaylistID: playlist.ID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Playlist imported successfully",
		"playlist_id":     playlist.ID,
		"cover_image":     playlist.CoverImage,
		"tracks_imported": importedTracksCount,
	})
}

// BatchImportPlaylists imports multiple playlists
func BatchImportPlaylists(c *gin.Context) {
	var input struct {
		Playlists []struct {
			Platform string `json:"platform"`
			SourceID string `json:"source_id"`
		} `json:"playlists" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if WorkerPool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Worker pool not initialized"})
		return
	}

	jobID := uuid.New()
	count := 0
	for _, p := range input.Playlists {
		WorkerPool.Submit(worker.Job{
			Type:       "import_playlist",
			JobID:      uuid.New(),
			UserID:     userID,
			PlaylistID: uuid.Nil,
			Platforms:  []string{p.Platform, p.SourceID}, // [0]=platform, [1]=sourceID
		})
		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Started import for %d playlists", count),
		"job_id":  jobID,
	})
}

// ImportPublicPlaylist imports a public playlist to the user's chosen platform with validation
func ImportPublicPlaylist(c *gin.Context) {
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
		Platform string `json:"platform" binding:"required,oneof=spotify youtube"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Platform must be 'spotify' or 'youtube'", "details": err.Error()})
		return
	}

	// Get user
	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if platform is connected
	platformConnected := false
	if input.Platform == "spotify" && dbUser.SpotifyID != "" {
		platformConnected = true
	} else if input.Platform == "youtube" && dbUser.YouTubeID != "" {
		platformConnected = true
	}

	if !platformConnected {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    fmt.Sprintf("Please connect your %s account first", strings.Title(input.Platform)),
			"code":     "PLATFORM_NOT_CONNECTED",
			"platform": input.Platform,
		})
		return
	}

	// Get the public playlist
	var playlist db.Playlist
	if err := db.DB.Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Ensure it's public (unless user owns it)
	if !playlist.IsPublic && playlist.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "This playlist is not public"})
		return
	}

	// Check if user already has this playlist on the target platform
	var existingPlaylist db.Playlist
	err = db.DB.Where("owner_id = ? AND source_id = ? AND platform = ?", userID, playlist.SourceID, input.Platform).First(&existingPlaylist).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":                fmt.Sprintf("You already have this playlist on %s", strings.Title(input.Platform)),
			"code":                 "PLAYLIST_ALREADY_EXISTS",
			"existing_playlist_id": existingPlaylist.ID,
		})
		return
	}

	// Get tracks
	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
		return
	}

	// Import to the chosen platform
	var targetPlaylistID string
	if input.Platform == "spotify" {
		client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to connect to Spotify"})
			return
		}
		targetPlaylistID, err = services.ImportToSpotify(c.Request.Context(), client, dbUser, playlist, tracks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import to Spotify", "details": err.Error()})
			return
		}
	} else if input.Platform == "youtube" {
		client, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to connect to YouTube"})
			return
		}
		targetPlaylistID, err = services.ImportToYouTube(c.Request.Context(), client, playlist, tracks)
		if err != nil {
			fmt.Printf("Error importing to YouTube: %v\n", err)
			if strings.Contains(err.Error(), "quotaExceeded") {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "YouTube API daily quota exceeded. Please try again tomorrow (usually resets at midnight PT).",
					"code":  "QUOTA_EXCEEDED",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import to YouTube", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     fmt.Sprintf("Playlist imported to %s successfully", strings.Title(input.Platform)),
		"platform":    input.Platform,
		"playlist_id": targetPlaylistID,
	})
}
