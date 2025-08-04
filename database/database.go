package database

import (
	"os"
	"path/filepath"
	"time"

	"klipd/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func New() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Create app data directory
	appDir := filepath.Join(homeDir, "Library", "Application Support", "Klipd")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(appDir, "clipboard.db")

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent in production
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}

	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, err
	}

	// Configure SQLite for better performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Execute SQLite pragmas for performance
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA temp_store=MEMORY")
	db.Exec("PRAGMA mmap_size=268435456")
	db.Exec("PRAGMA optimize")

	database := &Database{DB: db}

	if err := database.migrate(); err != nil {
		return nil, err
	}

	if err := database.initializeSettings(); err != nil {
		return nil, err
	}

	return database, nil
}

func (d *Database) migrate() error {
	return d.DB.AutoMigrate(
		&models.ClipboardItem{},
		&models.Settings{},
	)
}

func (d *Database) initializeSettings() error {
	var count int64
	if err := d.DB.Model(&models.Settings{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		defaultSettings := &models.Settings{
			GlobalHotkey:       "Cmd+Shift+V",
			PreviousItemHotkey: "Cmd+Shift+C",
			PollingInterval:    500,
			MaxItems:           100,
			MaxDays:            7,
			AutoLaunch:         true,
			EnableSounds:       false,
			MonitoringEnabled:  true,
			AllowPasswords:     false,
		}
		return d.DB.Create(defaultSettings).Error
	}

	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	err := d.DB.First(&settings).Error
	return &settings, err
}

func (d *Database) UpdateSettings(settings *models.Settings) error {
	return d.DB.Save(settings).Error
}

func (d *Database) CreateClipboardItem(item *models.ClipboardItem) error {
	return d.DB.Create(item).Error
}

func (d *Database) GetClipboardItems(limit int, offset int, contentType string, sortByRecent string) ([]models.ClipboardItem, error) {
	var items []models.ClipboardItem
	query := d.DB.Model(&models.ClipboardItem{})

	if contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}

	var orderClause string
	if sortByRecent == "copied" {
		orderClause = "is_pinned DESC, created_at DESC"
	} else {
		orderClause = "is_pinned DESC, last_accessed DESC"
	}

	err := query.Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	return items, err
}

func (d *Database) SearchClipboardItems(searchTerm string, limit int, offset int, sortByRecent string) ([]models.ClipboardItem, error) {
	var items []models.ClipboardItem

	var orderClause string
	if sortByRecent == "copied" {
		orderClause = "is_pinned DESC, created_at DESC"
	} else {
		orderClause = "is_pinned DESC, last_accessed DESC"
	}

	err := d.DB.Where("preview_text LIKE ?", "%"+searchTerm+"%").
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	return items, err
}

func (d *Database) SearchClipboardItemsRegex(regexPattern string, limit int, offset int, sortByRecent string) ([]models.ClipboardItem, error) {
	var items []models.ClipboardItem
	var orderClause string
	if sortByRecent == "copied" {
		orderClause = "is_pinned DESC, created_at DESC"
	} else {
		orderClause = "is_pinned DESC, last_accessed DESC"
	}

	// SQLite REGEXP operator (if available)
	err := d.DB.Where("preview_text REGEXP ?", regexPattern).
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&items).Error
	return items, err
}

func (d *Database) GetClipboardItemByID(id string) (*models.ClipboardItem, error) {
	var item models.ClipboardItem
	err := d.DB.Where("id = ?", id).First(&item).Error
	return &item, err
}

func (d *Database) UpdateClipboardItem(item *models.ClipboardItem) error {
	return d.DB.Save(item).Error
}

func (d *Database) DeleteClipboardItem(id string) error {
	return d.DB.Where("id = ?", id).Delete(&models.ClipboardItem{}).Error
}

func (d *Database) PinClipboardItem(id string, pinned bool) error {
	return d.DB.Model(&models.ClipboardItem{}).
		Where("id = ?", id).
		Update("is_pinned", pinned).Error
}

func (d *Database) CleanupOldItems(maxItems int, maxDays int) error {
	// Delete items older than maxDays (excluding pinned items)
	cutoffDate := time.Now().AddDate(0, 0, -maxDays)
	if err := d.DB.Where("created_at < ? AND is_pinned = false", cutoffDate).
		Delete(&models.ClipboardItem{}).Error; err != nil {
		return err
	}

	// Count total items (excluding pinned)
	var count int64
	if err := d.DB.Model(&models.ClipboardItem{}).
		Where("is_pinned = false").
		Count(&count).Error; err != nil {
		return err
	}

	// If we have more than maxItems, delete the oldest ones
	if int(count) > maxItems {
		itemsToDelete := int(count) - maxItems
		var oldestItems []models.ClipboardItem

		if err := d.DB.Where("is_pinned = false").
			Order("created_at ASC").
			Limit(itemsToDelete).
			Find(&oldestItems).Error; err != nil {
			return err
		}

		for _, item := range oldestItems {
			if err := d.DB.Delete(&item).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Database) GetItemByHash(hash string) (*models.ClipboardItem, error) {
	var item models.ClipboardItem
	err := d.DB.Where("hash = ?", hash).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *Database) ClearAllItems(preservePinned bool) error {
	query := d.DB
	if preservePinned {
		query = query.Where("is_pinned = false")
	}
	return query.Delete(&models.ClipboardItem{}).Error
}

func (d *Database) ClearItemsByType(contentType string, preservePinned bool) error {
	query := d.DB.Where("content_type = ?", contentType)
	if preservePinned {
		query = query.Where("is_pinned = false")
	}
	return query.Delete(&models.ClipboardItem{}).Error
}
