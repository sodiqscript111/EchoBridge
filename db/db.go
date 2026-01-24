package db

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User represents a user in the database
// User represents a user in the database
type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	AuthType string    // "google" or "email"
	AuthID   string    // Google ID or empty for email
	Email    string    `gorm:"uniqueIndex"`
	Username string
	Password string

	// EXPLICIT TAGS ADDED HERE:
	SpotifyID    string `gorm:"column:spotify_id"`
	YouTubeID    string `gorm:"column:youtube_id"`
	AppleMusicID string `gorm:"column:applemusic_id"`

	GoogleToken     string `gorm:"column:google_token"`
	SpotifyToken    string `gorm:"column:spotify_token"`
	YouTubeToken    string `gorm:"column:youtube_token"`
	AppleMusicToken string `gorm:"column:applemusic_token"`

	SpotifyTokenExpiry  time.Time `gorm:"column:spotify_token_expiry"`
	SpotifyRefreshToken string    `gorm:"column:spotify_refresh_token"`

	YouTubeTokenExpiry  time.Time `gorm:"column:youtube_token_expiry"`
	YouTubeRefreshToken string    `gorm:"column:youtube_refresh_token"`

	TokenExpiry  time.Time
	RefreshToken string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// Playlist represents a playlist in the database
type Playlist struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID     uuid.UUID `gorm:"type:uuid"`
	Title       string
	Description string
	Platform    string
	SourceID    string
	IsPublic    bool
	CoverImage  string
	Category    string // New field for AI categorization
	CreatedAt   time.Time
	Tracks      []Track `gorm:"foreignKey:PlaylistID"`
}

// Track represents a track in the database
type Track struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	PlaylistID   uuid.UUID `gorm:"type:uuid"`
	Title        string
	Artist       string
	Album        string
	SpotifyID    string
	YouTubeID    string
	AppleMusicID string
	PreviewURL   string // URL to 30s preview (from Spotify)
	CreatedAt    time.Time
}

// Share represents a shared track link
type Share struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	TrackID   uuid.UUID `gorm:"type:uuid"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	CustomURL string
	CreatedAt time.Time
}

// SyncJob represents a background sync job
type SyncJob struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid"`
	PlaylistID  uuid.UUID `gorm:"type:uuid"`
	Platforms   string    // JSON array: ["spotify", "youtube"]
	Status      string    // "pending", "processing", "completed", "failed"
	Result      string    // JSON: {"spotify": "playlist_id", "youtube": "playlist_id"}
	ErrorMsg    string    // Error message if failed
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// ConnectDatabase initializes the database connection
// ConnectDatabase initializes the database connection
func ConnectDatabase() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback to building from individual vars
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")
		sslmode := os.Getenv("DB_SSLMODE")

		if host == "" || user == "" {
			return fmt.Errorf("DATABASE_URL or DB_HOST/DB_USER not set")
		}
		if port == "" {
			port = "5432"
		}
		if sslmode == "" {
			sslmode = "require"
		}

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, dbname, port, sslmode)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return DB.AutoMigrate(&User{}, &Playlist{}, &Track{}, &Share{}, &SyncJob{})
}
