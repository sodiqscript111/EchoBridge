package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/auth"
	"EchoBridge/internal/services"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
)

var WorkerPool *worker.WorkerPool

// RegisterPlaylistRoutes sets up playlist-related routes
func RegisterPlaylistRoutes(r *gin.Engine) {
	r.GET("/playlists/public", GetPublicPlaylists)
	r.GET("/playlists/:id", GetPlaylist)
	r.GET("/debug/categorize/:id", DebugCategorizePlaylist) // Debug route
	protected := r.Group("/api").Use(auth.AuthMiddleware())
	protected.GET("/spotify/playlist/:id/tracks", SpotifyPlaylistTracks)
	protected.GET("/youtube/playlists", YouTubePlaylists)
	protected.GET("/my/playlists", GetUserPlaylists)
	protected.POST("/playlists", PostPlaylist)
	protected.PATCH("/playlists/:id/public", UpdateSinglePlaylistPublic)
	protected.POST("/import/playlist/:id/to/spotify", ImportToSpotify)
	protected.POST("/import/playlist/:id/to/youtube", ImportToYouTube)
	protected.GET("/export/spotify/:spotifyPlaylistID/to/youtube", ExportSpotifyToYouTube)
	protected.POST("/sync/playlist/:id", SyncPlaylist)
	protected.GET("/sync/status/:jobID", GetSyncStatus)
}

// GetPublicPlaylists retrieves all public playlists
func GetPublicPlaylists(c *gin.Context) {
	var playlists []db.Playlist
	if err := db.DB.Where("is_public = ?", true).Preload("Tracks").Find(&playlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public playlists", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, p := range playlists {
		tracks := []gin.H{}
		for _, t := range p.Tracks {
			tracks = append(tracks, gin.H{
				"title":         t.Title,
				"artist":        t.Artist,
				"album":         t.Album,
				"spotify_id":    t.SpotifyID,
				"youtube_id":    t.YouTubeID,
				"applemusic_id": t.AppleMusicID,
			})
		}
		result = append(result, gin.H{
			"id":          p.ID,
			"title":       p.Title,
			"description": p.Description,
			"platform":    p.Platform,
			"owner_id":    p.OwnerID,
			"tracks":      tracks,
			"cover_image": p.CoverImage,
			"category":    p.Category,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Public playlists retrieved", "playlists": result})
}

// SpotifyPlaylistTracks retrieves tracks from a Spotify playlist (Direct API call, not DB)
func SpotifyPlaylistTracks(c *gin.Context) {
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

	client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	tracks, err := services.GetSpotifyPlaylistTracks(c.Request.Context(), client, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify tracks", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, t := range tracks {
		result = append(result, gin.H{
			"name":    t.Title,
			"artists": t.Artist,
			"album":   t.Album,
			"uri":     t.SpotifyID,
			"id":      t.SpotifyID,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"platform":    "spotify",
		"playlist_id": c.Param("id"),
		"tracks":      result,
	})
}

// YouTubePlaylists retrieves user YouTube playlists
func YouTubePlaylists(c *gin.Context) {
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

	client, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
		return
	}

	playlists, err := services.GetYouTubePlaylists(c.Request.Context(), client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve YouTube playlists", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, p := range playlists {
		result = append(result, gin.H{
			"platform":    "youtube",
			"id":          p.SourceID,
			"title":       p.Title,
			"description": p.Description,
			"cover_image": p.CoverImage,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "YouTube playlists retrieved", "playlists": result})
}

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

// UpdateSinglePlaylistPublic toggles a playlist's public status
func UpdateSinglePlaylistPublic(c *gin.Context) {
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
		IsPublic bool `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	result := db.DB.Model(&db.Playlist{}).
		Where("id = ? AND owner_id = ?", playlistID, userID).
		Update("is_public", input.IsPublic)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlist", "details": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Playlist %s updated to is_public=%v", playlistID, input.IsPublic),
	})
}

// ImportToSpotify imports a playlist to Spotify
func ImportToSpotify(c *gin.Context) {
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

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var playlist db.Playlist
	if err := db.DB.Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check if user has access (Owner or Public)
	if playlist.OwnerID != userID && !playlist.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this playlist"})
		return
	}

	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
		return
	}

	client, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	spotifyID, err := services.ImportToSpotify(c.Request.Context(), client, dbUser, playlist, tracks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import to Spotify", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Playlist imported to Spotify",
		"spotify_playlist_id": spotifyID,
	})
}

// ImportToYouTube imports a playlist to YouTube
func ImportToYouTube(c *gin.Context) {
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

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var playlist db.Playlist
	if err := db.DB.Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	// Check if user has access (Owner or Public)
	if playlist.OwnerID != userID && !playlist.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to access this playlist"})
		return
	}

	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
		return
	}

	client, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
		return
	}

	youtubeID, err := services.ImportToYouTube(c.Request.Context(), client, playlist, tracks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import to YouTube", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Playlist imported to YouTube",
		"youtube_playlist_id": youtubeID,
	})
}

// ExportSpotifyToYouTube exports a Spotify playlist to YouTube
func ExportSpotifyToYouTube(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	spotifyPlaylistID := c.Param("spotifyPlaylistID")
	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	spotifyClient, err := services.GetSpotifyClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Spotify not linked"})
		return
	}

	youtubeClient, err := services.GetYouTubeClient(c.Request.Context(), dbUser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "YouTube not linked"})
		return
	}

	spotifyPlaylist, err := spotifyClient.GetPlaylist(c.Request.Context(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlist", "details": err.Error()})
		return
	}

	youtubePlaylistID, err := services.CreateYouTubePlaylist(c.Request.Context(), youtubeClient, spotifyPlaylist.Name, spotifyPlaylist.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create YouTube playlist", "details": err.Error()})
		return
	}

	tracksPage, err := spotifyClient.GetPlaylistTracks(c.Request.Context(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify tracks", "details": err.Error()})
		return
	}

	var exported, failed, skipped int
	for _, item := range tracksPage.Tracks {
		if item.Track.ID == "" {
			skipped++
			continue
		}
		videoID, err := services.SearchYouTubeVideo(c.Request.Context(), youtubeClient, item.Track.Name, strings.Join(services.GetArtists(item.Track.Artists), ", "))
		if err != nil || videoID == "" {
			failed++
			continue
		}
		if err := services.AddYouTubePlaylistItem(c.Request.Context(), youtubeClient, youtubePlaylistID, videoID); err != nil {
			failed++
		} else {
			exported++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                    "Playlist exported to YouTube",
		"spotify_playlist_name":      spotifyPlaylist.Name,
		"spotify_playlist_id":        spotifyPlaylistID,
		"youtube_playlist_id":        youtubePlaylistID,
		"total_spotify_tracks":       spotifyPlaylist.Tracks.Total,
		"tracks_exported_to_youtube": exported,
		"tracks_failed_to_export":    failed,
		"tracks_skipped_no_youtube":  skipped,
	})
}

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

// GetUserPlaylists retrieves all playlists for a user
func GetUserPlaylists(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var playlists []db.Playlist
	if err := db.DB.Where("owner_id = ?", userID).Find(&playlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch playlists", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, p := range playlists {
		result = append(result, gin.H{
			"id":          p.ID,
			"title":       p.Title,
			"description": p.Description,
			"platform":    p.Platform,
			"source_id":   p.SourceID,
			"is_public":   p.IsPublic,
			"cover_image": p.CoverImage,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Playlists retrieved", "playlists": result})
}

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
