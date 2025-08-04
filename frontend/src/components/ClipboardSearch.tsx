import React, { useState, useEffect, useRef } from "react";

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
  onSearch?: (query: string, useRegex?: boolean) => Promise<ClipboardItem[]>;
  isVisible: boolean;
}

const ClipboardSearch: React.FC<ClipboardSearchProps> = ({
  items,
  onItemSelect,
  onItemDelete,
  onItemPin,
  onClose,
  onSearch,
  isVisible,
}) => {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [searchResults, setSearchResults] = useState<ClipboardItem[]>(items);
  const [isSearching, setIsSearching] = useState(false);
  const [useRegex, setUseRegex] = useState(false);
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
          const results = await onSearch(searchQuery, useRegex);
          setSearchResults(results);
        } catch (error) {
          console.error("Search failed:", error);
          setSearchResults([]);
        } finally {
          setIsSearching(false);
        }
      } else if (!searchQuery.trim()) {
        setSearchResults(items);
      }
    }, 300); // 300ms debounce

    return () => clearTimeout(timeoutId);
  }, [searchQuery, useRegex, onSearch, items]);

  const sortedItems = [...searchResults].sort((a, b) => {
    if (a.isPinned && !b.isPinned) return -1;
    if (!a.isPinned && b.isPinned) return 1;
    return (
      new Date(b.lastAccessed).getTime() - new Date(a.lastAccessed).getTime()
    );
  });

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
            Math.min(prev + 1, sortedItems.length - 1)
          );
          break;
        case "ArrowUp":
          e.preventDefault();
          setSelectedIndex((prev) => Math.max(prev - 1, 0));
          break;
        case "Enter":
          e.preventDefault();
          if (sortedItems[selectedIndex]) {
            onItemSelect(sortedItems[selectedIndex]);
            onClose();
          }
          break;
        case "1":
        case "2":
        case "3":
        case "4":
        case "5":
        case "6":
        case "7":
        case "8":
        case "9":
          if (e.metaKey || e.ctrlKey) {
            e.preventDefault();
            const index = parseInt(e.key) - 1;
            if (sortedItems[index]) {
              onItemSelect(sortedItems[index]);
              onClose();
            }
          }
          break;
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isVisible, selectedIndex, sortedItems, onItemSelect, onClose]);

  // Click outside to close
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (contextMenu) {
        setContextMenu(null);
        return;
      }

      const target = e.target as HTMLElement;
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
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      itemId,
    });
  };

  const handleItemClick = (item: ClipboardItem, index: number) => {
    setSelectedIndex(index);
    onItemSelect(item);
    onClose();
  };

  if (!isVisible) return null;

  return (
    <>
      <div className="fixed inset-0 bg-black/20 backdrop-blur-sm z-40 animate-slide-in" />

      <div
        data-search-interface
        className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 w-[600px] max-h-[450px] bg-macos-bg-primary dark:bg-macos-dark-bg-primary backdrop-blur-macos rounded-macos shadow-macos dark:shadow-macos-dark border border-macos-border dark:border-macos-dark-border animate-slide-in"
      >
        <div className="p-4 border-b border-macos-border dark:border-macos-dark-border">
          <div className="relative">
            <span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-macos-text-tertiary dark:text-macos-dark-text-tertiary">
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
              className="w-full pl-10 pr-20 py-3 bg-macos-bg-secondary dark:bg-macos-dark-bg-secondary border border-macos-border dark:border-macos-dark-border rounded-macos-input text-macos-text-primary dark:text-macos-dark-text-primary placeholder-macos-text-tertiary dark:placeholder-macos-dark-text-tertiary focus:outline-none focus:ring-2 focus:ring-macos-accent-blue dark:focus:ring-macos-dark-accent-blue focus:border-transparent"
            />
            <button
              onClick={() => setUseRegex(!useRegex)}
              className={`absolute right-2 top-1/2 transform -translate-y-1/2 px-2 py-1 text-xs font-medium rounded transition-colors duration-150 ${
                useRegex
                  ? "bg-macos-accent-blue text-white"
                  : "bg-macos-bg-tertiary dark:bg-macos-dark-bg-tertiary text-macos-text-secondary dark:text-macos-dark-text-secondary hover:bg-macos-bg-quaternary dark:hover:bg-macos-dark-bg-quaternary"
              }`}
              title={useRegex ? "Disable regex search" : "Enable regex search"}
            >
              .*
            </button>
          </div>
        </div>

        {/* Results */}
        <div ref={itemsContainerRef} className="max-h-80 overflow-y-auto">
          {sortedItems.length > 0 ? (
            <div className="p-2">
              {sortedItems.slice(0, 9).map((item, index) => (
                <div
                  key={item.id}
                  className={`group flex items-center px-3 py-3 cursor-pointer rounded-macos-input transition-all duration-150 ${
                    index === selectedIndex
                      ? "bg-macos-accent-blue text-white scale-[1.02]"
                      : "hover:bg-macos-bg-tertiary dark:hover:bg-macos-dark-bg-tertiary"
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
                      <span className="text-xs ml-1 text-macos-accent-blue group-hover:text-white">
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
                          : "text-macos-text-tertiary dark:text-macos-dark-text-tertiary"
                      }`}
                    >
                      {formatTime(item.createdAt)}
                    </div>
                  </div>

                  {/* Keyboard Shortcut */}
                  <div
                    className={`text-xs ml-3 font-medium ${
                      index === selectedIndex
                        ? "text-white/70"
                        : "text-macos-text-tertiary dark:text-macos-dark-text-tertiary"
                    }`}
                  >
                    ⌘{index + 1}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="p-8 text-center">
              <div className="text-4xl mb-3">🔍</div>
              <div className="text-lg font-medium text-macos-text-primary dark:text-macos-dark-text-primary mb-1">
                No matching results
              </div>
              <div className="text-sm text-macos-text-secondary dark:text-macos-dark-text-secondary">
                Try a different search term
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={() => setContextMenu(null)}
          />
          <div
            className="fixed z-50 w-48 bg-macos-bg-primary dark:bg-macos-dark-bg-primary backdrop-blur-macos rounded-macos shadow-macos dark:shadow-macos-dark border border-macos-border dark:border-macos-dark-border p-1"
            style={{ left: contextMenu.x, top: contextMenu.y }}
          >
            {(() => {
              const item = items.find((i) => i.id === contextMenu.itemId);
              return item ? (
                <>
                  <button
                    onClick={() => {
                      onItemPin(contextMenu.itemId, !item.isPinned);
                      setContextMenu(null);
                    }}
                    className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-blue hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
                  >
                    <span className="mr-3">{item.isPinned ? "📌" : "📍"}</span>
                    {item.isPinned ? "Unpin" : "Pin"}
                  </button>

                  <button
                    onClick={() => {
                      onItemDelete(contextMenu.itemId);
                      setContextMenu(null);
                    }}
                    className="w-full flex items-center px-3 py-2 text-left hover:bg-macos-accent-red hover:text-white rounded-macos-input transition-colors duration-150 text-sm"
                  >
                    <span className="mr-3">🗑️</span>
                    Delete
                  </button>
                </>
              ) : null;
            })()}
          </div>
        </>
      )}
    </>
  );
};

export default ClipboardSearch;
