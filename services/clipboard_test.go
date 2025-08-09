package services

import (
	"os"
	"testing"
	"time"

	"klipd/config"
	"klipd/database"
	"klipd/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestClipboardMonitor(t *testing.T) (*ClipboardMonitor, *database.Database) {
	// Create temporary directory for test database
	tempDir := t.TempDir()

	// Set environment variable to use test database path
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// Cleanup function
	t.Cleanup(func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Failed to restore HOME: %v", err)
		}
	})

	db, err := database.New()
	require.NoError(t, err)

	cfg := &config.Config{
		PollingInterval:   time.Millisecond * 100, // Fast polling for tests
		MonitoringEnabled: true,
		MaxItems:          10,
		MaxDays:           1,
	}

	monitor := NewClipboardMonitor(db, cfg)

	t.Cleanup(func() {
		monitor.Stop()
		if err := db.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	})

	return monitor, db
}

func TestNewClipboardMonitor(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.db)
	assert.NotNil(t, monitor.config)
	assert.NotNil(t, monitor.ctx)
	assert.NotNil(t, monitor.cancel)
	assert.False(t, monitor.isRunning)

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestClipboardMonitorStart(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Test starting the monitor
	err := monitor.Start()
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	// Test starting again (should return error)
	err = monitor.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestClipboardMonitorStop(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Start the monitor
	err := monitor.Start()
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	// Stop the monitor
	monitor.Stop()

	// Give it a moment to stop
	time.Sleep(time.Millisecond * 50)
	assert.False(t, monitor.IsRunning())

	// Test stopping again (should not panic)
	monitor.Stop()
	assert.False(t, monitor.IsRunning())

	// Cleanup
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestClipboardMonitorIsRunning(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Initially not running
	assert.False(t, monitor.IsRunning())

	// Start and check
	err := monitor.Start()
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	// Stop and check
	monitor.Stop()
	time.Sleep(time.Millisecond * 50)
	assert.False(t, monitor.IsRunning())

	// Cleanup
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestUpdateConfig(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Create new config
	newConfig := &config.Config{
		PollingInterval:   time.Second,
		MonitoringEnabled: false,
		MaxItems:          50,
		MaxDays:           30,
	}

	// Update config
	monitor.UpdateConfig(newConfig)

	// Verify config was updated
	assert.Equal(t, newConfig.PollingInterval, monitor.config.PollingInterval)
	assert.Equal(t, newConfig.MonitoringEnabled, monitor.config.MonitoringEnabled)
	assert.Equal(t, newConfig.MaxItems, monitor.config.MaxItems)
	assert.Equal(t, newConfig.MaxDays, monitor.config.MaxDays)

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestGenerateHash(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Test hash generation
	content1 := "Hello, World!"
	content2 := "Hello, World!"
	content3 := "Different content"

	hash1 := monitor.generateHash(content1)
	hash2 := monitor.generateHash(content2)
	hash3 := monitor.generateHash(content3)

	// Same content should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different content should produce different hash
	assert.NotEqual(t, hash1, hash3)

	// Hash should not be empty
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash3)

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestDetectContentType(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	tests := []struct {
		content      string
		expectedType string
	}{
		{"Hello, World!", "text"},
		{"file:///Users/test/document.pdf", "file"},
		{"http://example.com", "text"},
		{"https://example.com", "text"},
		{"C:\\Users\\test\\file.txt", "file"},
		{"/Users/test/image.png", "image"},
		{"", "text"},
		{"Multi\nline\ntext", "text"},
		{"/Users/test/image.jpg", "image"},
		{"~/Downloads/photo.jpeg", "image"},
	}

	for _, test := range tests {
		contentType := monitor.detectContentType(test.content)
		assert.Equal(t, test.expectedType, contentType, "Content: %s", test.content)
	}

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestConfigUtilities(t *testing.T) {
	cfg := config.NewConfig()

	// Test TruncatePreview function
	shortText := "Short text"
	longText := "This is a very long text that should be truncated because it exceeds the maximum preview length that we allow for clipboard items and goes on and on and on"

	shortPreview := config.TruncatePreview(shortText, 200)
	longPreview := config.TruncatePreview(longText, 50)

	assert.Equal(t, shortText, shortPreview)
	assert.LessOrEqual(t, len(longPreview), 53) // 50 + "..."
	assert.Contains(t, longPreview, "...")

	// Test IsImageFormat function
	assert.True(t, config.IsImageFormat("test.jpg"))
	assert.True(t, config.IsImageFormat("image.PNG"))
	assert.True(t, config.IsImageFormat("photo.gif"))
	assert.False(t, config.IsImageFormat("document.pdf"))
	assert.False(t, config.IsImageFormat("file.txt"))

	// Test ShouldSkipContent function with passwords disabled (default)
	assert.True(t, cfg.ShouldSkipContent(""))
	assert.True(t, cfg.ShouldSkipContent("   "))
	assert.False(t, cfg.ShouldSkipContent("Normal text"))
	assert.True(t, cfg.ShouldSkipContent("Password123!@#")) // Should skip passwords by default

	// Test ShouldSkipContent function with passwords enabled
	cfg.AllowPasswords = true
	assert.True(t, cfg.ShouldSkipContent(""))
	assert.True(t, cfg.ShouldSkipContent("   "))
	assert.False(t, cfg.ShouldSkipContent("Normal text"))
	assert.False(t, cfg.ShouldSkipContent("Password123!@#")) // Should NOT skip passwords when enabled
}

func TestGetRecentItems(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Create test items directly in database
	items := []models.ClipboardItem{
		{ID: "item-1", ContentType: "text", ContentText: "Content 1", PreviewText: "Content 1", Hash: "hash-1"},
		{ID: "item-2", ContentType: "text", ContentText: "Content 2", PreviewText: "Content 2", Hash: "hash-2"},
		{ID: "item-3", ContentType: "text", ContentText: "Content 3", PreviewText: "Content 3", Hash: "hash-3"},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(&item)
		assert.NoError(t, err)
	}

	// Test getting recent items
	recentItems, err := monitor.GetRecentItems(2)
	assert.NoError(t, err)
	assert.Len(t, recentItems, 2)

	// Test getting all items
	allItems, err := monitor.GetRecentItems(10)
	assert.NoError(t, err)
	assert.Len(t, allItems, 3)

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestSearchItems(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Create test items with searchable content
	items := []models.ClipboardItem{
		{ID: "search-1", ContentType: "text", ContentText: "Hello World", PreviewText: "Hello World", Hash: "search-hash-1"},
		{ID: "search-2", ContentType: "text", ContentText: "Go programming", PreviewText: "Go programming", Hash: "search-hash-2"},
		{ID: "search-3", ContentType: "text", ContentText: "JavaScript code", PreviewText: "JavaScript code", Hash: "search-hash-3"},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(&item)
		assert.NoError(t, err)
	}

	// Test search
	results, err := monitor.SearchItems("Hello", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-1", results[0].ID)

	// Test case-insensitive search
	results, err = monitor.SearchItems("go", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "search-2", results[0].ID)

	// Test no results
	results, err = monitor.SearchItems("nonexistent", 10)
	assert.NoError(t, err)
	assert.Len(t, results, 0)

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestPinItem(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Create test item
	item := &models.ClipboardItem{
		ID:          "pin-test",
		ContentType: "text",
		ContentText: "Pin test content",
		PreviewText: "Pin test content",
		Hash:        "pin-test-hash",
		IsPinned:    false,
	}

	err := db.CreateClipboardItem(item)
	assert.NoError(t, err)

	// Pin the item
	err = monitor.PinItem("pin-test", true)
	assert.NoError(t, err)

	// Verify it's pinned
	retrieved, err := monitor.GetItemByID("pin-test")
	assert.NoError(t, err)
	assert.True(t, retrieved.IsPinned)

	// Unpin the item
	err = monitor.PinItem("pin-test", false)
	assert.NoError(t, err)

	// Verify it's unpinned
	retrieved, err = monitor.GetItemByID("pin-test")
	assert.NoError(t, err)
	assert.False(t, retrieved.IsPinned)

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestRunCleanup(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Create some test items with different ages
	oldTime := time.Now().AddDate(0, 0, -2) // 2 days ago

	oldItem := &models.ClipboardItem{
		ID:          "old-item",
		ContentType: "text",
		ContentText: "Old content",
		PreviewText: "Old content",
		Hash:        "old-hash",
		CreatedAt:   oldTime,
	}

	err := db.DB.Create(oldItem).Error
	assert.NoError(t, err)

	// Update the timestamp manually
	err = db.DB.Model(oldItem).Update("created_at", oldTime).Error
	assert.NoError(t, err)

	// Create a recent item
	recentItem := &models.ClipboardItem{
		ID:          "recent-item",
		ContentType: "text",
		ContentText: "Recent content",
		PreviewText: "Recent content",
		Hash:        "recent-hash",
	}

	err = db.CreateClipboardItem(recentItem)
	assert.NoError(t, err)

	// Manually trigger cleanup using the database method directly
	// Since the monitor's cleanup runs on a timer, we test the cleanup functionality directly
	err = db.CleanupOldItems(10, 1) // Max 10 items, older than 1 day
	assert.NoError(t, err)

	// Verify cleanup happened (old item should be removed)
	items, err := db.GetClipboardItems(10, 0, "", "copied")
	assert.NoError(t, err)

	// Should have only the recent item
	foundRecent := false
	foundOld := false
	for _, item := range items {
		if item.ID == "recent-item" {
			foundRecent = true
		}
		if item.ID == "old-item" {
			foundOld = true
		}
	}

	assert.True(t, foundRecent, "Recent item should still exist")
	assert.False(t, foundOld, "Old item should be cleaned up")

	// Cleanup
	monitor.Stop()
	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}

func TestMonitorWithContext(t *testing.T) {
	monitor, db := setupTestClipboardMonitor(t)

	// Start the monitor
	err := monitor.Start()
	assert.NoError(t, err)
	assert.True(t, monitor.IsRunning())

	// The context should be active
	select {
	case <-monitor.ctx.Done():
		t.Fatal("Context should not be done while monitor is running")
	default:
		// Context is still active
	}

	// Stop the monitor
	monitor.Stop()

	// Wait a bit for the goroutines to finish
	time.Sleep(time.Millisecond * 100)

	// The context should be cancelled
	select {
	case <-monitor.ctx.Done():

	case <-time.After(time.Millisecond * 100):
		t.Fatal("Context should be done after stopping monitor")
	}

	if err := db.Close(); err != nil {
		t.Logf("Failed to close database: %v", err)
	}
}
