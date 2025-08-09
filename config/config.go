package config

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"regexp"
	"unicode"
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
		GlobalHotkey:      "Cmd+Shift+Space",
		PreviousHotkey:    "Cmd+Shift+C",
		AutoLaunch:        true,
		EnableSounds:      false,
		AllowPasswords:    false,
	}
}

var (
	// Common patterns that are NOT passwords
	urlRegex      = regexp.MustCompile(`^https?://|^ftp://|^www\.`)
	filePathRegex = regexp.MustCompile(`^[a-zA-Z]:[\\\/]|^\/[^\/]|^\.\/|^\.\.\/|^\~\/`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Programming/code patterns
	functionCallRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*\(.*\)$`)
	methodCallRegex   = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\(\)$`)
	variableRegex     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*$`)

	// Common file extensions
	fileExtRegex = regexp.MustCompile(`\.[a-zA-Z0-9]{2,4}$`)

	// Base64 pattern - possibly password-like (TODO: more robust handling)
	base64Regex = regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)

	// API keys/tokens - potentially password-like
	apiKeyRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{32,}$`)

	// Common non-password words that might pass complexity checks
	commonNonPasswords = []string{
		"undefined", "function", "console.log", "document", "window",
		"localStorage", "sessionStorage", "className", "getElementById",
		"querySelector", "addEventListener", "preventDefault", "stopPropagation",
		"Promise.resolve", "JSON.stringify", "JSON.parse", "parseInt",
		"parseFloat", "toString", "valueOf", "hasOwnProperty", "iOS", "Android",
	}
)

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
	// TODO
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

	// Basic length checks
	if len(content) < 8 || len(content) > 128 {
		return false
	}

	// Check for whitespace (passwords usually don't have spaces/tabs/newlines)
	if strings.ContainsAny(content, " \n\t\r") {
		return false
	}

	// Check if it's a URL
	if urlRegex.MatchString(content) {
		return false
	}

	// Check if it's a file path
	if filePathRegex.MatchString(content) {
		return false
	}

	// Check if it's an email
	if emailRegex.MatchString(content) {
		return false
	}

	// Check if it's a function call (like "robotgo.Start()")
	if functionCallRegex.MatchString(content) || methodCallRegex.MatchString(content) {
		return false
	}

	// Check if it's a variable/property access
	if variableRegex.MatchString(content) {
		return false
	}

	// Check if it has a file extension
	if fileExtRegex.MatchString(content) {
		return false
	}

	// Check if it's likely Base64
	if len(content) > 20 && len(content)%4 == 0 && base64Regex.MatchString(content) {
		return true
	}

	// Check if it's an API key
	if len(content) > 32 && apiKeyRegex.MatchString(content) {
		return true
	}

	// Check against common non-password strings
	lowerContent := strings.ToLower(content)
	for _, nonPassword := range commonNonPasswords {
		if lowerContent == strings.ToLower(nonPassword) {
			return false
		}
	}

	// Check for programming language keywords/patterns
	if isProgrammingPattern(content) {
		return false
	}

	// Check character complexity
	hasUpper := strings.ToLower(content) != content
	hasLower := strings.ToUpper(content) != content
	hasDigit := containsDigit(content)
	hasSpecial := containsSpecialChar(content)

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

	// Must have at least 3 character types
	if charTypes < 3 {
		return false
	}

	// Additional heuristics for password-like content
	return hasPasswordLikePattern(content)
}

// Helper function to check for digits using unicode
func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// Helper function to check for special characters
func containsSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?/~`"
	return strings.ContainsAny(s, specialChars)
}

// Check if the content matches programming patterns
func isProgrammingPattern(content string) bool {
	patterns := []string{
		// JavaScript/TypeScript
		".then(", ".catch(", ".finally(", "async/await", "Promise",
		// Method chaining
		".map(", ".filter(", ".reduce(", ".forEach(",
		// Common object properties
		".length", ".prototype", ".constructor",
		// CSS/HTML-like
		"px", "em", "rem", "rgb(", "rgba(",
		// others
		"window", "document",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range patterns {
		if strings.Contains(lowerContent, strings.ToLower(pattern)) {
			return true
		}
	}

	// Check if it looks like a hex color code
	if len(content) == 6 && isHexString(content) {
		return true
	}

	return false
}

// Check if string is hexadecimal
func isHexString(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

// Additional heuristics to determine if content is password-like
func hasPasswordLikePattern(content string) bool {
	// Passwords often have random-looking character distribution

	// Check if it's all the same character repeated
	if isRepeatedChar(content) {
		return false
	}

	// Check if it follows common word patterns (like camelCase identifiers)
	if looksLikeCamelCase(content) {
		return false
	}

	// Check if it has too many consecutive identical characters
	if hasLongRepeatedSequence(content, 3) {
		return false
	}

	return true
}

// Check if string is just repeated characters
func isRepeatedChar(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}

// Check if string looks like camelCase identifier
func looksLikeCamelCase(s string) bool {
	// Must start with letter
	if !unicode.IsLetter(rune(s[0])) {
		return false
	}

	// Skip if it has slashes, dots, or other non-identifier chars
	if strings.ContainsAny(s, "/.@-+") {
		return false
	}

	hasUpper := false
	letterCount := 0

	for _, r := range s {
		if unicode.IsLetter(r) {
			letterCount++
			if unicode.IsUpper(r) {
				hasUpper = true
			}
		} else if !unicode.IsDigit(r) {
			return false
		}
	}

	// Must be mostly letters and have uppercase
	return hasUpper && letterCount > len(s)/2
}

// Check for long sequences of repeated characters
func hasLongRepeatedSequence(s string, maxLen int) bool {
	if len(s) < maxLen {
		return false
	}

	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= maxLen {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}
