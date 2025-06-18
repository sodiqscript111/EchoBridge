package db

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "user=postgres password=password dbname=Echobridge port=5432 sslmode=disable"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)

	err = DB.AutoMigrate(
		&User{},
		&Playlist{},
		&Track{},
		&Export{},
		&Like{},
		&Tag{},
		&PlaylistTag{},
	)
	if err != nil {
		log.Fatal("❌ AutoMigration failed:", err)
	}

	// Add indexes for performance
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_playlist_source_id ON playlists(source_id);")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_track_spotify_id ON tracks(spotify_id);")

	fmt.Println("✅ Database connected and models migrated successfully.")
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email        string    `gorm:"uniqueIndex"`
	Username     string
	Platform     string
	SpotifyID    string `gorm:"uniqueIndex"`
	YouTubeID    string `gorm:"uniqueIndex"`
	SpotifyToken string // JSON-encoded oauth2.Token
	YouTubeToken string // JSON-encoded oauth2.Token
	TokenExpiry  time.Time
	RefreshToken string
	CreatedAt    time.Time
}

type Playlist struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OwnerID     uuid.UUID
	Owner       User `gorm:"foreignKey:OwnerID"`
	Title       string
	Description string
	Platform    string // spotify, youtube, custom
	SourceID    string
	IsPublic    bool
	CoverImage  string
	CreatedAt   time.Time
	Tracks      []Track
}

type Track struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	PlaylistID uuid.UUID
	Playlist   Playlist `gorm:"foreignKey:PlaylistID"`
	Title      string
	Artist     string
	Album      string
	PreviewURL string
	SpotifyID  string
	YouTubeID  string
	CreatedAt  time.Time
}

type Export struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     uuid.UUID
	User       User `gorm:"foreignKey:UserID"`
	PlaylistID uuid.UUID
	Playlist   Playlist `gorm:"foreignKey:PlaylistID"`
	ExportedTo string
	ExportedAt time.Time
}

type Like struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     uuid.UUID
	PlaylistID uuid.UUID
	CreatedAt  time.Time
}

type Tag struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name string    `gorm:"unique"`
}

type PlaylistTag struct {
	PlaylistID uuid.UUID
	TagID      uuid.UUID
}
