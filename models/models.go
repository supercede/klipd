package models

import (
	"time"

	"gorm.io/gorm"
)

// clipboard history item
type ClipboardItem struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	ContentType   string    `gorm:"not null" json:"contentType"` // "text", "image", "file"
	ContentText   string    `json:"content"`                     // For text content
	ContentBinary []byte    `json:"-"`                           // For binary content (images, etc.)
	PreviewText   string    `json:"preview"`                     // Searchable preview text
	IsPinned      bool      `gorm:"default:false" json:"isPinned"`
	CreatedAt     time.Time `json:"createdAt"`
	LastAccessed  time.Time `json:"lastAccessed"`
	Hash          string    `gorm:"index" json:"-"` // For duplicate detection
}

// Settings represents application configuration
type Settings struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	GlobalHotkey       string    `gorm:"default:'Cmd+Shift+Space'" json:"globalHotkey"`
	PreviousItemHotkey string    `gorm:"default:'Cmd+Shift+C'" json:"previousItemHotkey"`
	PollingInterval    int       `gorm:"default:500" json:"pollingInterval"` // milliseconds
	MaxItems           int       `gorm:"default:100" json:"maxItems"`
	MaxDays            int       `gorm:"default:7" json:"maxDays"`
	AutoLaunch         bool      `gorm:"default:true" json:"autoLaunch"`
	EnableSounds       bool      `gorm:"default:false" json:"enableSounds"`
	MonitoringEnabled  bool      `gorm:"default:true" json:"monitoringEnabled"`
	AllowPasswords     bool      `gorm:"default:false" json:"allowPasswords"`  // Allow copying password-like content
	SortByRecent       string    `gorm:"default:'copied'" json:"sortByRecent"` // 'copied' or 'pasted' - secondary sort after pinned items
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (c *ClipboardItem) BeforeCreate(tx *gorm.DB) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.LastAccessed.IsZero() {
		c.LastAccessed = time.Now()
	}
	return nil
}

func (ClipboardItem) TableName() string {
	return "clipboard_items"
}

func (Settings) TableName() string {
	return "settings"
}
