package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Models
type ChatSession struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Mode      string    `json:"mode"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	Messages  []Message `gorm:"foreignKey:SessionID" json:"messages"`
}

type Message struct {
	ID        string   `gorm:"primaryKey" json:"id"`
	SessionID string   `json:"session_id"`
	Role      string   `json:"role"` // user, assistant, system
	Content   string   `json:"content"`
	Timestamp int64    `json:"timestamp"`
	Sources   []Source `gorm:"foreignKey:MessageID" json:"sources,omitempty"`
	Reasoning string   `json:"reasoning,omitempty"`
}

type Source struct {
	ID          uint    `gorm:"primaryKey" json:"-"`
	MessageID   string  `json:"-"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Snippet     string  `json:"snippet"`
	Credibility float64 `json:"credibility,omitempty"`
}

func InitDB(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if strings.HasPrefix(databaseURL, "postgres") || strings.HasPrefix(databaseURL, "postgresql") {
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// Default to SQLite
		dbPath := strings.TrimPrefix(databaseURL, "sqlite://")
		if dbPath == databaseURL {
			dbPath = "research_pro.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&ChatSession{},
		&Message{},
		&Source{},
	)
}
