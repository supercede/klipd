package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHotkeyManager(t *testing.T) {
	hm := NewHotkeyManager()

	assert.NotNil(t, hm)
	assert.NotNil(t, hm.callbacks)
	assert.NotNil(t, hm.registered)
	assert.False(t, hm.isRunning)
	assert.Equal(t, 0, len(hm.callbacks))
	assert.Equal(t, 0, len(hm.registered))
}

func TestHotkeyManagerStart(t *testing.T) {
	hm := NewHotkeyManager()

	// Initially not running
	assert.False(t, hm.IsRunning())

	// Start the manager
	err := hm.Start()
	assert.NoError(t, err)
	assert.True(t, hm.IsRunning())

	// Try to start again (should succeed since it's just a state change)
	err = hm.Start()
	assert.NoError(t, err)
	assert.True(t, hm.IsRunning())

	// Cleanup
	hm.Stop()
}

func TestHotkeyManagerStop(t *testing.T) {
	hm := NewHotkeyManager()

	// Start the manager
	err := hm.Start()
	require.NoError(t, err)
	assert.True(t, hm.IsRunning())

	// Stop the manager
	hm.Stop()
	assert.False(t, hm.IsRunning())

	// Test stopping again (should not cause issues)
	hm.Stop()
	assert.False(t, hm.IsRunning())
}

func TestHotkeyManagerIsRunning(t *testing.T) {
	hm := NewHotkeyManager()

	// Initially not running
	assert.False(t, hm.IsRunning())

	// Start and check
	err := hm.Start()
	require.NoError(t, err)
	assert.True(t, hm.IsRunning())

	// Stop and check
	hm.Stop()
	assert.False(t, hm.IsRunning())
}

func TestHotkeyManagerRegisterInvalidKey(t *testing.T) {
	hm := NewHotkeyManager()

	callback := func() {}

	// Test invalid key
	err := hm.Register("Cmd+Shift+INVALID", callback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown key")

	// Test invalid modifier
	err = hm.Register("INVALID+V", callback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown modifier")

	// Test empty string
	err = hm.Register("", callback)
	assert.Error(t, err)

	// Cleanup
	hm.Stop()
}

func TestParseHotkey(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
		description string
	}{
		{"Cmd+Shift+V", false, "valid hotkey"},
		{"Ctrl+C", false, "valid ctrl hotkey"},
		{"Alt+Tab", false, "valid alt hotkey"},
		{"Shift+Space", false, "valid shift space"},
		{"V", false, "single key"},
		{"", true, "empty string"},
		{"Cmd+INVALID", true, "invalid key"},
		{"INVALID+V", true, "invalid modifier"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mods, key, err := parseHotkey(test.input)
			if test.expectError {
				assert.Error(t, err, "Expected error for input: %s", test.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", test.input)
				// Modifiers can be empty slice for single keys, that's valid
				_ = mods // Ignore modifiers check since empty slice is valid
				assert.NotEqual(t, 0, key, "Key should not be zero")
			}
		})
	}
}

func TestKeyMap(t *testing.T) {
	// Test that all expected keys are in the map
	expectedKeys := []string{"A", "B", "C", "SPACE", "DELETE", "RETURN", "ESCAPE", "TAB"}

	for _, key := range expectedKeys {
		_, exists := keyMap[key]
		assert.True(t, exists, "Key %s should exist in keyMap", key)
	}
}
