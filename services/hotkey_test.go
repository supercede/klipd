package services

import (
	"fmt"
	"sync"
	"testing"
	"time"

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

func TestHotkeyManagerRegister(t *testing.T) {
	hm := NewHotkeyManager()
	callbackCalled := false

	callback := func() {
		callbackCalled = true
	}

	// Test registering a hotkey
	err := hm.Register("Cmd+Shift+V", callback)
	assert.NoError(t, err)

	// Verify it was registered
	hm.mu.RLock()
	assert.True(t, hm.registered["Cmd+Shift+V"])
	assert.NotNil(t, hm.callbacks["Cmd+Shift+V"])
	hm.mu.RUnlock()

	// Test triggering the hotkey manually
	hm.TriggerHotkey("Cmd+Shift+V")

	// Give it a moment to execute (it runs in a goroutine)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, callbackCalled)
}

func TestHotkeyManagerRegisterMultiple(t *testing.T) {
	hm := NewHotkeyManager()

	callback1Called := false
	callback2Called := false

	callback1 := func() { callback1Called = true }
	callback2 := func() { callback2Called = true }

	// Register multiple hotkeys
	err1 := hm.Register("Cmd+Shift+V", callback1)
	err2 := hm.Register("Cmd+Shift+C", callback2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify both are registered
	hm.mu.RLock()
	assert.True(t, hm.registered["Cmd+Shift+V"])
	assert.True(t, hm.registered["Cmd+Shift+C"])
	assert.Equal(t, 2, len(hm.callbacks))
	assert.Equal(t, 2, len(hm.registered))
	hm.mu.RUnlock()

	// Test triggering each hotkey
	hm.TriggerHotkey("Cmd+Shift+V")
	time.Sleep(10 * time.Millisecond)
	assert.True(t, callback1Called)
	assert.False(t, callback2Called)

	hm.TriggerHotkey("Cmd+Shift+C")
	time.Sleep(10 * time.Millisecond)
	assert.True(t, callback2Called)
}

func TestHotkeyManagerUnregister(t *testing.T) {
	hm := NewHotkeyManager()

	callback := func() {}

	// Register a hotkey
	err := hm.Register("Cmd+Shift+V", callback)
	assert.NoError(t, err)

	// Verify it's registered
	hm.mu.RLock()
	assert.True(t, hm.registered["Cmd+Shift+V"])
	assert.NotNil(t, hm.callbacks["Cmd+Shift+V"])
	hm.mu.RUnlock()

	// Unregister the hotkey
	hm.Unregister("Cmd+Shift+V")

	// Verify it's unregistered
	hm.mu.RLock()
	assert.False(t, hm.registered["Cmd+Shift+V"])
	assert.Nil(t, hm.callbacks["Cmd+Shift+V"])
	assert.Equal(t, 0, len(hm.callbacks))
	assert.Equal(t, 0, len(hm.registered))
	hm.mu.RUnlock()
}

func TestHotkeyManagerUnregisterNonexistent(t *testing.T) {
	hm := NewHotkeyManager()

	// Try to unregister a hotkey that was never registered
	// This should not panic or cause errors
	hm.Unregister("Cmd+Shift+V")

	// Verify maps are still empty
	hm.mu.RLock()
	assert.Equal(t, 0, len(hm.callbacks))
	assert.Equal(t, 0, len(hm.registered))
	hm.mu.RUnlock()
}

func TestHotkeyManagerStart(t *testing.T) {
	hm := NewHotkeyManager()

	// Initially not running
	assert.False(t, hm.IsRunning())

	// Start the manager
	err := hm.Start()
	assert.NoError(t, err)
	assert.True(t, hm.IsRunning())

	// Try to start again (should return error)
	err = hm.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
	assert.True(t, hm.IsRunning())
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

	// Try to stop again (should not cause issues)
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

func TestHotkeyManagerTriggerHotkey(t *testing.T) {
	hm := NewHotkeyManager()

	callbackCalled := false
	callback := func() {
		callbackCalled = true
	}

	// Register a hotkey
	err := hm.Register("Cmd+Shift+V", callback)
	require.NoError(t, err)

	// Trigger the hotkey
	hm.TriggerHotkey("Cmd+Shift+V")

	// Wait for goroutine to execute
	time.Sleep(10 * time.Millisecond)
	assert.True(t, callbackCalled)
}

func TestHotkeyManagerTriggerNonexistentHotkey(t *testing.T) {
	hm := NewHotkeyManager()

	// Try to trigger a hotkey that doesn't exist
	// This should not panic or cause errors
	hm.TriggerHotkey("Cmd+Shift+V")

	// Nothing should happen, test passes if no panic
}

func TestHotkeyManagerCallbackPanic(t *testing.T) {
	hm := NewHotkeyManager()

	// Register a callback that panics
	callback := func() {
		panic("test panic")
	}

	err := hm.Register("Cmd+Shift+V", callback)
	require.NoError(t, err)

	// Trigger the hotkey - should not crash the test
	hm.TriggerHotkey("Cmd+Shift+V")

	// Wait for goroutine to execute and recover from panic
	time.Sleep(10 * time.Millisecond)

	// Test passes if we get here without crashing
}

func TestHotkeyManagerConcurrency(t *testing.T) {
	hm := NewHotkeyManager()

	var callCount int
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	// Register hotkey
	err := hm.Register("Cmd+Shift+V", callback)
	require.NoError(t, err)

	// Start the manager
	err = hm.Start()
	require.NoError(t, err)

	// Trigger hotkey multiple times concurrently
	var wg sync.WaitGroup
	numTriggers := 10

	for i := 0; i < numTriggers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hm.TriggerHotkey("Cmd+Shift+V")
		}()
	}

	wg.Wait()

	// Wait for all callbacks to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, numTriggers, callCount)
	mu.Unlock()

	// Stop the manager
	hm.Stop()
}

func TestHotkeyManagerRegisterUnregisterConcurrency(t *testing.T) {
	hm := NewHotkeyManager()

	var wg sync.WaitGroup
	numOperations := 10

	// Concurrently register and unregister hotkeys
	for i := 0; i < numOperations; i++ {
		wg.Add(2)

		go func(index int) {
			defer wg.Done()
			hotkeyStr := fmt.Sprintf("Cmd+Shift+%d", index)
			callback := func() {}
			hm.Register(hotkeyStr, callback)
		}(i)

		go func(index int) {
			defer wg.Done()
			hotkeyStr := fmt.Sprintf("Cmd+Shift+%d", index)
			// Add small delay to allow registration to happen first sometimes
			time.Sleep(1 * time.Millisecond)
			hm.Unregister(hotkeyStr)
		}(i)
	}

	wg.Wait()

	// Test passes if no race conditions or panics occur
}

func TestHotkeyManagerStartStopConcurrency(t *testing.T) {
	hm := NewHotkeyManager()

	var wg sync.WaitGroup
	numOperations := 10

	// Concurrently start and stop the manager
	for i := 0; i < numOperations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			hm.Start()
		}()

		go func() {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			hm.Stop()
		}()
	}

	wg.Wait()

	// Test passes if no race conditions or panics occur
}

func TestHotkeyManagerOverwriteRegistration(t *testing.T) {
	hm := NewHotkeyManager()

	callback1Called := false
	callback2Called := false

	callback1 := func() { callback1Called = true }
	callback2 := func() { callback2Called = true }

	// Register first callback
	err := hm.Register("Cmd+Shift+V", callback1)
	require.NoError(t, err)

	// Register second callback with same hotkey (should overwrite)
	err = hm.Register("Cmd+Shift+V", callback2)
	require.NoError(t, err)

	// Verify only one registration exists
	hm.mu.RLock()
	assert.Equal(t, 1, len(hm.callbacks))
	assert.Equal(t, 1, len(hm.registered))
	hm.mu.RUnlock()

	// Trigger hotkey - should call second callback only
	hm.TriggerHotkey("Cmd+Shift+V")
	time.Sleep(10 * time.Millisecond)

	assert.False(t, callback1Called)
	assert.True(t, callback2Called)
}

func TestHotkeyManagerEmptyHotkeyString(t *testing.T) {
	hm := NewHotkeyManager()

	callback := func() {}

	// Register empty hotkey string
	err := hm.Register("", callback)
	assert.NoError(t, err)

	// Should be registered
	hm.mu.RLock()
	assert.True(t, hm.registered[""])
	assert.NotNil(t, hm.callbacks[""])
	hm.mu.RUnlock()

	// Should be triggerable
	callbackCalled := false
	callback = func() { callbackCalled = true }
	hm.Register("", callback)

	hm.TriggerHotkey("")
	time.Sleep(10 * time.Millisecond)
	assert.True(t, callbackCalled)
}
