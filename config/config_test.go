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
		{"verylongpasswordwithoutspaces123456", true, "long password-like"},
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

func TestIsHexString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"", true, "empty string"},
		{"0123456789", true, "all digits"},
		{"abcdef", true, "lowercase hex letters"},
		{"ABCDEF", true, "uppercase hex letters"},
		{"0123456789abcdefABCDEF", true, "mixed case hex"},
		{"deadbeef", true, "common hex pattern"},
		{"DEADBEEF", true, "uppercase hex pattern"},
		{"0xFF", false, "hex with prefix pattern - contains 'x'"},
		{"g", false, "invalid hex character"},
		{"123g456", false, "hex with invalid character"},
		{"hello", false, "non-hex string"},
		{"12 34", false, "hex with space"},
		{"12-34", false, "hex with dash"},
		{"0x123", false, "with 0x prefix"},
	}

	for _, test := range tests {
		result := isHexString(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestContainsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"", false, "empty string"},
		{"hello", false, "no digits"},
		{"Hello World", false, "no digits with spaces"},
		{"test123", true, "digits at end"},
		{"123test", true, "digits at start"},
		{"te5st", true, "digit in middle"},
		{"1", true, "single digit"},
		{"password1", true, "common password pattern"},
		{"MyPass0rd", true, "digit in password"},
		{"ABC", false, "uppercase letters only"},
		{"abc", false, "lowercase letters only"},
		{"!@#$%", false, "special characters only"},
	}

	for _, test := range tests {
		result := containsDigit(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestContainsSpecialChar(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"", false, "empty string"},
		{"hello", false, "no special chars"},
		{"Hello123", false, "letters and digits only"},
		{"test!", true, "exclamation mark"},
		{"pass@word", true, "at symbol"},
		{"secret#123", true, "hash symbol"},
		{"my$password", true, "dollar sign"},
		{"pwd%123", true, "percent sign"},
		{"test^case", true, "caret"},
		{"and&more", true, "ampersand"},
		{"star*fish", true, "asterisk"},
		{"(brackets)", true, "parentheses"},
		{"under_score", true, "underscore"},
		{"plus+sign", true, "plus sign"},
		{"equal=sign", true, "equal sign"},
		{"[array]", true, "square brackets"},
		{"{object}", true, "curly braces"},
		{"pipe|test", true, "pipe character"},
		{"semi;colon", true, "semicolon"},
		{"colon:test", true, "colon"},
		{"comma,test", true, "comma"},
		{"dot.test", true, "period"},
		{"less<than", true, "less than"},
		{"greater>than", true, "greater than"},
		{"question?mark", true, "question mark"},
		{"forward/slash", true, "forward slash"},
		{"tilde~test", true, "tilde"},
		{"backtick`test", true, "backtick"},
		{"back\\slash", false, "backslash - not in special chars list"},
		{"quote'test", false, "single quote - not in special chars list"},
		{"quote\"test", false, "double quote - not in special chars list"},
	}

	for _, test := range tests {
		result := containsSpecialChar(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestIsProgrammingPattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"", false, "empty string"},
		{"hello world", false, "normal text"},
		{"something.then()", true, "contains .then("},
		{"promise.catch()", true, "contains .catch("},
		{"async.finally()", true, "contains .finally("},
		{"async/await pattern", true, "contains async/await"},
		{"my Promise here", true, "contains Promise"},
		{"items.map()", true, "contains .map("},
		{"data.filter()", true, "contains .filter("},
		{"arr.reduce()", true, "contains .reduce("},
		{"list.forEach()", true, "contains .forEach("},
		{"string.length", true, "contains .length"},
		{"obj.prototype", true, "contains .prototype"},
		{"class.constructor", true, "contains .constructor"},
		{"margin: 10px", true, "contains px"},
		{"font-size: 2em", true, "contains em"},
		{"width: 1rem", true, "contains rem"},
		{"color: rgb(255,0,0)", true, "contains rgb("},
		{"background: rgba(0,0,0,0.5)", true, "contains rgba("},
		{"window.location", true, "contains window"},
		{"document.body", true, "contains document"},
		{"abcdef", true, "6-char hex string"},
		{"123456", true, "6-char hex digits"},
		{"console.log('hello')", false, "not in pattern list"},
		{"print('hello world')", false, "not in pattern list"},
		{"printf(\"hello\")", false, "not in pattern list"},
		{"import React from 'react'", false, "not in pattern list"},
		{"some regular text", false, "normal sentence"},
		{"This is a password", false, "normal sentence with password word"},
	}

	for _, test := range tests {
		result := isProgrammingPattern(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestHasPasswordLikePattern(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"hello", true, "simple word - doesn't match excluded patterns"},
		{"password", true, "common word - doesn't match excluded patterns"},
		{"mysecretpassword", true, "longer common words - doesn't match excluded patterns"},
		{"aGh7$mK9pL2@", true, "random password-like"},
		{"Xy9#bN4$rT6!", true, "complex password"},
		{"abcdefghijklmnop", true, "alphabetical sequence - doesn't match excluded patterns"},
		{"1234567890123456", true, "numerical sequence - doesn't match excluded patterns"},
		{"aaaaaaaaaaa", false, "repeated characters - excluded by isRepeatedChar"},
		{"Ab1!Ab1!Ab1!", true, "repeated pattern but not same char - not excluded"},
		{"MyPassword123!", true, "readable password - has ! so not camelCase, returns true"},
		{"4f8K2#nQ9@vL", true, "truly random looking"},
		{"x7Y!m3Z@p9W$", true, "mixed case with symbols"},
		{"randomStringHere", false, "readable camelCase - excluded by looksLikeCamelCase"},
		{"API_KEY_12345", true, "readable constant - has underscore so not camelCase"},
		{"secretkey", true, "readable compound word - all lowercase so not camelCase"},
		{"2f4a8c1b9e5d", true, "hex-like random"},
		{"Zx9!Qm3#Vn7@", true, "high entropy password"},
		{"mypassword123", true, "common pattern - not camelCase"},
		{"Password!123", true, "common password format - has special chars so not camelCase"},
		{"aaabbb", false, "has 3+ consecutive chars - excluded by hasLongRepeatedSequence"},
		{"111", false, "has 3+ consecutive chars - excluded by hasLongRepeatedSequence"},
	}

	for _, test := range tests {
		result := hasPasswordLikePattern(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestIsRepeatedChar(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"", false, "empty string"},
		{"a", true, "single character"},
		{"aa", true, "two same characters"},
		{"aaa", true, "three same characters"},
		{"aaaa", true, "four same characters"},
		{"1111", true, "repeated digits"},
		{"!!!!", true, "repeated special chars"},
		{"    ", true, "repeated spaces"},
		{"ab", false, "two different characters"},
		{"abc", false, "three different characters"},
		{"aba", false, "alternating pattern"},
		{"hello", false, "normal word"},
		{"aab", false, "mixed characters"},
		{"112", false, "mixed digits"},
		{"a1a", false, "mixed alphanumeric"},
	}

	for _, test := range tests {
		result := isRepeatedChar(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestLooksLikeCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"hello", false, "all lowercase - no uppercase"},
		{"HELLO", true, "all uppercase - has uppercase and mostly letters"},
		{"Hello", true, "single capital at start - has uppercase and mostly letters"},
		{"hello world", false, "space separated - space not allowed"},
		{"helloWorld", true, "camelCase"},
		{"myVariableName", true, "longer camelCase"},
		{"firstName", true, "common camelCase"},
		{"XMLHttpRequest", true, "camelCase with acronym"},
		{"getElementById", true, "DOM method name"},
		{"innerHTML", true, "DOM property"},
		{"backgroundColor", true, "CSS property in camelCase"},
		{"onClickHandler", true, "event handler name"},
		{"apiResponse", true, "API related camelCase"},
		{"userAccountInfo", true, "longer camelCase"},
		{"myPasswordHere", true, "camelCase that might be password"},
		{"thisIsMySecretKey", true, "camelCase secret"},
		{"hello_world", false, "snake_case - has underscore (not in excluded list but space will fail)"},
		{"HELLO_WORLD", false, "SCREAMING_SNAKE_CASE - has underscore (not in excluded list but space will fail)"},
		{"hello-world", false, "kebab-case - has dash"},
		{"Hello World", false, "Title Case with spaces"},
		{"helloWorld123", true, "camelCase with numbers"},
		{"myVar2", true, "camelCase ending with number"},
		{"123hello", false, "starts with number"},
		{"h", false, "single character - no uppercase"},
		{"H", true, "single uppercase character"},
		{"hH", true, "minimal camelCase"},
		{"hello.world", false, "has dot"},
		{"hello/world", false, "has slash"},
		{"hello@world", false, "has at symbol"},
		{"hello+world", false, "has plus"},
		{"hello_world", false, "has underscore - not excluded by special chars but fails other criteria"},
		{"A1", false, "uppercase letter + digit, 1/2 = 0.5, not > 0.5 so fails"},
		{"A12", false, "uppercase + 2 digits, not mostly letters (1/3 < 0.5)"},
	}

	for _, test := range tests {
		result := looksLikeCamelCase(test.input)
		assert.Equal(t, test.expected, result, "Input: %s (%s)", test.input, test.desc)
	}
}

func TestHasLongRepeatedSequence(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected bool
		desc     string
	}{
		{"", 3, false, "empty string"},
		{"abc", 3, false, "short string"},
		{"aaa", 3, true, "3 consecutive 'a' chars"},
		{"abcaaa", 3, true, "3 consecutive 'a' at end"},
		{"aaabcd", 3, true, "3 consecutive 'a' at start"},
		{"abaaabcd", 3, true, "3 consecutive 'a' in middle"},
		{"abcdefg", 3, false, "no repetition"},
		{"aabbcc", 2, true, "2 consecutive 'a' chars"},
		{"aabbcc", 3, false, "no 3 consecutive chars"},
		{"hello", 2, true, "2 consecutive 'l' chars"},
		{"hellllo", 3, true, "4 consecutive 'l' chars"},
		{"abcdef", 2, false, "no consecutive chars"},
		{"aabbccdd", 2, true, "multiple 2-char sequences"},
		{"111", 3, true, "3 consecutive '1' digits"},
		{"1223334444", 4, true, "4 consecutive '4' digits"},
		{"1223334444", 5, false, "no 5 consecutive chars"},
		{"   ", 3, true, "3 consecutive spaces"},
		{"!!!", 3, true, "3 consecutive exclamation marks"},
		{"a", 1, false, "single char - need at least 2 chars for a sequence"},
		{"aa", 1, true, "2 consecutive chars, maxLen 1"},
		{"ab", 1, false, "no consecutive chars, maxLen 1"},
	}

	for _, test := range tests {
		result := hasLongRepeatedSequence(test.input, test.maxLen)
		assert.Equal(t, test.expected, result, "Input: %s, maxLen: %d (%s)", test.input, test.maxLen, test.desc)
	}
}
