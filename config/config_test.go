package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 500*time.Millisecond, cfg.PollingInterval)
	assert.Equal(t, 100, cfg.MaxItems)
	assert.Equal(t, 7, cfg.MaxDays)
	assert.True(t, cfg.MonitoringEnabled)
	assert.Equal(t, "Cmd+Shift+Space", cfg.GlobalHotkey)
	assert.Equal(t, "Cmd+Shift+C", cfg.PreviousHotkey)
	assert.True(t, cfg.AutoLaunch)
	assert.False(t, cfg.EnableSounds)
	assert.False(t, cfg.AllowPasswords)
}

func TestUpdateFromSettings(t *testing.T) {
	cfg := NewConfig()

	settings := map[string]interface{}{
		"pollingInterval":    1000,
		"maxItems":           200,
		"maxDays":            14,
		"monitoringEnabled":  false,
		"globalHotkey":       "Ctrl+V",
		"previousItemHotkey": "Ctrl+Shift+V",
		"autoLaunch":         false,
		"enableSounds":       true,
		"allowPasswords":     true,
	}

	cfg.UpdateFromSettings(settings)

	assert.Equal(t, 1000*time.Millisecond, cfg.PollingInterval)
	assert.Equal(t, 200, cfg.MaxItems)
	assert.Equal(t, 14, cfg.MaxDays)
	assert.False(t, cfg.MonitoringEnabled)
	assert.Equal(t, "Ctrl+V", cfg.GlobalHotkey)
	assert.Equal(t, "Ctrl+Shift+V", cfg.PreviousHotkey)
	assert.False(t, cfg.AutoLaunch)
	assert.True(t, cfg.EnableSounds)
	assert.True(t, cfg.AllowPasswords)
}

func TestUpdateFromSettingsPartial(t *testing.T) {
	cfg := NewConfig()
	originalMaxItems := cfg.MaxItems

	// Update only some settings
	settings := map[string]interface{}{
		"pollingInterval": 2000,
		"maxDays":         30,
	}

	cfg.UpdateFromSettings(settings)

	// Updated values
	assert.Equal(t, 2000*time.Millisecond, cfg.PollingInterval)
	assert.Equal(t, 30, cfg.MaxDays)

	// Unchanged values
	assert.Equal(t, originalMaxItems, cfg.MaxItems)
	assert.True(t, cfg.MonitoringEnabled)
}

func TestUpdateFromSettingsInvalidTypes(t *testing.T) {
	cfg := NewConfig()
	originalPolling := cfg.PollingInterval

	// Provide invalid types
	settings := map[string]interface{}{
		"pollingInterval": "invalid", // should be int
		"maxItems":        "invalid", // should be int
		"enableSounds":    "invalid", // should be bool
	}

	cfg.UpdateFromSettings(settings)

	// Values should remain unchanged
	assert.Equal(t, originalPolling, cfg.PollingInterval)
	assert.Equal(t, 100, cfg.MaxItems)
	assert.False(t, cfg.EnableSounds)
}

func TestContentTypeString(t *testing.T) {
	assert.Equal(t, "text", ContentTypeText.String())
	assert.Equal(t, "image", ContentTypeImage.String())
	assert.Equal(t, "file", ContentTypeFile.String())

	// Test unknown content type
	var unknown ContentType = 99
	assert.Equal(t, "unknown", unknown.String())
}

func TestParseContentType(t *testing.T) {
	assert.Equal(t, ContentTypeText, ParseContentType("text"))
	assert.Equal(t, ContentTypeText, ParseContentType("TEXT"))
	assert.Equal(t, ContentTypeImage, ParseContentType("image"))
	assert.Equal(t, ContentTypeImage, ParseContentType("IMAGE"))
	assert.Equal(t, ContentTypeFile, ParseContentType("file"))
	assert.Equal(t, ContentTypeFile, ParseContentType("FILE"))
	assert.Equal(t, ContentTypeText, ParseContentType("unknown"))
	assert.Equal(t, ContentTypeText, ParseContentType(""))
}

func TestGenerateHash(t *testing.T) {
	content1 := "Hello, World!"
	content2 := "Hello, World!"
	content3 := "Different content"

	hash1 := GenerateHash(content1)
	hash2 := GenerateHash(content2)
	hash3 := GenerateHash(content3)

	// Same content should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different content should produce different hash
	assert.NotEqual(t, hash1, hash3)

	// Hash should not be empty
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash3)

	// Hash should be consistent
	assert.Equal(t, hash1, GenerateHash(content1))
}

func TestTruncatePreview(t *testing.T) {
	tests := []struct {
		text      string
		maxLength int
		expected  string
	}{
		{"Short text", 100, "Short text"},
		{"", 100, ""},
		{"Exactly ten characters!", 25, "Exactly ten characters!"},
		{"This is a very long text that needs to be truncated", 20, "This is a very long..."},
		{"No spaces in this verylongtext", 15, "No spaces in..."},
		{"Text with\nnewlines and spaces", 20, "Text with\nnewlines..."},
		{"Multiple    spaces", 10, "Multiple..."},
	}

	for _, test := range tests {
		result := TruncatePreview(test.text, test.maxLength)

		if len(test.text) <= test.maxLength {
			assert.Equal(t, test.expected, result)
		} else {
			// Should be truncated and have "..."
			assert.LessOrEqual(t, len(result), test.maxLength+3) // +3 for "..."
			if len(test.text) > test.maxLength {
				assert.Contains(t, result, "...")
			}
		}
	}
}

func TestIsImageFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"image.jpg", true},
		{"photo.jpeg", true},
		{"picture.png", true},
		{"animated.gif", true},
		{"bitmap.bmp", true},
		{"modern.webp", true},
		{"high-res.tiff", true},
		{"vector.svg", true},
		{"IMAGE.JPG", true}, // Case insensitive
		{"Photo.PNG", true},
		{"document.pdf", false},
		{"text.txt", false},
		{"video.mp4", false},
		{"archive.zip", false},
		{"", false},
		{"no-extension", false},
		{".jpg", true},          // Just extension
		{"file.jpg.txt", false}, // False positive check
	}

	for _, test := range tests {
		result := IsImageFormat(test.filename)
		assert.Equal(t, test.expected, result, "File: %s", test.filename)
	}
}

func TestCleanupInterval(t *testing.T) {
	cfg := NewConfig()
	interval := cfg.CleanupInterval()
	assert.Equal(t, time.Hour, interval)
}

func TestShouldSkipContent(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		content  string
		expected bool
		desc     string
	}{
		{"", true, "empty content"},
		{"   ", true, "whitespace only"},
		{"\n\t ", true, "whitespace and newlines"},
		{"Normal text", false, "normal text"},
		{"Hello, World!", false, "simple text"},
		{"Multi\nline\ntext", false, "multiline text"},
		{"Password123!@#", true, "password-like content (passwords disabled by default)"},
		{"MySecurePass!", true, "another password-like content (passwords disabled by default)"},
		{"simple password", false, "text with spaces (not password)"},
		{"verylongpasswordwithoutspaces123456", false, "long password-like"}, // Actually not detected as password
		{"Complex123!@#$%^&*()", true, "complex password-like"},
	}

	for _, test := range tests {
		result := cfg.ShouldSkipContent(test.content)
		assert.Equal(t, test.expected, result, "Content: %s (%s)", test.content, test.desc)
	}
}

func TestShouldSkipContentLongContent(t *testing.T) {
	cfg := NewConfig()

	// Create content larger than 1MB
	longContent := make([]byte, 1024*1024+1)
	for i := range longContent {
		longContent[i] = 'a'
	}

	result := cfg.ShouldSkipContent(string(longContent))
	assert.True(t, result, "Very long content should be skipped")
}

func TestIsLikelyPasswordHeuristics(t *testing.T) {
	// Test the password detection heuristics indirectly through ShouldSkipContent
	cfg := NewConfig() // passwords disabled by default

	tests := []struct {
		content     string
		isPassword  bool
		description string
	}{
		{"short", false, "too short to be a password we care about"},
		{"Password123!", true, "typical password pattern"},
		{"MySecure1!", true, "another password pattern"},
		{"hello world", false, "contains spaces"},
		{"multi\nline\ntext", false, "contains newlines"},
		{"simple text with spaces", false, "normal text"},
		{"UPPERCASE", false, "only uppercase, no other types"},
		{"lowercase", false, "only lowercase, no other types"},
		{"123456789", false, "only digits"},
		{"ComplexPassword1234567890!@#$%^&*()_+-=[]{}|;:,.<>?ABCDEFGHIJKLMNOPQRSTUVWXYZ", true, "too long but still detected as password due to complexity"},
		{"", false, "empty string"},
		{"VeryComplexP@ssw0rd!", true, "complex password with multiple character types"},
		{"Ab1!", false, "too short (less than 8 chars)"},
		{"SimpleP@ss1", true, "meets password criteria"},
	}

	for _, test := range tests {
		result := cfg.ShouldSkipContent(test.content)
		if test.isPassword {
			assert.True(t, result, "Should skip password-like content: %s (%s)", test.content, test.description)
		} else {
			// Note: We only test that passwords are detected, not that non-passwords are kept
			// because ShouldSkipContent has other rules (empty content, very long content, etc.)
			if len(test.content) > 0 && len(test.content) < 1024*1024 {
				// Only test non-empty, reasonable-length content
				expected := test.isPassword
				assert.Equal(t, expected, result, "Content: %s (%s)", test.content, test.description)
			}
		}
	}
}

func TestShouldSkipContentWithPasswordsAllowed(t *testing.T) {
	cfg := NewConfig()
	cfg.AllowPasswords = true

	tests := []struct {
		content  string
		expected bool
		desc     string
	}{
		{"", true, "empty content"},
		{"   ", true, "whitespace only"},
		{"Normal text", false, "normal text"},
		{"Password123!@#", false, "password-like content (passwords enabled)"},
		{"MySecurePass!", false, "another password-like content (passwords enabled)"},
		{"Complex123!@#$%^&*()", false, "complex password-like (passwords enabled)"},
		{"VeryComplexP@ssw0rd!", false, "complex password with multiple character types (passwords enabled)"},
	}

	for _, test := range tests {
		result := cfg.ShouldSkipContent(test.content)
		assert.Equal(t, test.expected, result, "Content: %s (%s)", test.content, test.desc)
	}
}
