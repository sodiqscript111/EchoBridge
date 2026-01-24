package handlers

import (
	"net/http"

	"EchoBridge/db"
	"EchoBridge/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterAuthRoutes sets up authentication routes
func RegisterAuthRoutes(r *gin.Engine) {
	r.POST("/register", auth.EmailRegister)
	r.POST("/login/email", auth.EmailLogin)
	r.GET("/login/google", auth.GoogleLogin)
	r.GET("/callback/google", auth.GoogleCallback)
	r.GET("/callback/spotify", auth.SpotifyCallback)
	r.GET("/callback/youtube", auth.YouTubeCallback)
	protected := r.Group("/api").Use(auth.AuthMiddleware())
	protected.GET("/link/spotify", auth.SpotifyLink)
	protected.GET("/link/youtube", auth.YouTubeLink)
	protected.GET("/connection/status", GetConnectionStatus)
}

// GetConnectionStatus checks which platforms are connected
func GetConnectionStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var user db.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"spotify_connected": user.SpotifyToken != "",
		"youtube_connected": user.YouTubeToken != "",
		"username":          user.Username,
	})
}
