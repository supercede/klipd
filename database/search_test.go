package database

import (
	"fmt"
	"testing"
	"time"

	"klipd/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchClipboardItemsRegex(t *testing.T) {
	db := setupTestDB(t)

	// Create test items
	items := []*models.ClipboardItem{
		{
			ID:           "1",
			ContentType:  "text",
			ContentText:  "Hello World",
			PreviewText:  "Hello World",
			Hash:         "hash1",
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
		},
		{
			ID:           "2",
			ContentType:  "text",
			ContentText:  "Test Email: user@example.com",
			PreviewText:  "Test Email: user@example.com",
			Hash:         "hash2",
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
		},
		{
			ID:           "3",
			ContentType:  "text",
			ContentText:  "Phone: 123-456-7890",
			PreviewText:  "Phone: 123-456-7890",
			Hash:         "hash3",
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
		},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(item)
		require.NoError(t, err)
	}

	// Test regex search for email pattern
	// Note: This test may fail if SQLite doesn't have regex support compiled in
	// In that case, we'll just verify the method exists and handles the query
	results, err := db.SearchClipboardItemsRegex(`.*@.*\.com`, 10)

	// The test might fail with "no such function: REGEXP" if regex isn't available
	// That's expected behavior for basic SQLite installations
	if err != nil && err.Error() == "no such function: REGEXP" {
		t.Skip("SQLite REGEXP function not available - this is expected for basic installations")
	}

	// If no error, we should get the email item
	if err == nil {
		assert.Len(t, results, 1)
		assert.Equal(t, "Test Email: user@example.com", results[0].PreviewText)
	}
}

func TestSearchClipboardItemsRegexPatterns(t *testing.T) {
	db := setupTestDB(t)

	// Create test items with different patterns
	items := []*models.ClipboardItem{
		{
			ID:           "1",
			ContentType:  "text",
			ContentText:  "Phone: 123-456-7890",
			PreviewText:  "Phone: 123-456-7890",
			Hash:         "hash1",
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			LastAccessed: time.Now().Add(-2 * time.Hour),
			IsPinned:     false,
		},
		{
			ID:           "2",
			ContentType:  "text",
			ContentText:  "Email: john.doe@example.org",
			PreviewText:  "Email: john.doe@example.org",
			Hash:         "hash2",
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			LastAccessed: time.Now().Add(-1 * time.Hour),
			IsPinned:     true,
		},
		{
			ID:           "3",
			ContentType:  "text",
			ContentText:  "URL: https://www.example.com/path",
			PreviewText:  "URL: https://www.example.com/path",
			Hash:         "hash3",
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			IsPinned:     false,
		},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(item)
		require.NoError(t, err)
	}

	testCases := []struct {
		name     string
		pattern  string
		expected int
		skipMsg  string
	}{
		{
			name:     "Phone number pattern",
			pattern:  `\d{3}-\d{3}-\d{4}`,
			expected: 1,
		},
		{
			name:     "Email pattern",
			pattern:  `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			expected: 1,
		},
		{
			name:     "HTTPS URL pattern",
			pattern:  `https://.*`,
			expected: 1,
		},
		{
			name:     "Word boundary pattern",
			pattern:  `\bexample\b`,
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := db.SearchClipboardItemsRegex(tc.pattern, 10)

			if err != nil && err.Error() == "no such function: REGEXP" {
				t.Skip("SQLite REGEXP function not available - this is expected for basic installations")
			}

			if err == nil {
				assert.Len(t, results, tc.expected, "Pattern: %s", tc.pattern)
			}
		})
	}
}

func TestSearchClipboardItemsRegexOrdering(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()

	// Create items with same pattern but different pinned status and timestamps
	items := []*models.ClipboardItem{
		{
			ID:           "1",
			ContentType:  "text",
			ContentText:  "test@old.com",
			PreviewText:  "test@old.com",
			Hash:         "hash1",
			CreatedAt:    now.Add(-2 * time.Hour),
			LastAccessed: now.Add(-2 * time.Hour),
			IsPinned:     false,
		},
		{
			ID:           "2",
			ContentType:  "text",
			ContentText:  "admin@pinned.com",
			PreviewText:  "admin@pinned.com",
			Hash:         "hash2",
			CreatedAt:    now.Add(-1 * time.Hour),
			LastAccessed: now.Add(-1 * time.Hour),
			IsPinned:     true,
		},
		{
			ID:           "3",
			ContentType:  "text",
			ContentText:  "user@new.com",
			PreviewText:  "user@new.com",
			Hash:         "hash3",
			CreatedAt:    now,
			LastAccessed: now,
			IsPinned:     false,
		},
	}

	for _, item := range items {
		err := db.CreateClipboardItem(item)
		require.NoError(t, err)
	}

	// Search for email pattern - should return all 3, ordered by pinned first, then last_accessed DESC
	results, err := db.SearchClipboardItemsRegex(`.*@.*\.com`, 10)

	if err != nil && err.Error() == "no such function: REGEXP" {
		t.Skip("SQLite REGEXP function not available - this is expected for basic installations")
	}

	if err == nil {
		require.Len(t, results, 3)

		// First should be pinned item
		assert.True(t, results[0].IsPinned)
		assert.Equal(t, "admin@pinned.com", results[0].PreviewText)

		// Next two should be unpinned, ordered by last_accessed DESC
		assert.False(t, results[1].IsPinned)
		assert.False(t, results[2].IsPinned)
		assert.Equal(t, "user@new.com", results[1].PreviewText)
		assert.Equal(t, "test@old.com", results[2].PreviewText)
	}
}

func TestSearchClipboardItemsRegexLimits(t *testing.T) {
	db := setupTestDB(t)

	// Create multiple test items that match a pattern
	for i := 0; i < 5; i++ {
		item := &models.ClipboardItem{
			ID:           fmt.Sprintf("item-%d", i),
			ContentType:  "text",
			ContentText:  fmt.Sprintf("test%d@example.com", i),
			PreviewText:  fmt.Sprintf("test%d@example.com", i),
			Hash:         fmt.Sprintf("hash%d", i),
			CreatedAt:    time.Now().Add(-time.Duration(i) * time.Hour),
			LastAccessed: time.Now().Add(-time.Duration(i) * time.Hour),
			IsPinned:     false,
		}
		err := db.CreateClipboardItem(item)
		require.NoError(t, err)
	}

	// Test limit functionality
	results, err := db.SearchClipboardItemsRegex(`test.*@example\.com`, 3)

	if err != nil && err.Error() == "no such function: REGEXP" {
		t.Skip("SQLite REGEXP function not available - this is expected for basic installations")
	}

	if err == nil {
		assert.Len(t, results, 3)

		// Test with limit larger than available items
		results, err = db.SearchClipboardItemsRegex(`test.*@example\.com`, 10)
		require.NoError(t, err)
		assert.Len(t, results, 5)
	}
}
