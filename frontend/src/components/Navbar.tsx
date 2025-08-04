import React, { useState, useEffect } from "react";
import * as WailsApp from "../../wailsjs/go/main/App";
import { models } from "../../wailsjs/go/models";

interface NavbarProps {
  isMonitoringPaused: boolean;
  totalItemCount: number;
  onPauseMonitoring: () => void;
  onShowSearch: () => void;
  onShowSettings: () => void;
}

interface ClipboardItem {
  id: string;
  contentType: "text" | "image" | "file";
  content: string;
  preview: string;
  createdAt: Date;
  isPinned: boolean;
  lastAccessed: Date;
}

export const Navbar: React.FC<NavbarProps> = ({
  isMonitoringPaused,
  totalItemCount,
  onPauseMonitoring,
  onShowSearch,
  onShowSettings,
}) => {
  const [isRecentItemsVisible, setIsRecentItemsVisible] = useState(false);
  const [recentItems, setRecentItems] = useState<ClipboardItem[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (isRecentItemsVisible) {
      loadRecentItems();
    }
  }, [isRecentItemsVisible]);

  const loadRecentItems = async () => {
    try {
      setLoading(true);
      const items = await WailsApp.GetRecentItems(8);
      const convertedItems: ClipboardItem[] = items.map((item) => ({
        id: item.id,
        contentType: item.contentType as "text" | "image" | "file",
        content: item.content,
        preview: item.preview,
        createdAt: new Date(item.createdAt),
        isPinned: item.isPinned,
        lastAccessed: new Date(item.lastAccessed),
      }));
      setRecentItems(convertedItems);
    } catch (error) {
      console.error("Failed to load recent items:", error);
      setRecentItems([]);
    } finally {
      setLoading(false);
    }
  };

  const handleItemSelect = async (item: ClipboardItem) => {
    try {
      await WailsApp.SelectClipboardItem(item.id);
      console.log("Selected item:", item.content);
      setIsRecentItemsVisible(false);
    } catch (error) {
      console.error("Failed to select item:", error);
    }
  };

  const handleRecentItemsToggle = () => {
    setIsRecentItemsVisible(!isRecentItemsVisible);
  };

  const formatPreview = (content: string, maxLength: number = 60) => {
    if (content.length <= maxLength) return content;
    return content.substring(0, maxLength) + "...";
  };

  const formatDate = (date: Date) => {
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return "Just now";
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="fixed top-0 left-0 right-0 h-12 bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200/30 dark:border-gray-700/30 z-40">
      <div className="flex items-center justify-between h-full px-4">
        {/* Left Side - App Info */}
        <div className="flex items-center space-x-3">
          <div className="text-lg">📋</div>
          <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
            Klipd
          </span>
          <div className="flex items-center space-x-1 text-xs text-gray-500 dark:text-gray-400">
            <div
              className={`w-1.5 h-1.5 rounded-full ${
                isMonitoringPaused ? "bg-red-500" : "bg-green-500"
              }`}
            />
            <span>{isMonitoringPaused ? "Paused" : "Active"}</span>
          </div>
        </div>

        {/* Center - Quick Actions */}
        <div className="flex items-center space-x-2">
          {/* Item Count */}
          <div className="flex items-center space-x-1 px-2 py-1 text-xs text-gray-600 dark:text-gray-400">
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14,2 14,8 20,8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
              <line x1="10" y1="9" x2="8" y2="9" />
            </svg>
            <span>{totalItemCount} items</span>
          </div>

          {/* Recent Items Dropdown */}
          <div className="relative">
            <button
              onClick={handleRecentItemsToggle}
              className="flex items-center space-x-1 px-2 py-1 text-xs text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
              title="Recent Items (⌘⇧M)"
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <rect x="8" y="2" width="8" height="4" rx="1" ry="1" />
                <path d="m16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
              </svg>
              <span>Recent</span>
              <svg
                width="10"
                height="10"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <polyline points="6,9 12,15 18,9" />
              </svg>
            </button>

            {/* Recent Items Dropdown */}
            {isRecentItemsVisible && (
              <>
                {/* Backdrop */}
                <div
                  className="fixed inset-0 z-40"
                  onClick={() => setIsRecentItemsVisible(false)}
                />

                {/* Dropdown Content */}
                <div className="absolute top-8 right-0 w-80 bg-white/95 dark:bg-gray-800/95 backdrop-blur-md rounded-lg shadow-xl border border-gray-200/50 dark:border-gray-700/50 z-50 overflow-hidden">
                  <div className="p-3 border-b border-gray-200/50 dark:border-gray-700/50">
                    <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      Recent Items
                    </h3>
                  </div>

                  <div className="max-h-96 overflow-y-auto">
                    {loading ? (
                      <div className="p-4 text-center">
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          Loading...
                        </div>
                      </div>
                    ) : recentItems.length > 0 ? (
                      <div className="py-2">
                        {recentItems.map((item, index) => (
                          <div
                            key={item.id}
                            onClick={() => handleItemSelect(item)}
                            className="px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700/50 cursor-pointer transition-colors border-b border-gray-100 dark:border-gray-700/30 last:border-b-0"
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex-1 min-w-0">
                                <div className="text-xs text-gray-900 dark:text-gray-100 font-medium truncate">
                                  {formatPreview(item.preview || item.content)}
                                </div>
                                <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                  {formatDate(item.createdAt)}
                                </div>
                              </div>
                              <div className="ml-2 flex-shrink-0">
                                <span className="text-xs text-gray-400 dark:text-gray-500">
                                  {index + 1}
                                </span>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="p-4 text-center">
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          No recent items
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="p-2 border-t border-gray-200/50 dark:border-gray-700/50 bg-gray-50/50 dark:bg-gray-800/50">
                    <button
                      onClick={() => {
                        onShowSearch();
                        setIsRecentItemsVisible(false);
                      }}
                      className="w-full text-xs text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 py-1"
                    >
                      Show All Items
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>

        {/* Right Side - Main Actions */}
        <div className="flex items-center space-x-2">
          {/* Toggle Monitoring */}
          <button
            onClick={onPauseMonitoring}
            className={`p-2 rounded-lg transition-colors ${
              isMonitoringPaused
                ? "text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                : "text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20"
            }`}
            title={
              isMonitoringPaused ? "Resume Monitoring" : "Pause Monitoring"
            }
          >
            {isMonitoringPaused ? (
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <polygon points="5,3 19,12 5,21" />
              </svg>
            ) : (
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <rect x="6" y="4" width="4" height="16" />
                <rect x="14" y="4" width="4" height="16" />
              </svg>
            )}
          </button>

          {/* Search History */}
          <button
            onClick={onShowSearch}
            className="p-2 text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
            title="Search History (⌘⇧Space)"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <circle cx="11" cy="11" r="8" />
              <path d="m21 21-4.35-4.35" />
            </svg>
          </button>

          {/* Settings */}
          <button
            onClick={onShowSettings}
            className="p-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            title="Preferences (⌘,)"
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <circle cx="12" cy="12" r="3" />
              <path d="M12 1v6m0 6v6m6-6h-6m0 0H6" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Navbar;
