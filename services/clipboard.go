package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
	"time"

	"klipd/config"
	"klipd/database"
	"klipd/models"

	"github.com/atotto/clipboard"
	"github.com/google/uuid"
)

// handles clipboard monitoring and management
type ClipboardMonitor struct {
	db            *database.Database
	config        *config.Config
	lastHash      string
	isRunning     bool
	ctx           context.Context
	cancel        context.CancelFunc
	cleanupTicker *time.Ticker
}

func NewClipboardMonitor(db *database.Database, cfg *config.Config) *ClipboardMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ClipboardMonitor{
		db:     db,
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (cm *ClipboardMonitor) Start() error {
	if cm.isRunning {
		return fmt.Errorf("clipboard monitor is already running")
	}

	cm.isRunning = true
	log.Println("Starting clipboard monitor...")

	// Get initial clipboard content to establish baseline
	if initialContent, err := clipboard.ReadAll(); err == nil {
		cm.lastHash = cm.generateHash(initialContent)
	}

	// Start monitoring goroutine
	go cm.monitorClipboard()

	// Start cleanup goroutine
	go cm.runCleanup()

	return nil
}

// Stop stops clipboard monitoring
func (cm *ClipboardMonitor) Stop() {
	if !cm.isRunning {
		return
	}

	log.Println("Stopping clipboard monitor...")
	cm.isRunning = false
	cm.cancel()

	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}
}

func (cm *ClipboardMonitor) IsRunning() bool {
	return cm.isRunning
}

func (cm *ClipboardMonitor) UpdateConfig(cfg *config.Config) {
	cm.config = cfg
}

// monitorClipboard is the main monitoring loop
func (cm *ClipboardMonitor) monitorClipboard() {
	ticker := time.NewTicker(cm.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			if cm.config.MonitoringEnabled {
				cm.checkClipboard()
			}
		}
	}
}

// checkClipboard checks for clipboard changes and processes new content
func (cm *ClipboardMonitor) checkClipboard() {
	content, err := clipboard.ReadAll()

	if err != nil {
		return
	}

	log.Printf("Clipboard content read: %s", config.TruncatePreview(content, 50))

	// Skip if content hasn't changed
	currentHash := cm.generateHash(content)
	if currentHash == cm.lastHash {
		return
	}

	cm.lastHash = currentHash

	// Skip if content should be ignored
	if cm.config.ShouldSkipContent(content) {
		return
	}

	// Check for duplicate content
	if existingItem, err := cm.db.GetItemByHash(currentHash); err == nil {
		// Update last accessed time for existing item
		existingItem.LastAccessed = time.Now()
		if err := cm.db.UpdateClipboardItem(existingItem); err != nil {
			log.Printf("Error updating existing clipboard item: %v", err)
		}
		return
	}

	// Create new clipboard item
	item := &models.ClipboardItem{
		ID:           uuid.New().String(),
		ContentType:  cm.detectContentType(content),
		ContentText:  content,
		PreviewText:  config.TruncatePreview(content, 200),
		Hash:         currentHash,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		IsPinned:     false,
	}

	// Handle binary content if needed
	if item.ContentType == "image" {
		// For now, we'll store image content as text (file paths, URLs, etc.)
		// In the future, this could/will be enhanced to handle actual binary data
		item.ContentBinary = nil
	}

	// Save to database
	if err := cm.db.CreateClipboardItem(item); err != nil {
		log.Printf("Error saving clipboard item: %v", err)
		return
	}

	log.Printf("New clipboard item saved: %s (type: %s)",
		config.TruncatePreview(content, 50), item.ContentType)
}

func (cm *ClipboardMonitor) detectContentType(content string) string {
	content = strings.TrimSpace(content)

	if cm.looksLikeFilePath(content) {
		if config.IsImageFormat(content) {
			return "image"
		}
		return "file"
	}

	// Check if it's a URL to an image
	if cm.looksLikeURL(content) && config.IsImageFormat(content) {
		return "image"
	}

	// Default to text
	return "text"
}

func (cm *ClipboardMonitor) looksLikeFilePath(content string) bool {
	// Simple heuristics for file paths
	return strings.HasPrefix(content, "/") || // Unix absolute path
		strings.HasPrefix(content, "~/") || // Unix home path
		strings.Contains(content, ":\\") || // Windows drive path
		strings.HasPrefix(content, "file://") // File URL
}

func (cm *ClipboardMonitor) looksLikeURL(content string) bool {
	return strings.HasPrefix(content, "http://") ||
		strings.HasPrefix(content, "https://") ||
		strings.HasPrefix(content, "ftp://")
}

func (cm *ClipboardMonitor) generateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

func (cm *ClipboardMonitor) runCleanup() {
	cm.cleanupTicker = time.NewTicker(cm.config.CleanupInterval())
	defer cm.cleanupTicker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.cleanupTicker.C:
			cm.performCleanup()
		}
	}
}

// performCleanup removes old clipboard items based on configuration
func (cm *ClipboardMonitor) performCleanup() {
	log.Println("Running clipboard cleanup...")

	settings, err := cm.db.GetSettings()
	if err != nil {
		log.Printf("Error getting settings for cleanup: %v", err)
		return
	}

	if err := cm.db.CleanupOldItems(settings.MaxItems, settings.MaxDays); err != nil {
		log.Printf("Error during cleanup: %v", err)
	} else {
		log.Println("Clipboard cleanup completed")
	}
}

func (cm *ClipboardMonitor) GetRecentItems(limit int) ([]models.ClipboardItem, error) {
	return cm.db.GetClipboardItems(limit, 0, "")
}

func (cm *ClipboardMonitor) SearchItems(query string, limit int) ([]models.ClipboardItem, error) {
	return cm.db.SearchClipboardItems(query, limit)
}

func (cm *ClipboardMonitor) PinItem(id string, pinned bool) error {
	return cm.db.PinClipboardItem(id, pinned)
}

func (cm *ClipboardMonitor) DeleteItem(id string) error {
	return cm.db.DeleteClipboardItem(id)
}

func (cm *ClipboardMonitor) GetItemByID(id string) (*models.ClipboardItem, error) {
	return cm.db.GetClipboardItemByID(id)
}

func (cm *ClipboardMonitor) CopyItemToClipboard(id string) error {
	item, err := cm.db.GetClipboardItemByID(id)
	if err != nil {
		return err
	}

	item.LastAccessed = time.Now()
	if err := cm.db.UpdateClipboardItem(item); err != nil {
		log.Printf("Error updating last accessed time: %v", err)
	}

	// Copy to clipboard
	return clipboard.WriteAll(item.ContentText)
}

func (cm *ClipboardMonitor) ClearAll(preservePinned bool) error {
	return cm.db.ClearAllItems(preservePinned)
}

func (cm *ClipboardMonitor) ClearByType(contentType string, preservePinned bool) error {
	return cm.db.ClearItemsByType(contentType, preservePinned)
}
