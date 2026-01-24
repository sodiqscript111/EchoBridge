package handlers

import (
	"net/http"

	"EchoBridge/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetPublicPlaylists retrieves all public playlists
func GetPublicPlaylists(c *gin.Context) {
	var playlists []db.Playlist
	if err := db.DB.Where("is_public = ?", true).Find(&playlists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public playlists", "details": err.Error()})
		return
	}

	result := []gin.H{}
	for _, p := range playlists {
		result = append(result, gin.H{
			"id":          p.ID,
			"title":       p.Title,
			"description": p.Description,
			"platform":    p.Platform,
			"owner_id":    p.OwnerID,
			"cover_image": p.CoverImage,
			"category":    p.Category,
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Public playlists retrieved", "playlists": result})
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
