package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"EchoBridge/db"
	"EchoBridge/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
)

// CONFIGURATION
var FrontendURL = "http://localhost:5173"

// GLOBAL VARIABLES (Initialized in init())
var (
	jwtSecret          = []byte("your-secret-key") // Replace with a secure key later
	googleOAuthConfig  *oauth2.Config
	spotifyAuth        *spotifyauth.Authenticator
	youtubeOAuthConfig *oauth2.Config
	state              = "random-state-string"
	WorkerPool         *worker.WorkerPool
)

// Initialize Credentials Here
func init() {
	if url := os.Getenv("FRONTEND_URL"); url != "" {
		FrontendURL = url
	}

	// --- SPOTIFY CREDENTIALS ---
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyRedirectURL := os.Getenv("SPOTIFY_REDIRECT_URL")
	if spotifyRedirectURL == "" {
		spotifyRedirectURL = "http://127.0.0.1:8000/callback/spotify"
	}

	// --- GOOGLE / YOUTUBE CREDENTIALS ---
	youtubeClientID := os.Getenv("YOUTUBE_CLIENT_ID")
	youtubeClientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")

	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		googleRedirectURL = "http://127.0.0.1:8000/callback/google"
	}
	youtubeRedirectURL := os.Getenv("YOUTUBE_REDIRECT_URL")
	if youtubeRedirectURL == "" {
		youtubeRedirectURL = "http://127.0.0.1:8000/callback/youtube"
	}

	// 1. Setup Spotify
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

	// 2. Setup Google Login Config
	googleOAuthConfig = &oauth2.Config{
		ClientID:     youtubeClientID,     // Using same ID
		ClientSecret: youtubeClientSecret, // Using same Secret
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	// 3. Setup YouTube Sync Config
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

// --- MIDDLEWARE ---

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[AuthMiddleware] Request: %s %s\n", c.Request.Method, c.Request.URL.Path)
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Fallback to query parameter "token"
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			c.Abort()
			return
		}

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

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in JWT"})
			c.Abort()
			return
		}

		c.Set("userID", userIDStr)
		c.Next()
	}
}

func GenerateJWT(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days expiry
	})
	return token.SignedString(jwtSecret)
}

// --- HANDLERS ---

// GoogleLogin initiates Google OAuth flow
func GoogleLogin(c *gin.Context) {
	// Request Offline Access to get a Refresh Token
	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	c.Redirect(http.StatusFound, url)
}

// GoogleCallback handles Google OAuth callback
func GoogleCallback(c *gin.Context) {
	if authError := c.Request.URL.Query().Get("error"); authError != "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Google authorization failed", "details": authError})
		return
	}

	token, err := googleOAuthConfig.Exchange(c.Request.Context(), c.Request.URL.Query().Get("code"))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to obtain Google token", "details": err.Error()})
		return
	}

	client := googleOAuthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Google user info", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode Google user info", "details": err.Error()})
		return
	}

	dbUser, err := saveUser(c.Request.Context(), "google", userInfo.ID, userInfo.Email, userInfo.Name, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user", "details": err.Error()})
		return
	}

	jwtToken, err := GenerateJWT(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT", "details": err.Error()})
		return
	}

	// Return JSON for Postman Testing
	c.JSON(http.StatusOK, gin.H{
		"message":   "Google Login Successful",
		"jwt_token": jwtToken,
		"user_id":   dbUser.ID,
		"note":      "Copy the jwt_token to use in Postman Authorization header",
	})
}

// EmailLogin handles email/password login
func EmailLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var dbUser db.User
	if err := db.DB.Where("email = ?", input.Email).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	jwtToken, err := GenerateJWT(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Email authentication successful",
		"user_id":   dbUser.ID,
		"jwt_token": jwtToken,
	})
}

// EmailRegister handles email/password registration
func EmailRegister(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var existingUser db.User
	if err := db.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password", "details": err.Error()})
		return
	}

	dbUser := db.User{
		ID:        uuid.New(),
		Email:     input.Email,
		Username:  input.Username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}
	if err := db.DB.Create(&dbUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	jwtToken, err := GenerateJWT(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Registration successful",
		"user_id":   dbUser.ID,
		"jwt_token": jwtToken,
	})
}

// SpotifyLink initiates Spotify OAuth flow
func SpotifyLink(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}
	// Use UserID as state to identify user in callback
	url := spotifyAuth.AuthURL(userID.String())
	c.Redirect(http.StatusFound, url)
}

// SpotifyCallback handles Spotify OAuth callback
func SpotifyCallback(c *gin.Context) {
	stateUserID := c.Request.URL.Query().Get("state")
	if stateUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state parameter"})
		return
	}
	userID, err := uuid.Parse(stateUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	token, err := spotifyAuth.Token(c.Request.Context(), stateUserID, c.Request)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to get token", "details": err.Error()})
		return
	}

	client := spotify.New(spotifyAuth.Client(c.Request.Context(), token))
	user, err := client.CurrentUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info", "details": err.Error()})
		return
	}

	if err := updateUserService(c.Request.Context(), userID, "spotify", user.ID, user.Email, user.DisplayName, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Spotify", "details": err.Error()})
		return
		return
	}

	// Redirect to settings with success parameter
	c.Redirect(http.StatusFound, FrontendURL+"/settings?connected=spotify")
}

// YouTubeLink initiates YouTube OAuth flow
func YouTubeLink(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}
	// Use UserID as state
	// Use AccessTypeOffline to ensure Refresh Token
	url := youtubeOAuthConfig.AuthCodeURL(userID.String(), oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	c.Redirect(http.StatusFound, url)
}

// YouTubeCallback handles YouTube OAuth callback
func YouTubeCallback(c *gin.Context) {
	stateUserID := c.Request.URL.Query().Get("state")
	if stateUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state parameter"})
		return
	}
	userID, err := uuid.Parse(stateUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	code := c.Request.URL.Query().Get("code")
	token, err := youtubeOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to get token", "details": err.Error()})
		return
	}

	service, err := youtube.NewService(c.Request.Context(), option.WithTokenSource(youtubeOAuthConfig.TokenSource(c.Request.Context(), token)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create YouTube client", "details": err.Error()})
		return
	}

	call := service.Channels.List([]string{"id", "snippet"}).Mine(true)
	response, err := call.Do()
	if err != nil {
		fmt.Printf("YouTube API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channel info", "details": err.Error()})
		return
	}
	if len(response.Items) == 0 {
		fmt.Println("YouTube API Warning: No channel found for this user. User might not have a YouTube channel created.")
		c.JSON(http.StatusNotFound, gin.H{"error": "No YouTube channel found. Please create a channel on YouTube first."})
		return
	}

	channel := response.Items[0]
	if err := updateUserService(c.Request.Context(), userID, "youtube", channel.Id, "", channel.Snippet.Title, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link YouTube", "details": err.Error()})
		return
		return
	}

	// Redirect to settings with success parameter
	c.Redirect(http.StatusFound, FrontendURL+"/settings?connected=youtube")
}

// --- HELPER FUNCTIONS ---

func saveUser(ctx context.Context, authType, authID, email, username string, token *oauth2.Token) (db.User, error) {
	var dbUser db.User
	tokenJSON, _ := json.Marshal(token)

	query := db.DB.Where("email = ?", email).First(&dbUser)

	if query.Error == gorm.ErrRecordNotFound {
		dbUser = db.User{
			ID:        uuid.New(),
			Email:     email,
			Username:  username,
			AuthType:  authType,
			AuthID:    authID,
			CreatedAt: time.Now(),
		}
		if token != nil {
			dbUser.GoogleToken = string(tokenJSON)
			dbUser.RefreshToken = token.RefreshToken // Save initial refresh token
			dbUser.TokenExpiry = token.Expiry
		}
		if err := db.DB.Create(&dbUser).Error; err != nil {
			return dbUser, fmt.Errorf("failed to create user: %w", err)
		}
	} else if query.Error != nil {
		return dbUser, fmt.Errorf("database error: %w", query.Error)
	} else {
		dbUser.Email = email
		dbUser.Username = username
		if token != nil {
			dbUser.GoogleToken = string(tokenJSON)
			dbUser.TokenExpiry = token.Expiry
			// Only overwrite RefreshToken if the new one is NOT empty
			if token.RefreshToken != "" {
				dbUser.RefreshToken = token.RefreshToken
			}
		}
		if err := db.DB.Save(&dbUser).Error; err != nil {
			return dbUser, fmt.Errorf("failed to update user: %w", err)
		}
	}
	return dbUser, nil
}

func updateUserService(ctx context.Context, userID uuid.UUID, platform, platformID, email, username string, token *oauth2.Token) error {
	var dbUser db.User
	if err := db.DB.Where("id = ?", userID).First(&dbUser).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	updateData := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if platform == "spotify" {
		updateData["spotify_id"] = platformID
		updateData["spotify_token"] = string(tokenJSON)
		updateData["spotify_token_expiry"] = token.Expiry
		// Check for empty refresh token
		if token.RefreshToken != "" {
			updateData["spotify_refresh_token"] = token.RefreshToken
		}
		if email != "" {
			updateData["email"] = email
		}
		if username != "" {
			updateData["username"] = username
		}
	} else if platform == "youtube" {
		updateData["youtube_id"] = platformID
		updateData["youtube_token"] = string(tokenJSON)
		updateData["youtube_token_expiry"] = token.Expiry
		// Check for empty refresh token
		if token.RefreshToken != "" {
			updateData["youtube_refresh_token"] = token.RefreshToken
		}
	}

	if err := db.DB.Model(&dbUser).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update %s: %w", platform, err)
	}
	return nil
}
