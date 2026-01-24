package handlers

import (
	"EchoBridge/internal/auth"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
)

var WorkerPool *worker.WorkerPool

// RegisterPlaylistRoutes sets up playlist-related routes
func RegisterPlaylistRoutes(r *gin.Engine) {
	r.GET("/playlists/public", GetPublicPlaylists)
	r.GET("/playlists/:id", GetPlaylist)
	r.GET("/debug/categorize/:id", DebugCategorizePlaylist) // Debug route
	r.GET("/debug/categorize-all", DebugCategorizeAll)      // Batch categorize
	protected := r.Group("/api").Use(auth.AuthMiddleware())
	protected.GET("/spotify/playlists", SpotifyPlaylists)
	protected.GET("/spotify/playlist/:id/tracks", SpotifyPlaylistTracks)
	protected.GET("/youtube/playlists", YouTubePlaylists)
	protected.GET("/my/playlists", GetUserPlaylists)
	protected.POST("/playlists", PostPlaylist)
	protected.POST("/playlists/batch-import", BatchImportPlaylists)
	protected.PATCH("/playlists/:id/public", UpdateSinglePlaylistPublic)
	protected.POST("/playlists/:id/import", ImportPublicPlaylist) // New unified import
	protected.POST("/import/playlist/:id/to/spotify", ImportToSpotify)
	protected.POST("/import/playlist/:id/to/youtube", ImportToYouTube)
	protected.GET("/export/spotify/:spotifyPlaylistID/to/youtube", ExportSpotifyToYouTube)
	protected.POST("/sync/playlist/:id", SyncPlaylist)
	protected.GET("/sync/status/:jobID", GetSyncStatus)
}
