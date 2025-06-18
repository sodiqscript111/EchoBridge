package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"

	"EchoBridge/db" // Replace with your actual module path
)

var (
	spotifyRedirectURL = "http://127.0.0.1:8000/callback/spotify"
	spotifyAuth        *spotifyauth.Authenticator
	youtubeRedirectURL = "http://127.0.0.1:8000/callback/youtube"
	youtubeOAuthConfig *oauth2.Config
	state              = "random-state-string"     // In production, generate dynamically
	jwtSecret          = []byte("your-secret-key") // Replace with a secure key
)

func init() {
	spotifyClientID := "6f55af09467d44319c84abf777042886"
	spotifyClientSecret := "879b791b755c4a9cbb97254bee4d3917"
	youtubeClientID := "1055231814715-mkfdshbi2l4rkmvn39l4f530o61o6c4u.apps.googleusercontent.com"
	youtubeClientSecret := "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"

	spotifyAuth = spotifyauth.New(
		spotifyauth.WithClientID(spotifyClientID),
		spotifyauth.WithClientSecret(spotifyClientSecret),
		spotifyauth.WithRedirectURL(spotifyRedirectURL),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopePlaylistModifyPublic,
		),
	)

	youtubeOAuthConfig = &oauth2.Config{
		ClientID:     youtubeClientID,
		ClientSecret: youtubeClientSecret,
		RedirectURL:  youtubeRedirectURL,
		Scopes: []string{
			youtube.YoutubeReadonlyScope,
			youtube.YoutubeScope,
		},
		Endpoint: google.Endpoint,
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT", "details": err.Error()})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in JWT"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func generateJWT(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func main() {
	db.ConnectDatabase()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	r.GET("/login/spotify", spotifyLoginHandler)
	r.GET("/callback/spotify", spotifyCallbackHandler)
	r.GET("/login/youtube", youtubeLoginHandler)
	r.GET("/callback/youtube", youtubeCallbackHandler)
	r.GET("/playlists/public", getPublicPlaylistsHandler)
	r.PATCH("/playlists/public", updatePlaylistsPublicHandler)

	protected := r.Group("/api").Use(authMiddleware())
	{
		protected.GET("/spotify/playlist/:id/tracks", spotifyPlaylistTracksHandler)
		protected.GET("/youtube/playlists", youtubePlaylistsHandler)
		protected.GET("/my/playlists", getUserPlaylistsHandler)

		protected.POST("/playlists", postPlaylistHandler)
		protected.PATCH("/playlists/:id/public", updateSinglePlaylistPublicHandler)
		protected.POST("/import/playlist/:id/to/spotify", importToSpotifyHandler)
		protected.GET("/export/spotify/:spotifyPlaylistID/to/youtube", exportSpotifyToYouTubeHandler)
	}

	r.Run(":8000")
}

// New handler to toggle a single playlist's public status
func updateSinglePlaylistPublicHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	playlistIDStr := c.Param("id")
	playlistID, err := uuid.Parse(playlistIDStr)
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

func updatePlaylistsPublicHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		PlaylistIDs []string `json:"playlist_ids" binding:"required"`
		IsPublic    bool     `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var playlistUUIDs []uuid.UUID
	for _, id := range input.PlaylistIDs {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid playlist ID", "details": id})
			return
		}
		playlistUUIDs = append(playlistUUIDs, parsedID)
	}

	result := db.DB.Model(&db.Playlist{}).
		Where("id IN ? AND owner_id = ?", playlistUUIDs, userID).
		Update("is_public", input.IsPublic)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update playlists", "details": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching playlists found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Updated %d playlist(s) to is_public=%v", result.RowsAffected, input.IsPublic),
	})
}

func spotifyLoginHandler(c *gin.Context) {
	url := spotifyAuth.AuthURL(state)
	c.Redirect(http.StatusFound, url)
}

func spotifyCallbackHandler(c *gin.Context) {
	r := c.Request
	token, err := spotifyAuth.Token(r.Context(), state, r)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to obtain Spotify access token", "details": err.Error()})
		return
	}

	client := spotify.New(spotifyAuth.Client(r.Context(), token))
	user, err := client.CurrentUser(r.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify user profile", "details": err.Error()})
		return
	}

	var dbUser db.User
	dbResult := db.DB.Where("spotify_id = ?", user.ID).First(&dbUser)
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode token", "details": err.Error()})
		return
	}

	if dbResult.Error == gorm.ErrRecordNotFound {
		dbUser = db.User{
			ID:           uuid.New(),
			SpotifyID:    user.ID,
			Email:        user.Email,
			Username:     user.DisplayName,
			Platform:     "spotify",
			SpotifyToken: string(tokenJSON),
			TokenExpiry:  token.Expiry,
			RefreshToken: token.RefreshToken,
			CreatedAt:    time.Now(),
		}
		if err := db.DB.Create(&dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}
	} else if dbResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": dbResult.Error.Error()})
		return
	} else {
		// Update existing user
		dbUser.SpotifyToken = string(tokenJSON)
		dbUser.TokenExpiry = token.Expiry
		dbUser.RefreshToken = token.RefreshToken
		dbUser.Email = user.Email
		dbUser.Username = user.DisplayName
		if err := db.DB.Save(&dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
			return
		}
	}

	jwtToken, err := generateJWT(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT", "details": err.Error()})
		return
	}

	playlists, err := client.CurrentUsersPlaylists(r.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlists", "details": err.Error()})
		return
	}

	var result []gin.H
	for _, p := range playlists.Playlists {
		fullPlaylist, err := client.GetPlaylist(r.Context(), p.ID)
		if err != nil {
			continue
		}
		var coverImage string
		if len(fullPlaylist.Images) > 0 {
			coverImage = fullPlaylist.Images[0].URL
		}
		playlist := db.Playlist{
			ID:          uuid.New(),
			OwnerID:     dbUser.ID,
			Title:       p.Name,
			Description: p.Description,
			Platform:    "spotify",
			SourceID:    p.ID.String(),
			IsPublic:    false,
			CoverImage:  coverImage,
			CreatedAt:   time.Now(),
		}
		if err := db.DB.Create(&playlist).Error; err != nil {
			continue
		}

		tracksPage, err := client.GetPlaylistTracks(context.Background(), p.ID)
		if err != nil {
			continue
		}
		for _, item := range tracksPage.Tracks {
			if item.Track.ID != "" {
				track := db.Track{
					ID:         uuid.New(),
					PlaylistID: playlist.ID,
					Title:      item.Track.Name,
					Artist:     strings.Join(getArtists(item.Track.Artists), ", "),
					Album:      item.Track.Album.Name,
					SpotifyID:  item.Track.ID.String(),
					CreatedAt:  time.Now(),
				}
				db.DB.Create(&track)
			}
		}

		result = append(result, gin.H{
			"platform":    "spotify",
			"id":          playlist.ID,
			"name":        p.Name,
			"cover_image": coverImage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully authenticated with Spotify",
		"user_id":   dbUser.ID,
		"jwt_token": jwtToken,
		"playlists": result,
	})
}

func youtubeLoginHandler(c *gin.Context) {
	url := youtubeOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusFound, url)
}

func youtubeCallbackHandler(c *gin.Context) {
	r := c.Request
	if authError := r.URL.Query().Get("error"); authError != "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":       "YouTube authorization failed",
			"details":     authError,
			"description": r.URL.Query().Get("error_description"),
		})
		return
	}

	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	token, err := youtubeOAuthConfig.Exchange(r.Context(), authCode)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to obtain YouTube access token", "details": err.Error()})
		return
	}

	httpClient := youtubeOAuthConfig.Client(context.Background(), token)
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize YouTube service client", "details": err.Error()})
		return
	}

	channelCall := service.Channels.List([]string{"id"}).Mine(true)
	channelResponse, err := channelCall.Do()
	if err != nil || len(channelResponse.Items) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve YouTube user ID", "details": err.Error()})
		return
	}
	youtubeUserID := channelResponse.Items[0].Id

	var dbUser db.User
	dbResult := db.DB.Where("youtube_id = ?", youtubeUserID).First(&dbUser)
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode token", "details": err.Error()})
		return
	}

	if dbResult.Error == gorm.ErrRecordNotFound {
		dbUser = db.User{
			ID:           uuid.New(),
			YouTubeID:    youtubeUserID,
			Platform:     "youtube",
			YouTubeToken: string(tokenJSON),
			TokenExpiry:  token.Expiry,
			RefreshToken: token.RefreshToken,
			CreatedAt:    time.Now(),
		}
		if err := db.DB.Create(&dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
			return
		}
	} else if dbResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": dbResult.Error.Error()})
		return
	} else {
		// Update existing user
		dbUser.YouTubeToken = string(tokenJSON)
		dbUser.TokenExpiry = token.Expiry
		dbUser.RefreshToken = token.RefreshToken
		if err := db.DB.Save(&dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user", "details": err.Error()})
			return
		}
	}

	jwtToken, err := generateJWT(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT", "details": err.Error()})
		return
	}

	playlistCall := service.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true)
	playlistResponse, err := playlistCall.Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve YouTube playlists", "details": err.Error()})
		return
	}

	var result []gin.H
	for _, playlist := range playlistResponse.Items {
		if playlist.Id != "" && playlist.Snippet != nil && playlist.Snippet.Title != "" {
			var coverImage string
			if playlist.Snippet.Thumbnails != nil && playlist.Snippet.Thumbnails.Default != nil {
				coverImage = playlist.Snippet.Thumbnails.Default.Url
			}
			dbPlaylist := db.Playlist{
				ID:          uuid.New(),
				OwnerID:     dbUser.ID,
				Title:       playlist.Snippet.Title,
				Description: playlist.Snippet.Description,
				Platform:    "youtube",
				SourceID:    playlist.Id,
				IsPublic:    false,
				CoverImage:  coverImage,
				CreatedAt:   time.Now(),
			}
			if err := db.DB.Create(&dbPlaylist).Error; err != nil {
				continue
			}

			result = append(result, gin.H{
				"platform":    "youtube",
				"id":          dbPlaylist.ID,
				"title":       playlist.Snippet.Title,
				"description": playlist.Snippet.Description,
				"item_count":  playlist.ContentDetails.ItemCount,
				"cover_image": coverImage,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully authenticated with YouTube",
		"user_id":   dbUser.ID,
		"jwt_token": jwtToken,
		"playlists": result,
	})
}

func spotifyPlaylistTracksHandler(c *gin.Context) {
	playlistID := c.Param("id")
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(dbUser.SpotifyToken), &token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Spotify token"})
		return
	}
	client := spotify.New(spotifyAuth.Client(context.Background(), &token))

	tracksPage, err := client.GetPlaylistTracks(context.Background(), spotify.ID(playlistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlist tracks", "details": err.Error()})
		return
	}

	var tracks []gin.H
	for _, item := range tracksPage.Tracks {
		if item.Track.ID != "" {
			tracks = append(tracks, gin.H{
				"name":    item.Track.Name,
				"artists": getArtists(item.Track.Artists),
				"album":   item.Track.Album.Name,
				"uri":     item.Track.URI,
				"id":      item.Track.ID.String(),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"platform":    "spotify",
		"playlist_id": playlistID,
		"tracks":      tracks,
	})
}

func youtubePlaylistsHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(dbUser.YouTubeToken), &token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid YouTube token"})
		return
	}
	httpClient := youtubeOAuthConfig.Client(context.Background(), &token)
	service, err := youtube.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize YouTube service", "details": err.Error()})
		return
	}

	call := service.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true)
	response, err := call.Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve YouTube playlists", "details": err.Error()})
		return
	}

	var result []gin.H
	for _, playlist := range response.Items {
		if playlist.Id != "" && playlist.Snippet != nil && playlist.Snippet.Title != "" {
			var coverImage string
			if playlist.Snippet.Thumbnails != nil && playlist.Snippet.Thumbnails.Default != nil {
				coverImage = playlist.Snippet.Thumbnails.Default.Url
			}
			result = append(result, gin.H{
				"platform":    "youtube",
				"id":          playlist.Id,
				"title":       playlist.Snippet.Title,
				"description": playlist.Snippet.Description,
				"item_count":  playlist.ContentDetails.ItemCount,
				"cover_image": coverImage,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully retrieved YouTube playlists",
		"playlists": result,
	})
}

func postPlaylistHandler(c *gin.Context) {
	var input struct {
		Platform string `json:"platform" binding:"required"`
		SourceID string `json:"source_id" binding:"required"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
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
	if input.Platform == "spotify" {
		var token oauth2.Token
		if err := json.Unmarshal([]byte(dbUser.SpotifyToken), &token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Spotify token"})
			return
		}
		client := spotify.New(spotifyAuth.Client(context.Background(), &token))
		spPlaylist, err := client.GetPlaylist(context.Background(), spotify.ID(input.SourceID))
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to fetch Spotify playlist", "details": err.Error()})
			return
		}
		var coverImage string
		if len(spPlaylist.Images) > 0 {
			coverImage = spPlaylist.Images[0].URL
		}
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

		tracksPage, err := client.GetPlaylistTracks(context.Background(), spotify.ID(input.SourceID))
		if err != nil {
		}
		for _, item := range tracksPage.Tracks {
			if item.Track.ID != "" {
				track := db.Track{
					ID:         uuid.New(),
					PlaylistID: playlist.ID,
					Title:      item.Track.Name,
					Artist:     strings.Join(getArtists(item.Track.Artists), ", "),
					Album:      item.Track.Album.Name,
					SpotifyID:  item.Track.ID.String(),
					CreatedAt:  time.Now(),
				}
				db.DB.Create(&track)
			}
		}
	} else if input.Platform == "youtube" {
		var token oauth2.Token
		if err := json.Unmarshal([]byte(dbUser.YouTubeToken), &token); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid YouTube token"})
			return
		}
		httpClient := youtubeOAuthConfig.Client(context.Background(), &token)
		service, err := youtube.NewService(context.Background(), option.WithHTTPClient(httpClient))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize YouTube service"})
			return
		}
		call := service.Playlists.List([]string{"id", "snippet"}).Id(input.SourceID)
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

		itemsCall := service.PlaylistItems.List([]string{"snippet"}).PlaylistId(input.SourceID).MaxResults(50)
		itemsResponse, err := itemsCall.Do()
		if err != nil {
		}
		for _, item := range itemsResponse.Items {
			if item.Snippet != nil && item.Snippet.ResourceId != nil && item.Snippet.ResourceId.VideoId != "" {
				track := db.Track{
					ID:         uuid.New(),
					PlaylistID: playlist.ID,
					Title:      item.Snippet.Title,
					Artist:     "",
					YouTubeID:  item.Snippet.ResourceId.VideoId,
					CreatedAt:  time.Now(),
				}
				db.DB.Create(&track)
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported platform"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Playlist posted successfully",
		"playlist_id": playlist.ID,
		"cover_image": playlist.CoverImage,
	})
}

func getUserPlaylistsHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
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

	c.JSON(http.StatusOK, gin.H{
		"message":   "Playlists retrieved successfully",
		"playlists": result,
	})
}

func getPublicPlaylistsHandler(c *gin.Context) {
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
				"title":      t.Title,
				"artist":     t.Artist,
				"album":      t.Album,
				"spotify_id": t.SpotifyID,
				"youtube_id": t.YouTubeID,
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
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Public playlists retrieved successfully",
		"playlists": result,
	})
}

func importToSpotifyHandler(c *gin.Context) {
	playlistID := c.Param("id")
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
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
	if err := db.DB.Where("id = ?", playlistID).First(&playlist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Playlist not found"})
		return
	}

	var tracks []db.Track
	if err := db.DB.Where("playlist_id = ?", playlistID).Find(&tracks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tracks"})
		return
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(dbUser.SpotifyToken), &token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Spotify token"})
		return
	}
	client := spotify.New(spotifyAuth.Client(context.Background(), &token))

	spPlaylist, err := client.CreatePlaylistForUser(context.Background(), dbUser.SpotifyID, playlist.Title, playlist.Description, playlist.IsPublic, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Spotify playlist", "details": err.Error()})
		return
	}

	var trackIDs []spotify.ID
	for _, t := range tracks {
		if t.SpotifyID != "" {
			trackIDs = append(trackIDs, spotify.ID(t.SpotifyID))
		}
	}
	if len(trackIDs) > 0 {
		_, err = client.AddTracksToPlaylist(context.Background(), spPlaylist.ID, trackIDs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tracks to Spotify playlist", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Playlist imported to Spotify successfully",
		"spotify_playlist_id": spPlaylist.ID,
	})
}

func exportSpotifyToYouTubeHandler(c *gin.Context) {
	spotifyPlaylistID := c.Param("spotifyPlaylistID")
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var spotifyToken oauth2.Token
	if err := json.Unmarshal([]byte(dbUser.SpotifyToken), &spotifyToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Spotify token"})
		return
	}
	spotifyClient := spotify.New(spotifyAuth.Client(context.Background(), &spotifyToken))

	var youtubeToken oauth2.Token
	if err := json.Unmarshal([]byte(dbUser.YouTubeToken), &youtubeToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid YouTube token"})
		return
	}
	httpClient := youtubeOAuthConfig.Client(context.Background(), &youtubeToken)
	youtubeService, err := youtube.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize YouTube service", "details": err.Error()})
		return
	}

	spotifyPlaylist, err := spotifyClient.GetPlaylist(context.Background(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlist details", "details": err.Error()})
		return
	}

	youtubePlaylistID, err := findYouTubePlaylistByName(context.Background(), youtubeService, spotifyPlaylist.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing YouTube playlist", "details": err.Error()})
		return
	}

	if youtubePlaylistID == "" {
		youtubePlaylistID, err = createYouTubePlaylist(context.Background(), youtubeService, spotifyPlaylist.Name, spotifyPlaylist.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create YouTube playlist", "details": err.Error()})
			return
		}
	}

	tracksPage, err := spotifyClient.GetPlaylistTracks(context.Background(), spotify.ID(spotifyPlaylistID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Spotify playlist tracks", "details": err.Error()})
		return
	}

	var exportedTrackCount, failedTrackCount, skippedTrackCount int
	for page := 1; ; page++ {
		for _, item := range tracksPage.Tracks {
			if item.Track.ID == "" {
				skippedTrackCount++
				continue
			}

			trackName := item.Track.Name
			artistNames := strings.Join(getArtists(item.Track.Artists), ", ")

			youtubeVideoID, err := searchYouTubeVideo(context.Background(), youtubeService, trackName, artistNames)
			if err != nil {
				failedTrackCount++
				continue
			}

			if youtubeVideoID != "" {
				err = addYouTubePlaylistItem(context.Background(), youtubeService, youtubePlaylistID, youtubeVideoID)
				if err != nil {
					failedTrackCount++
				} else {
					exportedTrackCount++
				}
			} else {
				skippedTrackCount++
			}
		}

		if tracksPage.Next == "" {
			break
		}
		tracksPage, err = spotifyClient.GetPlaylistTracks(
			context.Background(),
			spotify.ID(spotifyPlaylistID),
			spotify.Offset(int(tracksPage.Offset)+len(tracksPage.Tracks)),
		)
		if err != nil {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                    "Playlist export completed successfully",
		"spotify_playlist_name":      spotifyPlaylist.Name,
		"spotify_playlist_id":        spotifyPlaylistID,
		"youtube_playlist_id":        youtubePlaylistID,
		"total_spotify_tracks":       spotifyPlaylist.Tracks.Total,
		"tracks_exported_to_youtube": exportedTrackCount,
		"tracks_failed_to_export":    failedTrackCount,
		"tracks_skipped_no_youtube":  skippedTrackCount,
	})
}

func searchYouTubeVideo(ctx context.Context, service *youtube.Service, title, artist string) (string, error) {
	searchQuery := title + " " + artist + " official music video"
	call := service.Search.List([]string{"id", "snippet"}).
		Q(searchQuery).
		MaxResults(1).
		Type("video")
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	if len(response.Items) > 0 {
		return response.Items[0].Id.VideoId, nil
	}
	return "", nil
}

func createYouTubePlaylist(ctx context.Context, service *youtube.Service, playlistName, description string) (string, error) {
	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlistName,
			Description: description,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "private",
		},
	}
	call := service.Playlists.Insert([]string{"snippet", "status"}, playlist)
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	return response.Id, nil
}

func addYouTubePlaylistItem(ctx context.Context, service *youtube.Service, playlistID, videoID string) error {
	playlistItem := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}
	call := service.PlaylistItems.Insert([]string{"snippet"}, playlistItem)
	_, err := call.Do()
	return err
}

func findYouTubePlaylistByName(ctx context.Context, service *youtube.Service, playlistName string) (string, error) {
	call := service.Playlists.List([]string{"id", "snippet"}).Mine(true)
	response, err := call.Do()
	if err != nil {
		return "", err
	}
	for _, playlist := range response.Items {
		if playlist.Snippet != nil && playlist.Snippet.Title == playlistName {
			return playlist.Id, nil
		}
	}
	return "", nil
}

func getArtists(artists []spotify.SimpleArtist) []string {
	var names []string
	for _, a := range artists {
		names = append(names, a.Name)
	}
	return names
}
