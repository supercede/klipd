# Klipd - Intelligent macOS Clipboard Manager

A fast, native macOS clipboard manager built with Wails (Go + React) that provides instant access to your clipboard history with smart organization and search capabilities.

## Features

### Core Functionality

- **Automatic Clipboard Monitoring**: Captures text content with configurable polling intervals
- **Smart Duplicate Detection**: Prevents storing identical consecutive clipboard entries
- **Real-time Search**: Instant filtering of clipboard history with regex support
- **Persistent Storage**: SQLite-based storage that survives app restarts
- **Pin System**: Keep important items protected from auto-cleanup
- **Auto-cleanup**: Configurable cleanup by age (days) and count limits

### Quick Access

- **Global Hotkeys**: Access clipboard history from anywhere
- **Keyboard Navigation**: Full keyboard control for power users
- **Recent Items**: Quick access to your most recent clipboard entries

### Native macOS Design

- **Dark Mode Support**: Automatic theme switching
- **Native Fonts**: Uses SF Pro system fonts

## Installation

### Prerequisites

- macOS 10.13+ (High Sierra or later)
- No additional dependencies required

### Quick Start

1. Download the latest release from [releases page]
2. Drag `Klipd.app` to your Applications folder
3. Launch Klipd from Applications or Spotlight
4. Grant clipboard access permissions when prompted

### Building from Source

```bash
# Prerequisites
brew install go node
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and build
git clone <repository-url>
cd klipd
make build

# Development mode
make dev
```

## Usage

### Global Hotkeys

| Hotkey    | Action          | Description                             |
| --------- | --------------- | --------------------------------------- |
| `⌘⇧Space` | Show History    | Opens searchable clipboard history      |
| `⌘⇧k`     | Focus on window | Opens the main application window       |
| `⌘⇧C`     | Paste Previous  | Immediately pastes the last copied item |

#### In Main Interface:

- `⌘,` - Open settings
- `⌘Q` - Quit application

### Search Features

#### Basic Search

- Type any text to filter clipboard history
- Search works across all text content and previews
- Real-time results as you type (300ms debounce)

#### Advanced Search

- **Regex Toggle**: Enable regex patterns for complex searches

#### Search Examples

```
hello world          # Simple text search
^https://            # URLs (with regex enabled)
console\.log         # Code patterns (with regex enabled)
```

## Settings & Configuration

### General Settings

#### Monitoring

- **Polling Interval**: How often to check clipboard (100ms - 2000ms)
  - Lower = more responsive, higher CPU usage
  - Higher = less responsive, better battery life
  - Default: 500ms (recommended)
- **Enable Monitoring**: Toggle clipboard capture on/off

#### Privacy & Security

- **Allow Passwords**: Whether to capture password-like content
  - When disabled: Content matching password patterns is skipped
  - Patterns: Long random strings, base64-like content, etc.

### History Management

#### Auto-cleanup

- **Max Items**: Maximum clipboard items to keep (default: 100)
- **Max Days**: Auto-delete items older than X days (default: 7)
- **Cleanup Interval**: How often cleanup runs (default: 1 hour)

#### Manual Management

- **Pin Items**: Protect items from auto-cleanup

### Advanced Settings

## Content Types & Icons

### Supported Content Types

| Type | Icon | Description           | Storage             |
| ---- | ---- | --------------------- | ------------------- |
| Text | 📄   | Plain text, code, etc | Full text + preview |

### Content Detection

- **Automatic**: Klipd automatically detects content types
- **Smart Previews**: Truncated previews optimized for quick recognition
- **Metadata**: File sizes, dimensions, creation timestamps

## Privacy & Security

### Data Storage

- **Local Only**: All data stored locally on your Mac
- **No Cloud Sync**: No data transmitted to external servers
- **Encrypted**: Database uses SQLite with file system encryption

### Sensitive Content

- **Password Detection**: Heuristic detection of password-like content
- **Opt-in Capture**: Passwords only stored if explicitly enabled

### Permissions

- **Clipboard Access**: Required for core functionality
- **Accessibility**: Optional, needed for global hotkeys
- **Notifications**: Optional, for status updates

## Troubleshooting

### Common Issues

#### Hotkeys Not Working

1. Check System Preferences → Security & Privacy → Accessibility
2. Ensure Klipd is in the allowed apps list
3. Try different hotkey combinations in settings
4. Restart the application

#### Clipboard Not Capturing

1. Verify monitoring is enabled in settings
2. Check if content matches skip patterns (passwords, etc.)
3. Ensure polling interval isn't too high
4. Check system clipboard permissions

#### Performance Issues

1. Increase polling interval in settings
2. Reduce max items in auto-cleanup
3. Clear old clipboard history
4. Check available disk space

#### Search Not Working

1. Try disabling regex mode
2. Check for special characters in search
3. Clear search and try again
4. Restart application if persistent

### Advanced Troubleshooting

#### Database Issues

```bash
# Check database location
ls ~/Library/Application\ Support/Klipd/

# View database size
du -h ~/Library/Application\ Support/Klipd/clipboard.db

# Reset database (WARNING: Deletes all history)
rm ~/Library/Application\ Support/Klipd/clipboard.db
```

## Development

### Tech Stack

- **Backend**: Go 1.19+ with Wails v2
- **Frontend**: React 18 + TypeScript + Tailwind CSS
- **Database**: SQLite with GORM
- **Clipboard**: `github.com/atotto/clipboard`
- **Hotkeys**: `golang.design/x/hotkey`

### Build Commands

```bash
# Development
make dev                 # Start development server
make build-debug        # Build with debug console

# Production
make build              # Build production app
make test               # Run tests
make lint               # Run linters

# Maintenance
make clean              # Clean build artifacts
make check              # Run all checks
make package            # Create distribution package
```

### Project Structure

```
klipd/
├── app.go              # Main application logic
├── main.go             # Entry point
├── config/             # Configuration management
├── database/           # SQLite operations
├── models/             # Data structures
├── services/           # Business logic
│   ├── clipboard.go    # Clipboard monitoring
│   └── hotkey.go       # Global hotkey handling
└── frontend/           # React frontend
    ├── src/
    │   ├── components/  # React components
    │   ├── hooks/       # Custom hooks
    │   └── utils/       # Utility functions
    └── wailsjs/        # Generated Wails bindings
```

## Contributing

### Getting Started

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to verify
6. Submit a pull request

### Code Style

- **Go**: Follow `gofmt` and `golint` standards
- **TypeScript**: ESLint + Prettier configuration
- **Commits**: Conventional commit format

### Testing

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Test specific packages
go test ./database/
go test ./config/
```

## License

[License information here]

## Acknowledgements

- App icon by [pngtree.com](https://pngtree.com/freepng/clipboard-vector_7965293.html)
