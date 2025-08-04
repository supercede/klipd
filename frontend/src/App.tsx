import React, { useState, useEffect } from "react";
import MenuBar from "./components/MenuBar";
import ClipboardSearch from "./components/ClipboardSearch";
import Settings from "./components/Settings";
import "./style.css";

// Import Wails bindings
import * as WailsApp from "../wailsjs/go/main/App";
import { models } from "../wailsjs/go/models";
import { EventsOn, EventsOff } from "../wailsjs/runtime/runtime";

interface ClipboardItem {
  id: string;
  contentType: "text" | "image" | "file";
  content: string;
  preview: string;
  createdAt: Date;
  isPinned: boolean;
  lastAccessed: Date;
}

function App() {
  // State management
  const [clipboardItems, setClipboardItems] = useState<ClipboardItem[]>([]);
  const [isSearchVisible, setIsSearchVisible] = useState(false);
  const [isSettingsVisible, setIsSettingsVisible] = useState(false);
  const [isMonitoringPaused, setIsMonitoringPaused] = useState(false);
  const [settings, setSettings] = useState<models.Settings | null>(null);

  // Load initial data on component mount
  useEffect(() => {
    loadSettings();
    loadClipboardItems();
    checkMonitoringStatus();

    // Listen for backend events
    const showSearchUnsubscribe = EventsOn("show-search-interface", () => {
      setIsSearchVisible(true);
    });

    const hideSearchUnsubscribe = EventsOn("hide-search-interface", () => {
      setIsSearchVisible(false);
    });

    // Listen for clipboard updates
    const clipboardAddedUnsubscribe = EventsOn(
      "clipboard-item-added",
      (_newItem: any) => {
        loadClipboardItems();
      }
    );

    const clipboardUpdatedUnsubscribe = EventsOn(
      "clipboard-item-updated",
      (_updatedItem: any) => {
        loadClipboardItems();
      }
    );

    return () => {
      showSearchUnsubscribe();
      hideSearchUnsubscribe();
      clipboardAddedUnsubscribe();
      clipboardUpdatedUnsubscribe();
    };
  }, []);

  const loadSettings = async () => {
    try {
      const loadedSettings = await WailsApp.GetSettings();
      setSettings(loadedSettings);
    } catch (error) {
      console.error("Failed to load settings:", error);
      // Fallback to default settings
      const defaultSettings = new models.Settings({
        id: 0,
        globalHotkey: "Cmd+Shift+Space",
        previousItemHotkey: "Cmd+Shift+C",
        pollingInterval: 500,
        maxItems: 100,
        maxDays: 7,
        autoLaunch: true,
        enableSounds: false,
        monitoringEnabled: true,
        allowPasswords: false,
        createdAt: new Date(),
        updatedAt: new Date(),
      });
      setSettings(defaultSettings);
    }
  };

  const loadClipboardItems = async () => {
    try {
      const items = await WailsApp.GetClipboardItems(50, 0, "");
      const convertedItems: ClipboardItem[] = items.map((item) => ({
        id: item.id,
        contentType: item.contentType as "text" | "image" | "file",
        content: item.content,
        preview: item.preview,
        createdAt: new Date(item.createdAt),
        isPinned: item.isPinned,
        lastAccessed: new Date(item.lastAccessed),
      }));
      setClipboardItems(convertedItems);
    } catch (error) {
      console.error("Failed to load clipboard items:", error);
    }
  };

  const checkMonitoringStatus = async () => {
    try {
      const isEnabled = await WailsApp.IsMonitoringEnabled();
      setIsMonitoringPaused(!isEnabled);
    } catch (error) {
      console.error("Failed to check monitoring status:", error);
    }
  };

  // Global hotkey handling
  useEffect(() => {
    if (!settings) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // Check for global hotkey (default: Cmd+Shift+Space)
      if (e.metaKey && e.shiftKey && e.code === "Space") {
        e.preventDefault();
        setIsSearchVisible(true);
      }

      // Check for previous item hotkey (Cmd+Shift+C)
      if (e.metaKey && e.shiftKey && e.key === "C") {
        e.preventDefault();
        // Select the most recent item (excluding pinned items for this demo)
        const recentItem = clipboardItems
          .filter((item) => !item.isPinned)
          .sort(
            (a, b) =>
              new Date(b.lastAccessed).getTime() -
              new Date(a.lastAccessed).getTime()
          )[0];

        if (recentItem) {
          handleItemSelect(recentItem);
        }
      }

      // Settings hotkey (Cmd+,)
      if (e.metaKey && e.key === ",") {
        e.preventDefault();
        setIsSettingsVisible(true);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [clipboardItems, settings]);

  // Handlers
  const handleItemSelect = async (item: ClipboardItem) => {
    console.log("Selected item:", item);
    try {
      // Copy to clipboard via Wails binding
      await WailsApp.SelectClipboardItem(item.id);

      // Update last accessed time locally
      setClipboardItems((prev) =>
        prev.map((i) =>
          i.id === item.id ? { ...i, lastAccessed: new Date() } : i
        )
      );

      // Close search interface
      setIsSearchVisible(false);
    } catch (error) {
      console.error("Failed to select clipboard item:", error);
    }
  };

  const handleItemDelete = async (itemId: string) => {
    try {
      await WailsApp.DeleteClipboardItem(itemId);
      setClipboardItems((prev) => prev.filter((item) => item.id !== itemId));
    } catch (error) {
      console.error("Failed to delete clipboard item:", error);
    }
  };

  const handleItemPin = async (itemId: string, pinned: boolean) => {
    try {
      await WailsApp.PinClipboardItem(itemId, pinned);
      setClipboardItems((prev) =>
        prev.map((item) =>
          item.id === itemId ? { ...item, isPinned: pinned } : item
        )
      );
    } catch (error) {
      console.error("Failed to pin/unpin clipboard item:", error);
    }
  };

  const handleShowAll = () => {
    setIsSearchVisible(true);
  };

  const handlePreferences = () => {
    setIsSettingsVisible(true);
  };

  const handlePauseMonitoring = async () => {
    try {
      const newStatus = await WailsApp.ToggleMonitoring();
      setIsMonitoringPaused(!newStatus);
    } catch (error) {
      console.error("Failed to toggle monitoring:", error);
    }
  };

  const handleQuit = () => {
    // TODO: Implement quit via Wails binding
    console.log("Quit requested");
  };

  const handleSettingsChange = async (newSettings: models.Settings) => {
    try {
      await WailsApp.UpdateSettings(newSettings);
      setSettings(newSettings);

      // Reload clipboard items if settings changed
      await loadClipboardItems();
    } catch (error) {
      console.error("Failed to update settings:", error);
    }
  };

  // Handle search with real-time backend integration
  const handleSearch = async (
    query: string,
    useRegex: boolean = false
  ): Promise<ClipboardItem[]> => {
    try {
      let items;
      if (useRegex) {
        items = await WailsApp.SearchClipboardItemsRegex(query, 50);
      } else {
        items = await WailsApp.SearchClipboardItems(query, 50);
      }

      return items.map((item) => ({
        id: item.id,
        contentType: item.contentType as "text" | "image" | "file",
        content: item.content,
        preview: item.preview,
        createdAt: new Date(item.createdAt),
        isPinned: item.isPinned,
        lastAccessed: new Date(item.lastAccessed),
      }));
    } catch (error) {
      console.error("Failed to search clipboard items:", error);
      return [];
    }
  };

  // Don't render until settings are loaded
  if (!settings) {
    return (
      <div className="min-h-screen bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary font-system flex items-center justify-center">
        <div className="text-center">
          <div className="text-8xl mb-6">📋</div>
          <p className="text-lg text-macos-text-secondary dark:text-macos-dark-text-secondary">
            Loading Klipd...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary font-system">
      {/* Menu Bar */}
      <div className="fixed top-4 right-4 z-30">
        <MenuBar
          recentItems={clipboardItems.slice(0, 5)}
          onShowAll={handleShowAll}
          onPreferences={handlePreferences}
          onPauseMonitoring={handlePauseMonitoring}
          onQuit={handleQuit}
          isMonitoringPaused={isMonitoringPaused}
        />
      </div>

      {/* Main Content */}
      <div className="flex items-center justify-center min-h-screen p-8">
        <div className="text-center">
          <div className="text-8xl mb-6">📋</div>
          <h1 className="text-3xl font-bold text-macos-text-primary dark:text-macos-dark-text-primary mb-3">
            Klipd
          </h1>
          <p className="text-lg text-macos-text-secondary dark:text-macos-dark-text-secondary mb-6 max-w-md">
            Your intelligent clipboard manager is running in the background.
          </p>

          <div className="space-y-3 text-sm text-macos-text-tertiary dark:text-macos-dark-text-tertiary">
            <div className="flex items-center justify-center space-x-2">
              <span className="inline-flex items-center px-2 py-1 bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary rounded font-mono text-xs">
                {settings.globalHotkey}
              </span>
              <span>to open clipboard history</span>
            </div>
            <div className="flex items-center justify-center space-x-2">
              <span className="inline-flex items-center px-2 py-1 bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary rounded font-mono text-xs">
                {settings.previousItemHotkey}
              </span>
              <span>to paste previous item</span>
            </div>
            <div className="flex items-center justify-center space-x-2">
              <span className="inline-flex items-center px-2 py-1 bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary rounded font-mono text-xs">
                ⌘,
              </span>
              <span>to open preferences</span>
            </div>
          </div>

          <div className="mt-8 space-x-3">
            <button
              onClick={() => setIsSearchVisible(true)}
              className="px-6 py-2 bg-macos-accent-blue text-white rounded-macos-button hover:bg-blue-500 transition-colors"
            >
              Show History
            </button>
            <button
              onClick={() => setIsSettingsVisible(true)}
              className="px-6 py-2 bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary text-macos-text-primary dark:text-macos-dark-text-primary rounded-macos-button hover:bg-macos-border dark:hover:bg-macos-dark-border transition-colors"
            >
              Preferences
            </button>
          </div>

          {/* Status */}
          <div className="mt-6 flex items-center justify-center space-x-4 text-xs text-macos-text-tertiary dark:text-macos-dark-text-tertiary">
            <div className="flex items-center space-x-1">
              <div
                className={`w-2 h-2 rounded-full ${
                  isMonitoringPaused
                    ? "bg-macos-accent-red"
                    : "bg-macos-accent-green"
                }`}
              />
              <span>
                {isMonitoringPaused ? "Monitoring Paused" : "Monitoring Active"}
              </span>
            </div>
            <span>•</span>
            <span>{clipboardItems.length} items in history</span>
          </div>
        </div>
      </div>

      {/* Search Interface */}
      <ClipboardSearch
        items={clipboardItems}
        onItemSelect={handleItemSelect}
        onItemDelete={handleItemDelete}
        onItemPin={handleItemPin}
        onSearch={handleSearch}
        onClose={() => setIsSearchVisible(false)}
        isVisible={isSearchVisible}
      />

      {/* Settings */}
      <Settings
        settings={settings}
        onSettingsChange={handleSettingsChange}
        onClose={() => setIsSettingsVisible(false)}
        isVisible={isSettingsVisible}
      />
    </div>
  );
}

export default App;
