package main

import (
	"context"
	"log"

	"klipd/config"
	"klipd/database"
	"klipd/models"
	"klipd/services"
)

// App struct
type App struct {
	ctx              context.Context
	db               *database.Database
	config           *config.Config
	clipboardMonitor *services.ClipboardMonitor
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database
	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	a.db = db

	// Initialize configuration
	a.config = config.NewConfig()

	// Load settings from database and update config
	if settings, err := a.db.GetSettings(); err == nil {
		settingsMap := map[string]interface{}{
			"pollingInterval":    settings.PollingInterval,
			"maxItems":           settings.MaxItems,
			"maxDays":            settings.MaxDays,
			"monitoringEnabled":  settings.MonitoringEnabled,
			"globalHotkey":       settings.GlobalHotkey,
			"previousItemHotkey": settings.PreviousItemHotkey,
			"autoLaunch":         settings.AutoLaunch,
			"enableSounds":       settings.EnableSounds,
		}
		a.config.UpdateFromSettings(settingsMap)
	}

	// Initialize clipboard monitor
	a.clipboardMonitor = services.NewClipboardMonitor(a.db, a.config)

	// Start clipboard monitoring
	if err := a.clipboardMonitor.Start(); err != nil {
		log.Printf("Failed to start clipboard monitor: %v", err)
	}

	log.Println("Klipd clipboard manager started successfully")
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	log.Println("Shutting down Klipd...")

	if a.clipboardMonitor != nil {
		a.clipboardMonitor.Stop()
	}

	if a.db != nil {
		a.db.Close()
	}
}

// GetClipboardItems returns clipboard items with optional pagination and filtering
func (a *App) GetClipboardItems(limit int, offset int, contentType string) ([]models.ClipboardItem, error) {
	return a.db.GetClipboardItems(limit, offset, contentType)
}

// SearchClipboardItems searches clipboard items by content
func (a *App) SearchClipboardItems(query string, limit int) ([]models.ClipboardItem, error) {
	if query == "" {
		return a.GetClipboardItems(limit, 0, "")
	}
	return a.clipboardMonitor.SearchItems(query, limit)
}

// GetClipboardItemByID retrieves a specific clipboard item
func (a *App) GetClipboardItemByID(id string) (*models.ClipboardItem, error) {
	return a.clipboardMonitor.GetItemByID(id)
}

// SelectClipboardItem copies a clipboard item back to the system clipboard
func (a *App) SelectClipboardItem(id string) error {
	return a.clipboardMonitor.CopyItemToClipboard(id)
}

// PinClipboardItem toggles the pin status of a clipboard item
func (a *App) PinClipboardItem(id string, pinned bool) error {
	return a.clipboardMonitor.PinItem(id, pinned)
}

// DeleteClipboardItem removes a clipboard item
func (a *App) DeleteClipboardItem(id string) error {
	return a.clipboardMonitor.DeleteItem(id)
}

// ClearAllClipboardItems removes all clipboard items
func (a *App) ClearAllClipboardItems(preservePinned bool) error {
	return a.clipboardMonitor.ClearAll(preservePinned)
}

// ClearClipboardItemsByType removes all clipboard items of a specific type
func (a *App) ClearClipboardItemsByType(contentType string, preservePinned bool) error {
	return a.clipboardMonitor.ClearByType(contentType, preservePinned)
}

// GetSettings returns the current application settings
func (a *App) GetSettings() (*models.Settings, error) {
	return a.db.GetSettings()
}

// UpdateSettings updates the application settings
func (a *App) UpdateSettings(settings *models.Settings) error {
	if err := a.db.UpdateSettings(settings); err != nil {
		return err
	}

	// Update runtime configuration
	settingsMap := map[string]interface{}{
		"pollingInterval":    settings.PollingInterval,
		"maxItems":           settings.MaxItems,
		"maxDays":            settings.MaxDays,
		"monitoringEnabled":  settings.MonitoringEnabled,
		"globalHotkey":       settings.GlobalHotkey,
		"previousItemHotkey": settings.PreviousItemHotkey,
		"autoLaunch":         settings.AutoLaunch,
		"enableSounds":       settings.EnableSounds,
	}
	a.config.UpdateFromSettings(settingsMap)

	// Update clipboard monitor configuration
	if a.clipboardMonitor != nil {
		a.clipboardMonitor.UpdateConfig(a.config)
	}

	return nil
}

// ToggleMonitoring pauses or resumes clipboard monitoring
func (a *App) ToggleMonitoring() bool {
	if a.config.MonitoringEnabled {
		a.config.MonitoringEnabled = false
		log.Println("Clipboard monitoring paused")
	} else {
		a.config.MonitoringEnabled = true
		log.Println("Clipboard monitoring resumed")
	}

	// Update the setting in database
	if settings, err := a.db.GetSettings(); err == nil {
		settings.MonitoringEnabled = a.config.MonitoringEnabled
		a.db.UpdateSettings(settings)
	}

	return a.config.MonitoringEnabled
}

// IsMonitoringEnabled returns the current monitoring status
func (a *App) IsMonitoringEnabled() bool {
	return a.config.MonitoringEnabled
}

// GetMonitoringStatus returns detailed monitoring status
func (a *App) GetMonitoringStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":         a.config.MonitoringEnabled,
		"pollingInterval": a.config.PollingInterval.Milliseconds(),
		"isRunning":       a.clipboardMonitor != nil && a.clipboardMonitor.IsRunning(),
	}
}
