package handlers

import (
	"EchoBridge/internal/auth"

	"github.com/gin-gonic/gin"
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
}
