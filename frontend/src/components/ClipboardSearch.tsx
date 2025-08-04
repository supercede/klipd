import React, { useState, useEffect, useRef } from "react";
import { Check } from "./check";

interface ClipboardItem {
  id: string;
  contentType: "text" | "image" | "file";
  content: string;
  preview: string;
  createdAt: Date;
  isPinned: boolean;
  lastAccessed: Date;
}

interface ClipboardSearchProps {
  items: ClipboardItem[];
  onItemSelect: (item: ClipboardItem) => void;
  onItemDelete: (id: string) => void;
  onItemPin: (id: string, pinned: boolean) => void;
  onClose: () => void;
  onSearch?: (
    query: string,
    useRegex?: boolean,
    limit?: number,
    offset?: number
  ) => Promise<ClipboardItem[]>;
  onLoadMore?: (limit: number, offset: number) => Promise<ClipboardItem[]>;
  isVisible: boolean;
  sortByRecent?: "copied" | "pasted";
}

const ClipboardSearch: React.FC<ClipboardSearchProps> = ({
  items,
  onItemSelect,
  onItemDelete,
  onItemPin,
  onClose,
  onSearch,
  onLoadMore,
  isVisible,
  sortByRecent = "copied",
}) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [searchResults, setSearchResults] = useState<ClipboardItem[]>(items);
  const [isSearching, setIsSearching] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMoreItems, setHasMoreItems] = useState(true);
  const [useRegex, setUseRegex] = useState(false);
  const [copiedItemId, setCopiedItemId] = useState<string | null>(null);
  const [contextMenu, setContextMenu] = useState<{
    x: number;
    y: number;
    itemId: string;
  } | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const itemsContainerRef = useRef<HTMLDivElement>(null);

  // Update search results when items change or when there's no search query
  useEffect(() => {
    if (!searchQuery.trim()) {
      setSearchResults(items);
    }
  }, [items, searchQuery]);

  // Handle search with debouncing
  useEffect(() => {
    const timeoutId = setTimeout(async () => {
      if (searchQuery.trim() && onSearch) {
        setIsSearching(true);
        try {
          const results = await onSearch(searchQuery, useRegex, 15, 0);
          setSearchResults(results);
          setHasMoreItems(results.length === 15); // If we got 15 items, there might be more
        } catch (error) {
          setSearchResults([]);
          setHasMoreItems(false);
        } finally {
          setIsSearching(false);
        }
      } else if (!searchQuery.trim()) {
        // For initial view, load first 15 items using onLoadMore if available
        if (onLoadMore) {
          setIsSearching(true);
          try {
            const results = await onLoadMore(15, 0);
            setSearchResults(results);
            setHasMoreItems(results.length === 15);
          } catch (error) {
            setSearchResults(items.slice(0, 15));
            setHasMoreItems(items.length > 15);
          } finally {
            setIsSearching(false);
          }
        } else {
          setSearchResults(items.slice(0, 15));
          setHasMoreItems(items.length > 15);
        }
      }
    }, 300); // 300ms debounce

    return () => clearTimeout(timeoutId);
  }, [searchQuery, useRegex, onSearch, items]);

  const loadMoreItems = async () => {
    if (!hasMoreItems || isLoadingMore || !onLoadMore) return;

    setIsLoadingMore(true);
    try {
      let moreItems: ClipboardItem[];
      if (searchQuery.trim()) {
        if (onSearch) {
          moreItems = await onSearch(
            searchQuery,
            useRegex,
            15,
            searchResults.length
          );
        } else {
          moreItems = [];
        }
      } else {
        moreItems = await onLoadMore(15, searchResults.length);
      }

      if (moreItems.length > 0) {
        setSearchResults((prev) => [...prev, ...moreItems]);
        setHasMoreItems(moreItems.length === 15);
      } else {
        setHasMoreItems(false);
      }
    } catch (error) {
      console.error("Failed to load more items:", error);
      setHasMoreItems(false);
    } finally {
      setIsLoadingMore(false);
    }
  };

  const handleLoadMoreClick = () => {
    loadMoreItems();
  };

  useEffect(() => {
    if (isVisible && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [isVisible]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [searchQuery]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isVisible) return;

      switch (e.key) {
        case "Escape":
          onClose();
          break;
        case "ArrowDown":
          e.preventDefault();
          setSelectedIndex((prev) =>
            Math.min(prev + 1, searchResults.length - 1)
          );
          break;
        case "ArrowUp":
          e.preventDefault();
          setSelectedIndex((prev) => Math.max(prev - 1, 0));
          break;
        case "Enter":
          e.preventDefault();
          if (searchResults[selectedIndex]) {
            const selectedItem = searchResults[selectedIndex];
            setCopiedItemId(selectedItem.id);

            // Show copy animation for 850ms before closing (matches SVG animation timing)
            setTimeout(() => {
              onItemSelect(selectedItem);
              onClose();
              setCopiedItemId(null);
            }, 850);
          }
          break;
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isVisible, selectedIndex, searchResults, onItemSelect, onClose]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement;

      // If clicking on context menu, don't close anything
      if (target.closest("[data-context-menu]")) {
        return;
      }

      // If context menu is open and clicking outside of it, close context menu
      if (contextMenu && !target.closest("[data-context-menu]")) {
        setContextMenu(null);
        return;
      }

      // If clicking outside the search interface, close the whole component
      if (!target.closest("[data-search-interface]")) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("mousedown", handleClickOutside);
      return () =>
        document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [isVisible, onClose, contextMenu]);

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

  const formatPreview = (preview: string, maxLength: number = 60) => {
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

  const handleContextMenu = (e: React.MouseEvent, itemId: string) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      itemId,
    });
  };

  const handleItemClick = (item: ClipboardItem, index: number) => {
    setSelectedIndex(index);
    setCopiedItemId(item.id);

    // Show copy animation for 600ms before closing
    setTimeout(() => {
      onItemSelect(item);
      onClose();
      setCopiedItemId(null);
    }, 800);
  };

  const handlePinClick = (
    e: React.MouseEvent,
    itemId: string,
    isPinned: boolean
  ) => {
    e.preventDefault();
    e.stopPropagation();

    onItemPin(itemId, !isPinned);
    setContextMenu(null);
  };

  const handleDeleteClick = (e: React.MouseEvent, itemId: string) => {
    e.preventDefault();
    e.stopPropagation();
    console.log(`Context menu delete clicked for item:`, itemId);
    onItemDelete(itemId);
    setContextMenu(null);
  };

  if (!isVisible) return null;

  return (
    <>
      <div className="fixed inset-0 bg-black/20 backdrop-blur-sm z-40 transition-opacity duration-200" />

      <div
        data-search-interface
        className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 w-[600px] max-h-[450px] bg-white dark:bg-gray-800 backdrop-blur rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 transition-all duration-700 ease-out scale-100 opacity-100"
      >
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="relative">
            <span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400">
              {isSearching ? "⏳" : "🔍"}
            </span>
            <input
              ref={searchInputRef}
              type="text"
              placeholder={
                useRegex
                  ? "Search with regex patterns..."
                  : "Search clipboard history..."
              }
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-20 py-3 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <button
              onClick={() => setUseRegex(!useRegex)}
              className={`absolute right-2 top-1/2 transform -translate-y-1/2 px-2 py-1 text-xs font-medium rounded transition-colors duration-150 ${
                useRegex
                  ? "bg-blue-500 text-white"
                  : "bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-500"
              }`}
              title={useRegex ? "Disable regex search" : "Enable regex search"}
            >
              .*
            </button>
          </div>
        </div>

        {/* Results */}
        <div ref={itemsContainerRef} className="max-h-80 overflow-y-auto">
          {searchResults.length > 0 ? (
            <div className="p-2">
              {searchResults.map((item, index) => (
                <div
                  key={item.id}
                  className={`group flex items-center px-3 py-3 cursor-pointer rounded-lg transition-all duration-150 ${
                    index === selectedIndex
                      ? "bg-blue-500 text-white scale-[1.02]"
                      : "hover:bg-gray-100 dark:hover:bg-gray-700"
                  }`}
                  onClick={() => handleItemClick(item, index)}
                  onContextMenu={(e) => handleContextMenu(e, item.id)}
                  onMouseEnter={() => setSelectedIndex(index)}
                >
                  {/* Content Icon */}
                  <div className="flex items-center mr-3 flex-shrink-0">
                    <span className="text-lg">
                      {getContentIcon(item.contentType)}
                    </span>
                    {item.isPinned && (
                      <span className="text-xs ml-1 text-blue-500 group-hover:text-white">
                        📌
                      </span>
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium truncate">
                      {formatPreview(item.preview)}
                    </div>
                    <div
                      className={`text-xs mt-1 ${
                        index === selectedIndex
                          ? "text-white/70"
                          : "text-gray-500 dark:text-gray-400"
                      }`}
                    >
                      {formatTime(item.createdAt)}
                    </div>
                  </div>

                  {/* Copy Success Checkmark */}
                  {copiedItemId === item.id && (
                    <div className="ml-3 flex-shrink-0">
                      <Check />
                    </div>
                  )}
                </div>
              ))}

              {/* Tiny Load More Button */}
              {hasMoreItems && !isLoadingMore && (
                <div className="flex justify-center py-2">
                  <button
                    onClick={handleLoadMoreClick}
                    className="px-3 py-1 text-xs bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-600 dark:text-gray-300 rounded-md transition-colors duration-150 border border-gray-200 dark:border-gray-600"
                  >
                    Load More
                  </button>
                </div>
              )}

              {isLoadingMore && (
                <div className="flex justify-center py-4">
                  <div className="flex items-center space-x-2">
                    <div className="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                    <span className="text-gray-400">...</span>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="p-8 text-center">
              <div className="text-4xl mb-3">🔍</div>
              <div className="text-lg font-medium text-gray-900 dark:text-white mb-1">
                No matching results
              </div>
              <div className="text-sm text-gray-500 dark:text-gray-400">
                Try a different search term
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <div
          data-context-menu
          className="fixed z-50 w-48 bg-macos-bg-primary dark:bg-macos-dark-bg-primary backdrop-blur-macos rounded-macos shadow-macos dark:shadow-macos-dark border border-macos-border dark:border-macos-dark-border p-1"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          {(() => {
            const item = searchResults.find((i) => i.id === contextMenu.itemId);
            return item ? (
              <>
                <button
                  onClick={(e) =>
                    handlePinClick(e, contextMenu.itemId, item.isPinned)
                  }
                  className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-blue hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
                >
                  <span className="mr-3">{item.isPinned ? "📌" : "📍"}</span>
                  {item.isPinned ? "Unpin" : "Pin"}
                </button>

                <button
                  onClick={(e) => handleDeleteClick(e, contextMenu.itemId)}
                  className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-red hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
                >
                  <span className="mr-3">🗑️</span>
                  Delete
                </button>
              </>
            ) : null;
          })()}
        </div>
      )}
    </>
  );
};

export default ClipboardSearch;
