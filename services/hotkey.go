package services

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"

	"golang.design/x/hotkey"
)

// HotkeyCallback represents a function to be called when a hotkey is pressed
type HotkeyCallback func()

// HotkeyManager manages global hotkeys using golang.design/x/hotkey
type HotkeyManager struct {
	mu         sync.RWMutex
	isRunning  bool
	callbacks  map[string]HotkeyCallback
	registered map[string]*hotkey.Hotkey
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager() *HotkeyManager {
	return &HotkeyManager{
		callbacks:  make(map[string]HotkeyCallback),
		registered: make(map[string]*hotkey.Hotkey),
		isRunning:  false,
	}
}

// Register registers a global hotkey with a callback
func (hm *HotkeyManager) Register(hotkeyStr string, callback HotkeyCallback) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, exists := hm.registered[hotkeyStr]; exists {
		return fmt.Errorf("hotkey %s already registered", hotkeyStr)
	}

	mods, key, err := parseHotkey(hotkeyStr)
	if err != nil {
		return err
	}

	hk := hotkey.New(mods, key)
	err = hk.Register()
	if err != nil {
		return fmt.Errorf("failed to register hotkey %s: %w", hotkeyStr, err)
	}

	hm.registered[hotkeyStr] = hk
	hm.callbacks[hotkeyStr] = callback

	go func() {
		for range hk.Keydown() {
			log.Printf("Global hotkey triggered: %s", hotkeyStr)
			if cb, ok := hm.callbacks[hotkeyStr]; ok {
				go cb()
			}
		}
	}()

	log.Printf("Registered global hotkey: %s", hotkeyStr)
	return nil
}

// parseHotkey converts a string like "Cmd+Shift+C" into hotkey library types
func parseHotkey(hotkeyStr string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(hotkeyStr, "+")
	if len(parts) == 0 {
		return nil, 0, fmt.Errorf("invalid hotkey string: %s", hotkeyStr)
	}

	keyStr := parts[len(parts)-1]
	modStrs := parts[:len(parts)-1]

	var mods []hotkey.Modifier
	for _, modStr := range modStrs {
		switch strings.ToLower(modStr) {
		case "cmd", "command", "super":
			if runtime.GOOS == "darwin" {
				mods = append(mods, hotkey.ModCmd)
			} else {
				mods = append(mods, hotkey.ModCtrl) // Use Ctrl on non-macOS
			}
		case "shift":
			mods = append(mods, hotkey.ModShift)
		case "ctrl", "control":
			mods = append(mods, hotkey.ModCtrl)
		case "alt", "option":
			mods = append(mods, hotkey.ModOption)
		default:
			return nil, 0, fmt.Errorf("unknown modifier: %s", modStr)
		}
	}

	key, ok := keyMap[strings.ToUpper(keyStr)]
	if !ok {
		return nil, 0, fmt.Errorf("unknown key: %s", keyStr)
	}

	return mods, key, nil
}

// Unregister removes a hotkey registration
func (hm *HotkeyManager) Unregister(hotkeyStr string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hk, exists := hm.registered[hotkeyStr]; exists {
		if err := hk.Unregister(); err != nil {
			log.Printf("Failed to unregister hotkey %s: %v", hotkeyStr, err)
		}
		delete(hm.registered, hotkeyStr)
		delete(hm.callbacks, hotkeyStr)
		log.Printf("Unregistered hotkey: %s", hotkeyStr)
	}
}

// Start is a placeholder, as registration happens immediately
func (hm *HotkeyManager) Start() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.isRunning = true
	log.Println("Global hotkey manager started")
	return nil
}

// Stop stops the hotkey manager by unregistering all hotkeys
func (hm *HotkeyManager) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.isRunning {
		return
	}

	for str, hk := range hm.registered {
		if err := hk.Unregister(); err != nil {
			log.Printf("Failed to unregister hotkey %s: %v", str, err)
		}
		log.Printf("Unregistered hotkey on stop: %s", str)
	}

	hm.registered = make(map[string]*hotkey.Hotkey)
	hm.callbacks = make(map[string]HotkeyCallback)
	hm.isRunning = false
	log.Println("Hotkey manager stopped")
}

// IsRunning returns whether the hotkey manager is currently running
func (hm *HotkeyManager) IsRunning() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.isRunning
}

// A map to convert string representations of keys to hotkey.Key constants
var keyMap = map[string]hotkey.Key{
	"A":     hotkey.KeyA,
	"B":     hotkey.KeyB,
	"C":     hotkey.KeyC,
	"D":     hotkey.KeyD,
	"E":     hotkey.KeyE,
	"F":     hotkey.KeyF,
	"G":     hotkey.KeyG,
	"H":     hotkey.KeyH,
	"I":     hotkey.KeyI,
	"J":     hotkey.KeyJ,
	"K":     hotkey.KeyK,
	"L":     hotkey.KeyL,
	"M":     hotkey.KeyM,
	"N":     hotkey.KeyN,
	"O":     hotkey.KeyO,
	"P":     hotkey.KeyP,
	"Q":     hotkey.KeyQ,
	"R":     hotkey.KeyR,
	"S":     hotkey.KeyS,
	"T":     hotkey.KeyT,
	"U":     hotkey.KeyU,
	"V":     hotkey.KeyV,
	"W":     hotkey.KeyW,
	"X":     hotkey.KeyX,
	"Y":     hotkey.KeyY,
	"Z":     hotkey.KeyZ,
	"SPACE": hotkey.KeySpace,
	// ",":      hotkey.KeyComma,
	// ".":      hotkey.KeyPeriod,
	// "/":      hotkey.KeySlash,
	"DELETE": hotkey.KeyDelete,
	"RETURN": hotkey.KeyReturn,
	"ESCAPE": hotkey.KeyEscape,
	"TAB":    hotkey.KeyTab,
}
