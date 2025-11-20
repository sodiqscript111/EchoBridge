package db

import (
	"fmt"
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
func ConnectDatabase() error {
	// UPDATE THIS STRING WITH YOUR REAL CREDENTIALS
	dsn := "host=localhost user=postgres password=password dbname=testing port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return DB.AutoMigrate(&User{}, &Playlist{}, &Track{}, &Share{}, &SyncJob{})
}
