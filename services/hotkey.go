package services

import (
	"fmt"
	"log"
	"sync"
)

// HotkeyCallback represents a function to be called when a hotkey is pressed
type HotkeyCallback func()

// HotkeyManager manages global hotkeys
// TODO: global hotkey support
type HotkeyManager struct {
	mu         sync.RWMutex
	isRunning  bool
	callbacks  map[string]HotkeyCallback
	registered map[string]bool
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager() *HotkeyManager {
	return &HotkeyManager{
		callbacks:  make(map[string]HotkeyCallback),
		registered: make(map[string]bool),
	}
}

// Register registers a global hotkey with a callback
func (hm *HotkeyManager) Register(hotkeyStr string, callback HotkeyCallback) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Store the callback
	hm.callbacks[hotkeyStr] = callback
	hm.registered[hotkeyStr] = true

	log.Printf("Registered hotkey: %s (Note: Global hotkeys not yet implemented)", hotkeyStr)
	return nil
}

// Unregister removes a hotkey registration
func (hm *HotkeyManager) Unregister(hotkeyStr string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.callbacks, hotkeyStr)
	delete(hm.registered, hotkeyStr)

	log.Printf("Unregistered hotkey: %s", hotkeyStr)
}

// Start begins listening for hotkey events
func (hm *HotkeyManager) Start() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.isRunning {
		return fmt.Errorf("hotkey manager is already running")
	}

	hm.isRunning = true

	log.Println("Hotkey manager started (simplified implementation)")
	return nil
}

// Stop stops the hotkey manager
func (hm *HotkeyManager) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.isRunning {
		return
	}

	hm.isRunning = false

	log.Println("Hotkey manager stopped")
}

// IsRunning returns whether the hotkey manager is currently running
func (hm *HotkeyManager) IsRunning() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.isRunning
}

// TriggerHotkey manually triggers a hotkey (for testing purposes)
func (hm *HotkeyManager) TriggerHotkey(hotkeyStr string) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if callback, exists := hm.callbacks[hotkeyStr]; exists {
		log.Printf("Manually triggering hotkey: %s", hotkeyStr)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Hotkey callback panic: %v", r)
				}
			}()
			callback()
		}()
	}
}
