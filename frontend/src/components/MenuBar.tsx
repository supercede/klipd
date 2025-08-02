import React, { useState } from "react";

interface ClipboardItem {
  id: string;
  contentType: "text" | "image" | "file";
  preview: string;
  createdAt: Date;
  isPinned: boolean;
}

interface MenuBarProps {
  recentItems: ClipboardItem[];
  onShowAll: () => void;
  onPreferences: () => void;
  onPauseMonitoring: () => void;
  onQuit: () => void;
  isMonitoringPaused: boolean;
}

const MenuBar: React.FC<MenuBarProps> = ({
  recentItems,
  onShowAll,
  onPreferences,
  onPauseMonitoring,
  onQuit,
  isMonitoringPaused,
}) => {
  const [isOpen, setIsOpen] = useState(false);

  const getContentIcon = (contentType: string) => {
    switch (contentType) {
      case "text":
        return "📄";
      case "image":
        return "🖼️";
      case "file":
        return "📁";
      default:
        return "📄";
    }
  };

  const formatPreview = (preview: string, maxLength: number = 40) => {
    return preview.length > maxLength
      ? preview.substring(0, maxLength) + "..."
      : preview;
  };

  const formatTime = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return `${days}d ago`;
  };

  return (
    <div className="relative">
      {/* Menu Bar Icon */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-6 h-6 flex items-center justify-center text-macos-text-primary dark:text-macos-dark-text-primary hover:bg-macos-bg-tertiary dark:hover:bg-macos-dark-bg-tertiary rounded transition-colors duration-150"
        aria-label="Klipd Menu"
      >
        📋
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />

          {/* Menu Content */}
          <div className="absolute right-0 top-8 z-20 w-80 bg-macos-bg-primary dark:bg-macos-dark-bg-primary backdrop-blur-macos rounded-macos shadow-macos dark:shadow-macos-dark border border-macos-border dark:border-macos-dark-border animate-scale-in">
            {/* Recent Items */}
            <div className="p-2">
              {recentItems.length > 0 ? (
                <>
                  <div className="text-xs font-medium text-macos-text-secondary dark:text-macos-dark-text-secondary uppercase tracking-wide px-3 py-2">
                    Recent
                  </div>
                  {recentItems.slice(0, 5).map((item, index) => (
                    <button
                      key={item.id}
                      className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-blue hover:text-white rounded-macos-input transition-colors duration-150 group"
                      onClick={() => {
                        // TODO: Handle item selection
                        setIsOpen(false);
                      }}
                    >
                      <span className="text-base mr-3 flex-shrink-0">
                        {getContentIcon(item.contentType)}
                        {item.isPinned && <span className="text-xs">📌</span>}
                      </span>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-medium truncate">
                          {formatPreview(item.preview)}
                        </div>
                        <div className="text-xs text-macos-text-tertiary dark:text-macos-dark-text-tertiary group-hover:text-white/70">
                          {formatTime(item.createdAt)}
                        </div>
                      </div>
                      <div className="text-xs text-macos-text-tertiary dark:text-macos-dark-text-tertiary group-hover:text-white/70 ml-2">
                        ⌘{index + 1}
                      </div>
                    </button>
                  ))}
                </>
              ) : (
                <div className="px-3 py-8 text-center">
                  <div className="text-4xl mb-2">📋</div>
                  <div className="text-sm text-macos-text-secondary dark:text-macos-dark-text-secondary">
                    No clipboard history
                  </div>
                  <div className="text-xs text-macos-text-tertiary dark:text-macos-dark-text-tertiary mt-1">
                    Copy something to get started
                  </div>
                </div>
              )}
            </div>

            {/* Separator */}
            <div className="h-px bg-macos-border dark:bg-macos-dark-border mx-2" />

            {/* Actions */}
            <div className="p-2">
              <button
                onClick={() => {
                  onShowAll();
                  setIsOpen(false);
                }}
                className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-blue hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
              >
                <span className="mr-3">⌘</span>
                Show All History
              </button>

              <button
                onClick={() => {
                  onPreferences();
                  setIsOpen(false);
                }}
                className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-blue hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
              >
                <span className="mr-3">⚙️</span>
                Preferences...
              </button>
            </div>

            {/* Separator */}
            <div className="h-px bg-macos-border dark:bg-macos-dark-border mx-2" />

            {/* System Actions */}
            <div className="p-2">
              <button
                onClick={() => {
                  onPauseMonitoring();
                  setIsOpen(false);
                }}
                className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-blue hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
              >
                <span className="mr-3">{isMonitoringPaused ? "▶️" : "⏸️"}</span>
                {isMonitoringPaused ? "Resume Monitoring" : "Pause Monitoring"}
              </button>

              <button
                onClick={() => {
                  onQuit();
                  setIsOpen(false);
                }}
                className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-red hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
              >
                <span className="mr-3">🚪</span>
                Quit Klipd
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default MenuBar;
