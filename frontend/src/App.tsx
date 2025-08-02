import React, { useState, useEffect } from "react";
import MenuBar from "./components/MenuBar";
import ClipboardSearch from "./components/ClipboardSearch";
import Settings from "./components/Settings";
import "./style.css";

interface ClipboardItem {
  id: string;
  contentType: "text" | "image" | "file";
  content: string;
  preview: string;
  createdAt: Date;
  isPinned: boolean;
  lastAccessed: Date;
}

interface SettingsData {
  globalHotkey: string;
  previousItemHotkey: string;
  pollingInterval: number;
  maxItems: number;
  maxDays: number;
  autoLaunch: boolean;
  enableSounds: boolean;
}

function App() {
  // TODO: useContext?
  const [clipboardItems, setClipboardItems] = useState<ClipboardItem[]>([]);
  const [isSearchVisible, setIsSearchVisible] = useState(false);
  const [isSettingsVisible, setIsSettingsVisible] = useState(false);
  const [isMonitoringPaused, setIsMonitoringPaused] = useState(false);
  const [settings, setSettings] = useState<SettingsData>({
    globalHotkey: "⌘⇧V",
    previousItemHotkey: "⌘⇧C",
    pollingInterval: 500,
    maxItems: 100,
    maxDays: 7,
    autoLaunch: true,
    enableSounds: false,
  });

  // Mock data for demonstration
  useEffect(() => {
    const mockItems: ClipboardItem[] = [
      {
        id: "1",
        contentType: "text",
        content:
          "Hello world from the clipboard! This is a longer piece of text to demonstrate how the preview works.",
        preview:
          "Hello world from the clipboard! This is a longer piece of text to demonstrate how the preview works.",
        createdAt: new Date(Date.now() - 300000), // 5 minutes ago
        isPinned: true,
        lastAccessed: new Date(Date.now() - 300000),
      },
      {
        id: "2",
        contentType: "image",
        content: "/path/to/screenshot.png",
        preview: "Screenshot_2024_08_02_15_30_45.png",
        createdAt: new Date(Date.now() - 600000), // 10 minutes ago
        isPinned: false,
        lastAccessed: new Date(Date.now() - 600000),
      },
      {
        id: "3",
        contentType: "text",
        content:
          'console.log("Debugging the clipboard manager functionality");',
        preview:
          'console.log("Debugging the clipboard manager functionality");',
        createdAt: new Date(Date.now() - 900000), // 15 minutes ago
        isPinned: false,
        lastAccessed: new Date(Date.now() - 900000),
      },
      {
        id: "4",
        contentType: "file",
        content: "/Users/ade/Documents/project-files.zip",
        preview: "~/Documents/project-files.zip",
        createdAt: new Date(Date.now() - 1800000), // 30 minutes ago
        isPinned: false,
        lastAccessed: new Date(Date.now() - 1800000),
      },
      {
        id: "5",
        contentType: "text",
        content: "https://github.com/wailsapp/wails",
        preview: "https://github.com/wailsapp/wails",
        createdAt: new Date(Date.now() - 3600000), // 1 hour ago
        isPinned: false,
        lastAccessed: new Date(Date.now() - 3600000),
      },
    ];
    setClipboardItems(mockItems);
  }, []);

  // Global hotkey handling
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Check for global hotkey (Cmd+Shift+V)
      if (e.metaKey && e.shiftKey && e.key === "V") {
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
  }, [clipboardItems]);

  // Handlers
  const handleItemSelect = (item: ClipboardItem) => {
    console.log("Selected item:", item);
    // Update last accessed time
    setClipboardItems((prev) =>
      prev.map((i) =>
        i.id === item.id ? { ...i, lastAccessed: new Date() } : i
      )
    );
    // TODO: Copy to clipboard via Wails binding
  };

  const handleItemDelete = (itemId: string) => {
    setClipboardItems((prev) => prev.filter((item) => item.id !== itemId));
  };

  const handleItemPin = (itemId: string, pinned: boolean) => {
    setClipboardItems((prev) =>
      prev.map((item) =>
        item.id === itemId ? { ...item, isPinned: pinned } : item
      )
    );
  };

  const handleShowAll = () => {
    setIsSearchVisible(true);
  };

  const handlePreferences = () => {
    setIsSettingsVisible(true);
  };

  const handlePauseMonitoring = () => {
    setIsMonitoringPaused((prev) => !prev);
    // TODO: Implement actual monitoring pause via Wails binding
  };

  const handleQuit = () => {
    // TODO: Implement quit via Wails binding
    console.log("Quit requested");
  };

  const handleSettingsChange = (newSettings: SettingsData) => {
    setSettings(newSettings);
    // TODO: Save settings via Wails binding
  };

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
