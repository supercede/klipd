package config

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// Config holds runtime configuration for the clipboard manager
type Config struct {
	PollingInterval   time.Duration
	MaxItems          int
	MaxDays           int
	MonitoringEnabled bool
	GlobalHotkey      string
	PreviousHotkey    string
	AutoLaunch        bool
	EnableSounds      bool
	AllowPasswords    bool
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		PollingInterval:   500 * time.Millisecond,
		MaxItems:          100,
		MaxDays:           7,
		MonitoringEnabled: true,
		GlobalHotkey:      "Cmd+Shift+V",
		PreviousHotkey:    "Cmd+Shift+C",
		AutoLaunch:        true,
		EnableSounds:      false,
		AllowPasswords:    false,
	}
}

// UpdateFromSettings updates config from database settings
func (c *Config) UpdateFromSettings(settings map[string]interface{}) {
	if val, ok := settings["pollingInterval"].(int); ok {
		c.PollingInterval = time.Duration(val) * time.Millisecond
	}
	if val, ok := settings["maxItems"].(int); ok {
		c.MaxItems = val
	}
	if val, ok := settings["maxDays"].(int); ok {
		c.MaxDays = val
	}
	if val, ok := settings["monitoringEnabled"].(bool); ok {
		c.MonitoringEnabled = val
	}
	if val, ok := settings["globalHotkey"].(string); ok {
		c.GlobalHotkey = val
	}
	if val, ok := settings["previousItemHotkey"].(string); ok {
		c.PreviousHotkey = val
	}
	if val, ok := settings["autoLaunch"].(bool); ok {
		c.AutoLaunch = val
	}
	if val, ok := settings["enableSounds"].(bool); ok {
		c.EnableSounds = val
	}
	if val, ok := settings["allowPasswords"].(bool); ok {
		c.AllowPasswords = val
	}
}

// ContentType represents the type of clipboard content
type ContentType int

const (
	ContentTypeText ContentType = iota
	ContentTypeImage
	ContentTypeFile
)

func (ct ContentType) String() string {
	switch ct {
	case ContentTypeText:
		return "text"
	case ContentTypeImage:
		return "image"
	case ContentTypeFile:
		return "file"
	default:
		return "unknown"
	}
}

func ParseContentType(s string) ContentType {
	switch strings.ToLower(s) {
	case "text":
		return ContentTypeText
	case "image":
		return ContentTypeImage
	case "file":
		return ContentTypeFile
	default:
		return ContentTypeText
	}
}

func GenerateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// TruncatePreview creates a preview text with specified length
func TruncatePreview(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	// Find a good break point (space, newline, etc.)
	truncated := text[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	lastNewline := strings.LastIndex(truncated, "\n")

	breakPoint := lastSpace
	if lastNewline > lastSpace {
		breakPoint = lastNewline
	}

	if breakPoint > maxLength/2 {
		return text[:breakPoint] + "..."
	}

	return text[:maxLength] + "..."
}

func IsImageFormat(filename string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff", ".svg"}
	lower := strings.ToLower(filename)
	for _, ext := range extensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func (c *Config) CleanupInterval() time.Duration {
	// Run cleanup every hour
	return time.Hour
}

// ShouldSkipContent determines if content should be skipped from clipboard monitoring
func (c *Config) ShouldSkipContent(content string) bool {
	// Skip empty content
	if strings.TrimSpace(content) == "" {
		return true
	}

	// Skip very long content (>1MB) to avoid performance issues
	if len(content) > 1024*1024 {
		return true
	}

	// Skip content that looks like passwords (simple heuristic) unless allowed
	if !c.AllowPasswords && isLikelyPassword(content) {
		return true
	}

	return false
}

func isLikelyPassword(content string) bool {
	content = strings.TrimSpace(content)

	if len(content) < 8 {
		return false
	}

	if len(content) > 128 {
		return false
	}

	if strings.Contains(content, " ") ||
		strings.Contains(content, "\n") ||
		strings.Contains(content, "\t") {
		return false
	}

	hasUpper := strings.ToLower(content) != content
	hasLower := strings.ToUpper(content) != content
	hasDigit := strings.ContainsAny(content, "0123456789")
	hasSpecial := strings.ContainsAny(content, "!@#$%^&*()_+-=[]{}|;:,.<>?")

	charTypes := 0
	if hasUpper {
		charTypes++
	}
	if hasLower {
		charTypes++
	}
	if hasDigit {
		charTypes++
	}
	if hasSpecial {
		charTypes++
	}

	return charTypes >= 3
}
