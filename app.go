package main

import (
	"context"
	"log"

	"klipd/config"
	"klipd/database"
	"klipd/models"
	"klipd/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	db               *database.Database
	config           *config.Config
	clipboardMonitor *services.ClipboardMonitor
	hotkeyManager    *services.HotkeyManager
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

	// Set Wails context for event emission
	a.clipboardMonitor.SetWailsContext(a.ctx)

	// Initialize hotkey manager
	a.hotkeyManager = services.NewHotkeyManager()

	// Register and start global hotkeys
	if err := a.setupHotkeys(); err != nil {
		log.Printf("Failed to setup hotkeys: %v", err)
	}
	if err := a.hotkeyManager.Start(); err != nil {
		log.Printf("Failed to start hotkey manager: %v", err)
	}

	log.Println("Klipd clipboard manager started successfully")
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	log.Println("Shutting down Klipd...")

	if a.clipboardMonitor != nil {
		a.clipboardMonitor.Stop()
	}

	if a.hotkeyManager != nil {
		a.hotkeyManager.Stop()
	}

	if a.db != nil {
		a.db.Close()
	}
}

// setupHotkeys configures global hotkeys
func (a *App) setupHotkeys() error {
	// Register global hotkey for showing clipboard history (focus app + show search)
	settings, err := a.db.GetSettings()
	if err != nil {
		return err
	}
	hotkeyStr := settings.GlobalHotkey
	if hotkeyStr == "" {
		hotkeyStr = "Cmd+Shift+Space"
	}
	err = a.hotkeyManager.Register(hotkeyStr, func() {
		log.Printf("Global hotkey triggered: %s", hotkeyStr)
		// Bring window to front and show search interface
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "show-search-interface")
	})
	if err != nil {
		return err
	}

	// Register previous item hotkey
	previousHotkey := a.config.PreviousHotkey
	if previousHotkey == "" {
		previousHotkey = "Cmd+Shift+C" // Default hotkey
	}

	err = a.hotkeyManager.Register(previousHotkey, func() {
		log.Printf("Previous item hotkey triggered: %s", previousHotkey)
		// Get the most recent clipboard item and paste it
		a.pasteLastItem()
	})
	if err != nil {
		return err
	}

	// Register show window hotkey
	showWindowHotkey := "Cmd+Shift+K" // Show main window hotkey
	err = a.hotkeyManager.Register(showWindowHotkey, func() {
		log.Printf("Show window hotkey triggered: %s", showWindowHotkey)
		a.ShowMainWindow()
	})
	if err != nil {
		return err
	}

	return nil
}

// pasteLastItem copies the most recent clipboard item to system clipboard
func (a *App) pasteLastItem() {
	items, err := a.db.GetClipboardItems(1, 0, "", "copied")
	if err != nil {
		log.Printf("Failed to get recent items: %v", err)
		return
	}

	if len(items) > 0 {
		err := a.clipboardMonitor.CopyItemToClipboard(items[0].ID)
		if err != nil {
			log.Printf("Failed to copy item to clipboard: %v", err)
		} else {
			log.Printf("Pasted last clipboard item: %s", items[0].PreviewText)
		}
	}
}

// ShowSearchInterface emits an event to show the search interface (callable from frontend)
func (a *App) ShowSearchInterface() {
	runtime.EventsEmit(a.ctx, "show-search-interface")
}

// HideSearchInterface emits an event to hide the search interface (callable from frontend)
func (a *App) HideSearchInterface() {
	runtime.EventsEmit(a.ctx, "hide-search-interface")
}

// TriggerGlobalHotkey manually triggers the global hotkey (for testing)
// This function is now a placeholder as the new library doesn't support manual triggering.
func (a *App) TriggerGlobalHotkey() {
	log.Println("Manual hotkey triggering is not supported by the new library.")
}

// GetClipboardItems returns clipboard items with optional pagination and filtering
func (a *App) GetClipboardItems(limit int, offset int, contentType string) ([]models.ClipboardItem, error) {
	settings, err := a.db.GetSettings()
	if err != nil {
		return a.db.GetClipboardItems(limit, offset, contentType, "copied")
	}
	return a.db.GetClipboardItems(limit, offset, contentType, settings.SortByRecent)
}

func (a *App) GetClipboardItemsPaginated(limit int, offset int, contentType string) ([]models.ClipboardItem, error) {
	return a.GetClipboardItems(limit, offset, contentType)
}

// SearchClipboardItems searches clipboard items by content
func (a *App) SearchClipboardItems(query string, limit int) ([]models.ClipboardItem, error) {
	if query == "" {
		return a.GetClipboardItems(limit, 0, "")
	}
	return a.SearchClipboardItemsPaginated(query, limit, 0, false)
}

func (a *App) SearchClipboardItemsPaginated(query string, limit int, offset int, useRegex bool) ([]models.ClipboardItem, error) {
	settings, err := a.db.GetSettings()
	sortByRecent := "copied"
	if err == nil {
		sortByRecent = settings.SortByRecent
	}

	if query == "" {
		return a.db.GetClipboardItems(limit, offset, "", sortByRecent)
	}

	if useRegex {
		return a.db.SearchClipboardItemsRegex(query, limit, offset, sortByRecent)
	}
	return a.db.SearchClipboardItems(query, limit, offset, sortByRecent)
}

// SearchClipboardItemsRegex searches clipboard items using regex patterns
func (a *App) SearchClipboardItemsRegex(regexPattern string, limit int) ([]models.ClipboardItem, error) {
	settings, err := a.db.GetSettings()
	sortByRecent := "copied"
	if err == nil {
		sortByRecent = settings.SortByRecent
	}
	return a.db.SearchClipboardItemsRegex(regexPattern, limit, 0, sortByRecent)
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
	// Re-register hotkeys on settings update
	a.hotkeyManager.Stop()
	a.hotkeyManager = services.NewHotkeyManager()
	if err := a.setupHotkeys(); err != nil {
		log.Printf("Failed to re-setup hotkeys after settings update: %v", err)
	}
	if err := a.hotkeyManager.Start(); err != nil {
		log.Printf("Failed to re-start hotkey manager: %v", err)
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

// ShowMainWindow shows the main application window
func (a *App) ShowMainWindow() {
	runtime.WindowShow(a.ctx)
}

// ShowPreferences shows the preferences window
func (a *App) ShowPreferences() {
	runtime.WindowShow(a.ctx)
	// The frontend will handle showing the preferences modal
	runtime.EventsEmit(a.ctx, "show-preferences")
}

// Quit gracefully shuts down the application
func (a *App) Quit() {
	runtime.Quit(a.ctx)
}

// GetRecentItems returns the most recent clipboard items for the menu bar
func (a *App) GetRecentItems(limit int) ([]models.ClipboardItem, error) {
	if limit <= 0 {
		limit = 5 // Default to 5 items
	}
	return a.db.GetClipboardItems(limit, 0, "", "recent") // Get recent items, all types
}
