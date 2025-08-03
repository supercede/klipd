package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestClipboardItem(t *testing.T) {
	// Test ClipboardItem struct creation
	item := ClipboardItem{
		ID:          "test-id",
		ContentType: "text",
		ContentText: "Hello, World!",
		PreviewText: "Hello, World!",
		IsPinned:    false,
		Hash:        "test-hash",
	}

	assert.Equal(t, "test-id", item.ID)
	assert.Equal(t, "text", item.ContentType)
	assert.Equal(t, "Hello, World!", item.ContentText)
	assert.Equal(t, "Hello, World!", item.PreviewText)
	assert.False(t, item.IsPinned)
	assert.Equal(t, "test-hash", item.Hash)
}

func TestClipboardItemTableName(t *testing.T) {
	item := ClipboardItem{}
	assert.Equal(t, "clipboard_items", item.TableName())
}

func TestSettingsTableName(t *testing.T) {
	settings := Settings{}
	assert.Equal(t, "settings", settings.TableName())
}

func TestClipboardItemBeforeCreate(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&ClipboardItem{})
	assert.NoError(t, err)

	// Test BeforeCreate hook with zero times
	item := &ClipboardItem{
		ID:          "test-id",
		ContentType: "text",
		ContentText: "test content",
		PreviewText: "test content",
	}

	// Save the item (should trigger BeforeCreate)
	err = db.Create(item).Error
	assert.NoError(t, err)

	// Verify timestamps were set
	assert.False(t, item.CreatedAt.IsZero())
	assert.False(t, item.LastAccessed.IsZero())
	assert.WithinDuration(t, time.Now(), item.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), item.LastAccessed, time.Second)
}

func TestClipboardItemBeforeCreateWithExistingTimes(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&ClipboardItem{})
	assert.NoError(t, err)

	// Test BeforeCreate hook with existing times
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	item := &ClipboardItem{
		ID:           "test-id-2",
		ContentType:  "text",
		ContentText:  "test content",
		PreviewText:  "test content",
		CreatedAt:    fixedTime,
		LastAccessed: fixedTime,
	}

	// Save the item
	err = db.Create(item).Error
	assert.NoError(t, err)

	// Verify timestamps were not changed
	assert.Equal(t, fixedTime, item.CreatedAt)
	assert.Equal(t, fixedTime, item.LastAccessed)
}

func TestSettings(t *testing.T) {
	settings := Settings{
		ID:                 1,
		GlobalHotkey:       "Cmd+Shift+V",
		PreviousItemHotkey: "Cmd+Shift+C",
		PollingInterval:    500,
		MaxItems:           100,
		MaxDays:            7,
		AutoLaunch:         true,
		EnableSounds:       false,
		MonitoringEnabled:  true,
	}

	assert.Equal(t, uint(1), settings.ID)
	assert.Equal(t, "Cmd+Shift+V", settings.GlobalHotkey)
	assert.Equal(t, "Cmd+Shift+C", settings.PreviousItemHotkey)
	assert.Equal(t, 500, settings.PollingInterval)
	assert.Equal(t, 100, settings.MaxItems)
	assert.Equal(t, 7, settings.MaxDays)
	assert.True(t, settings.AutoLaunch)
	assert.False(t, settings.EnableSounds)
	assert.True(t, settings.MonitoringEnabled)
}

func TestSettingsWithDefaults(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&Settings{})
	assert.NoError(t, err)

	// Create settings with minimal data (should use defaults)
	settings := &Settings{}
	err = db.Create(settings).Error
	assert.NoError(t, err)

	// Retrieve the settings to check defaults
	var retrieved Settings
	err = db.First(&retrieved, settings.ID).Error
	assert.NoError(t, err)

	// Note: GORM defaults are set at the database level, not in Go
	// So we need to check what actually gets stored
	assert.NotZero(t, retrieved.ID)
}

func TestClipboardItemBinaryContent(t *testing.T) {
	// Test with binary content
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	item := ClipboardItem{
		ID:            "binary-test",
		ContentType:   "image",
		ContentBinary: binaryData,
		PreviewText:   "PNG Image",
	}

	assert.Equal(t, "image", item.ContentType)
	assert.Equal(t, binaryData, item.ContentBinary)
	assert.Equal(t, "PNG Image", item.PreviewText)
	assert.Empty(t, item.ContentText) // Should be empty for binary content
}

func TestClipboardItemPinnedStatus(t *testing.T) {
	// Test pinned item
	pinnedItem := ClipboardItem{
		ID:          "pinned-test",
		ContentType: "text",
		ContentText: "Important text",
		IsPinned:    true,
	}

	assert.True(t, pinnedItem.IsPinned)

	// Test unpinned item
	unpinnedItem := ClipboardItem{
		ID:          "unpinned-test",
		ContentType: "text",
		ContentText: "Regular text",
		IsPinned:    false,
	}

	assert.False(t, unpinnedItem.IsPinned)
}
