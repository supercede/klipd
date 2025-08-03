package database

import (
	"os"
	"testing"
	"time"

	"klipd/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *Database {
	tempDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)

	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	db, err := New()
	require.NoError(t, err)
	require.NotNil(t, db)

	return db
}

func TestNew(t *testing.T) {
	db := setupTestDB(t)
	assert.NotNil(t, db)
	assert.NotNil(t, db.DB)

	// Test that tables were created
	var tableCount int64
	err := db.DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('clipboard_items', 'settings')").Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(2), tableCount)
}

func TestCreateClipboardItem(t *testing.T) {
	db := setupTestDB(t)

	item := &models.ClipboardItem{
		ID:          "test-id-1",
		ContentType: "text",
		ContentText: "Hello, World!",
		PreviewText: "Hello, World!",
		Hash:        "test-hash-1",
	}

	err := db.CreateClipboardItem(item)
	assert.NoError(t, err)

	retrieved, err := db.GetClipboardItemByID("test-id-1")
	assert.NoError(t, err)
	assert.Equal(t, item.ID, retrieved.ID)
	assert.Equal(t, item.ContentText, retrieved.ContentText)
	assert.Equal(t, item.ContentType, retrieved.ContentType)
}

func TestCreateClipboardItemDuplicate(t *testing.T) {
	db := setupTestDB(t)

	item := &models.ClipboardItem{
		ID:          "test-id-dup",
		ContentType: "text",
		ContentText: "Duplicate test",
		PreviewText: "Duplicate test",
		Hash:        "duplicate-hash",
	}

	err := db.CreateClipboardItem(item)
	assert.NoError(t, err)

	item2 := &models.ClipboardItem{
		ID:          "test-id-dup-2",
		ContentType: "text",
		ContentText: "Duplicate test",
		PreviewText: "Duplicate test",
		Hash:        "duplicate-hash",
	}

	err = db.CreateClipboardItem(item2)
	assert.NoError(t, err)

	items, err := db.GetClipboardItems(10, 0, "")
	assert.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetClipboardItemByID(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.GetClipboardItemByID("non-existent")
	assert.Error(t, err)

	item := &models.ClipboardItem{
		ID:          "get-test-id",
		ContentType: "text",
		ContentText: "Get test content",
		PreviewText: "Get test content",
		Hash:        "get-test-hash",
	}

	err = db.CreateClipboardItem(item)
	assert.NoError(t, err)

	retrieved, err := db.GetClipboardItemByID("get-test-id")
	assert.NoError(t, err)
	assert.Equal(t, item.ID, retrieved.ID)
	assert.Equal(t, item.ContentText, retrieved.ContentText)
}

func TestGetClipboardItems(t *testing.T) {
	db := setupTestDB(t)

	items := []models.ClipboardItem{
		{ID: "item-1", ContentType: "text", ContentText: "Content 1", PreviewText: "Content 1", Hash: "hash-1"},
		{ID: "item-2", ContentType: "text", ContentText: "Content 2", PreviewText: "Content 2", Hash: "hash-2"},
		{ID: "item-3", ContentType: "text", ContentText: "Content 3", PreviewText: "Content 3", Hash: "hash-3"},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(&item)
		assert.NoError(t, err)
	}

	// Test pagination
	retrieved, err := db.GetClipboardItems(2, 0, "")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Test with offset
	retrieved, err = db.GetClipboardItems(2, 1, "")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Test content type filter
	retrieved, err = db.GetClipboardItems(10, 0, "text")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Test non-matching content type filter
	retrieved, err = db.GetClipboardItems(10, 0, "image")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 0)
}

func TestSearchClipboardItems(t *testing.T) {
	db := setupTestDB(t)

	items := []models.ClipboardItem{
		{ID: "search-1", ContentType: "text", ContentText: "Hello World", PreviewText: "Hello World", Hash: "search-hash-1"},
		{ID: "search-2", ContentType: "text", ContentText: "Go programming", PreviewText: "Go programming", Hash: "search-hash-2"},
		{ID: "search-3", ContentType: "text", ContentText: "JavaScript code", PreviewText: "JavaScript code", Hash: "search-hash-3"},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(&item)
		assert.NoError(t, err)
	}

	results, err := db.SearchClipboardItems("Hello", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-1", results[0].ID)

	results, err = db.SearchClipboardItems("hello", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = db.SearchClipboardItems("program", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-2", results[0].ID)

	results, err = db.SearchClipboardItems("nonexistent", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestUpdateClipboardItemPin(t *testing.T) {
	db := setupTestDB(t)

	item := &models.ClipboardItem{
		ID:          "pin-test-id",
		ContentType: "text",
		ContentText: "Pin test content",
		PreviewText: "Pin test content",
		Hash:        "pin-test-hash",
		IsPinned:    false,
	}

	err := db.CreateClipboardItem(item)
	assert.NoError(t, err)

	// Pin the item
	err = db.PinClipboardItem("pin-test-id", true)
	assert.NoError(t, err)

	// Verify it's pinned
	retrieved, err := db.GetClipboardItemByID("pin-test-id")
	assert.NoError(t, err)
	assert.True(t, retrieved.IsPinned)

	// Unpin the item
	err = db.PinClipboardItem("pin-test-id", false)
	assert.NoError(t, err)

	// Verify it's unpinned
	retrieved, err = db.GetClipboardItemByID("pin-test-id")
	assert.NoError(t, err)
	assert.False(t, retrieved.IsPinned)
}

func TestDeleteClipboardItem(t *testing.T) {
	db := setupTestDB(t)

	item := &models.ClipboardItem{
		ID:          "delete-test-id",
		ContentType: "text",
		ContentText: "Delete test content",
		PreviewText: "Delete test content",
		Hash:        "delete-test-hash",
	}

	err := db.CreateClipboardItem(item)
	assert.NoError(t, err)

	_, err = db.GetClipboardItemByID("delete-test-id")
	assert.NoError(t, err)

	err = db.DeleteClipboardItem("delete-test-id")
	assert.NoError(t, err)

	_, err = db.GetClipboardItemByID("delete-test-id")
	assert.Error(t, err)
}

func TestClearAllClipboardItems(t *testing.T) {
	db := setupTestDB(t)

	// Create multiple items, some pinned
	items := []models.ClipboardItem{
		{ID: "clear-1", ContentType: "text", ContentText: "Content 1", PreviewText: "Content 1", Hash: "clear-hash-1", IsPinned: false},
		{ID: "clear-2", ContentType: "text", ContentText: "Content 2", PreviewText: "Content 2", Hash: "clear-hash-2", IsPinned: true},
		{ID: "clear-3", ContentType: "text", ContentText: "Content 3", PreviewText: "Content 3", Hash: "clear-hash-3", IsPinned: false},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(&item)
		assert.NoError(t, err)
	}

	// Clear all items preserving pinned
	err := db.ClearAllItems(true)
	assert.NoError(t, err)

	// Verify only pinned item remains
	allItems, err := db.GetClipboardItems(10, 0, "")
	assert.NoError(t, err)
	assert.Len(t, allItems, 1)
	assert.Equal(t, "clear-2", allItems[0].ID)
	assert.True(t, allItems[0].IsPinned)

	// Clear all items including pinned
	err = db.DB.Unscoped().Where("1 = 1").Delete(&models.ClipboardItem{}).Error
	assert.NoError(t, err)

	// Verify no items remain
	allItems, err = db.GetClipboardItems(10, 0, "")
	assert.NoError(t, err)
	assert.Len(t, allItems, 0)
}

func TestClearClipboardItemsByType(t *testing.T) {
	db := setupTestDB(t)

	// Create items of different types
	items := []models.ClipboardItem{
		{ID: "type-1", ContentType: "text", ContentText: "Text content", PreviewText: "Text content", Hash: "type-hash-1"},
		{ID: "type-2", ContentType: "image", PreviewText: "Image content", Hash: "type-hash-2"},
		{ID: "type-3", ContentType: "text", ContentText: "More text", PreviewText: "More text", Hash: "type-hash-3", IsPinned: true},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(&item)
		assert.NoError(t, err)
	}

	// Clear only text items, preserving pinned
	err := db.ClearItemsByType("text", true)
	assert.NoError(t, err)

	// Verify results
	allItems, err := db.GetClipboardItems(10, 0, "")
	assert.NoError(t, err)
	assert.Len(t, allItems, 2) // Should have image item and pinned text item

	// Clear all text items including pinned
	err = db.ClearItemsByType("text", false)
	assert.NoError(t, err)

	// Verify only image item remains
	allItems, err = db.GetClipboardItems(10, 0, "")
	assert.NoError(t, err)
	assert.Len(t, allItems, 1)
	assert.Equal(t, "image", allItems[0].ContentType)
}

func TestCleanupOldItems(t *testing.T) {
	db := setupTestDB(t)

	// Create items with different ages
	oldTime := time.Now().AddDate(0, 0, -10)   // 10 days ago
	recentTime := time.Now().AddDate(0, 0, -1) // 1 day ago

	// Create old items
	oldItem := &models.ClipboardItem{
		ID:          "old-item",
		ContentType: "text",
		ContentText: "Old content",
		PreviewText: "Old content",
		Hash:        "old-hash",
		CreatedAt:   oldTime,
	}

	// We need to insert directly to bypass the BeforeCreate hook
	err := db.DB.Create(oldItem).Error
	assert.NoError(t, err)

	// Update the created_at timestamp manually
	err = db.DB.Model(oldItem).Update("created_at", oldTime).Error
	assert.NoError(t, err)

	// Create recent item
	recentItem := &models.ClipboardItem{
		ID:          "recent-item",
		ContentType: "text",
		ContentText: "Recent content",
		PreviewText: "Recent content",
		Hash:        "recent-hash",
		CreatedAt:   recentTime,
	}

	err = db.DB.Create(recentItem).Error
	assert.NoError(t, err)

	err = db.DB.Model(recentItem).Update("created_at", recentTime).Error
	assert.NoError(t, err)

	// Create old pinned item
	oldPinnedItem := &models.ClipboardItem{
		ID:          "old-pinned",
		ContentType: "text",
		ContentText: "Old pinned",
		PreviewText: "Old pinned",
		Hash:        "old-pinned-hash",
		IsPinned:    true,
		CreatedAt:   oldTime,
	}

	err = db.DB.Create(oldPinnedItem).Error
	assert.NoError(t, err)

	err = db.DB.Model(oldPinnedItem).Update("created_at", oldTime).Error
	assert.NoError(t, err)

	// Cleanup items older than 7 days
	err = db.CleanupOldItems(100, 7) // Use 100 max items, 7 max days
	assert.NoError(t, err)

	// Verify results - old unpinned items should be removed
	allItems, err := db.GetClipboardItems(10, 0, "")
	assert.NoError(t, err)

	// Should have recent item and old pinned item (old unpinned item should be removed)
	foundRecent := false
	foundOldPinned := false
	foundOld := false

	for _, item := range allItems {
		switch item.ID {
		case "recent-item":
			foundRecent = true
		case "old-pinned":
			foundOldPinned = true
		case "old-item":
			foundOld = true
		}
	}

	assert.True(t, foundRecent, "Recent item should still exist")
	assert.True(t, foundOldPinned, "Old pinned item should still exist")
	assert.False(t, foundOld, "Old unpinned item should be cleaned up")
}

func TestSettings(t *testing.T) {
	db := setupTestDB(t)

	// First get the default settings that were created during initialization
	defaultSettings, err := db.GetSettings()
	assert.NoError(t, err)

	// Update the existing settings
	defaultSettings.GlobalHotkey = "Cmd+V"
	defaultSettings.PreviousItemHotkey = "Cmd+Shift+V"
	defaultSettings.PollingInterval = 1000
	defaultSettings.MaxItems = 50
	defaultSettings.MaxDays = 14
	defaultSettings.AutoLaunch = false
	defaultSettings.EnableSounds = true
	defaultSettings.MonitoringEnabled = false
	defaultSettings.AllowPasswords = true

	err = db.UpdateSettings(defaultSettings)
	assert.NoError(t, err)

	// Test retrieving updated settings
	retrieved, err := db.GetSettings()
	assert.NoError(t, err)
	assert.Equal(t, "Cmd+V", retrieved.GlobalHotkey)
	assert.Equal(t, 1000, retrieved.PollingInterval)
	assert.False(t, retrieved.AutoLaunch)
	assert.True(t, retrieved.AllowPasswords)

	// Test updating settings again
	defaultSettings.MaxItems = 200
	defaultSettings.EnableSounds = false
	defaultSettings.AllowPasswords = false
	err = db.UpdateSettings(defaultSettings)
	assert.NoError(t, err)

	// Verify update
	retrieved, err = db.GetSettings()
	assert.NoError(t, err)
	assert.Equal(t, 200, retrieved.MaxItems)
	assert.False(t, retrieved.EnableSounds)
	assert.False(t, retrieved.AllowPasswords)
}

func TestClose(t *testing.T) {
	db := setupTestDB(t)

	err := db.Close()
	assert.NoError(t, err)

	_, err = db.GetClipboardItems(10, 0, "")
	assert.Error(t, err)
}
