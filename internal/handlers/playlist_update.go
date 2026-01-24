package handlers

import (
	"fmt"
	"net/http"

	"EchoBridge/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
