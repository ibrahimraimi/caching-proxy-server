# Caching Proxy Server

A high-performance caching proxy server built in Go that forwards requests to origin servers and caches responses. If the same request is made again, it returns the cached response instead of forwarding to the server.

**✨ Now featuring a beautiful Terminal User Interface (TUI) built with Bubble Tea!**

## Features

- **Smart Caching**: In-memory cache with configurable TTL (default: 5 minutes)
- **Cache Headers**: Automatic `X-Cache: HIT/MISS` headers to indicate cache status
- **Thread-Safe**: Concurrent request handling with proper locking
- **CLI Interface**: Easy-to-use command-line interface with Cobra
- **Flexible Configuration**: Customizable port and origin server
- **Cache Management**: Built-in cache clearing functionality
- **Comprehensive Testing**: Unit tests covering all major functionality
- **Automated Examples**: Scripts for testing and demonstration
- **🎨 Beautiful TUI**: Terminal User Interface with real-time monitoring and control

## Requirements

- Go 1.21 or higher
- Linux/macOS/Windows (cross-platform)
- Terminal with support for colors and UTF-8

## Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd caching-proxy-server
```

2. Install dependencies:

```bash
make deps
# or manually:
go mod download
```

3. Build the project:

```bash
make build
# or manually:
go build -o caching-proxy main.go
```

## Usage

### Basic Usage

Start the caching proxy server:

```bash
caching-proxy --port <number> --origin <url>
```

**Parameters:**

- `--port` (required): Port on which the caching proxy server will run
- `--origin` (required): URL of the server to which requests will be forwarded

### TUI Mode (Recommended)

For the best experience, run with the beautiful terminal interface:

```bash
caching-proxy --port <number> --origin <url> --tui
```

**TUI Features:**

- 🎯 Real-time request monitoring
- 📊 Live cache statistics
- 🎨 Beautiful color-coded interface
- ⌨️ Interactive keyboard controls
- 📱 Responsive design

**TUI Controls:**

- `↑/↓` - Navigate through requests
- `c` - Clear cache
- `r` - Refresh data
- `q` - Quit

### Examples

#### Example 1: Forward to DummyJSON API with TUI

```bash
caching-proxy --port 3000 --origin http://dummyjson.com --tui
```

This will:

- Start the proxy server on port 3000
- Forward requests to http://dummyjson.com
- Cache responses for 5 minutes
- Display beautiful TUI interface

#### Example 2: Forward to a different API

```bash
caching-proxy --port 8080 --origin https://api.example.com --tui
```

#### Example 3: Using short flags

```bash
caching-proxy -p 3000 -o http://dummyjson.com -t
```

### Testing the Proxy

Once the server is running, you can test it:

1. **First request** (cache miss):

```bash
curl http://localhost:3000/products
# Response will include: X-Cache: MISS
```

2. **Second request** (cache hit):

```bash
curl http://localhost:3000/products
# Response will include: X-Cache: HIT
```

3. **Check cache headers**:

```bash
curl -I http://localhost:3000/products
```

### Cache Management

#### Clear the Cache

```bash
caching-proxy clear-cache
```

#### Using Makefile Commands

The project includes a Makefile for easier management:

```bash
# Build and run with example settings
make run

# Build and run with TUI (recommended)
make run-tui

# Run with custom settings
make run-custom PORT=8080 ORIGIN=http://api.example.com

# Run custom with TUI
make run-custom-tui PORT=8080 ORIGIN=http://api.example.com

# Clear cache
make clear-cache

# Clean build artifacts
make clean

# Show all available commands
make help
```

#### Automated Testing

Run the complete example script to test all functionality:

```bash
./example.sh
```

This script will:

- Build the project
- Start the server
- Test cache hits and misses
- Clean up automatically

## How It Works

1. **Request Processing**: When a request comes in, the proxy generates a unique cache key based on the HTTP method, URL, and User-Agent header.

2. **Cache Lookup**: The proxy checks if a cached response exists for the request.

3. **Cache Hit**: If found and not expired, returns the cached response with `X-Cache: HIT` header.

4. **Cache Miss**: If not found or expired, forwards the request to the origin server.

5. **Response Caching**: Successful responses (status 200-399) are cached with a 5-minute TTL.

6. **Response Delivery**: The response is returned to the client with `X-Cache: MISS` header.

7. **TUI Updates**: All requests are logged and displayed in real-time in the beautiful interface.

## Cache Key Generation

The cache key is generated using MD5 hash of:

- HTTP method (GET, POST, etc.)
- Full request URL
- User-Agent header

This ensures that different types of requests are cached separately.

## Configuration

### Cache TTL

The default cache TTL is 5 minutes. To modify this, edit the `TTL` field in the `CacheEntry` struct in `main.go`.

### Cache Storage

Currently uses in-memory storage. For production use, consider implementing:

- Redis backend
- File-based persistence
- Database storage

## Project Structure

```
caching-proxy-server/
├── main.go              # Main application with server logic and TUI
├── main_test.go         # Comprehensive test suite
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Makefile             # Build and management commands
├── example.sh           # Automated testing script
├── demo-tui.sh          # TUI demonstration script
├── config.example.yaml  # Future configuration structure
└── README.md            # This file
```

## Dependencies

- **github.com/spf13/cobra**: CLI framework for command-line interface
- **github.com/charmbracelet/bubbletea**: TUI framework for beautiful terminal interfaces
- **github.com/charmbracelet/lipgloss**: Styling library for terminal UI
- **Standard Go libraries**: net/http, crypto/md5, sync, etc.

## Performance Considerations

- **Memory Usage**: Cache entries are stored in memory, so monitor memory usage for high-traffic scenarios
- **Concurrency**: Uses read-write mutexes for optimal concurrent access
- **TTL Management**: Expired entries are automatically cleaned up on access
- **Cache Efficiency**: MD5-based keys provide fast lookups with minimal collision probability
- **TUI Performance**: Real-time updates with minimal overhead

## Testing

### Run Unit Tests

```bash
make test
```

### Run Example Script

```bash
./example.sh
```

### Manual Testing

```bash
# Start server with TUI
make run-tui

# In another terminal, test caching
curl http://localhost:3000/products
curl http://localhost:3000/products  # Should show cache HIT
```

## Troubleshooting

### Common Issues

1. **Port already in use**: Choose a different port or stop the service using the current port
2. **Invalid origin URL**: Ensure the origin URL is properly formatted (include http:// or https://)
3. **Permission denied**: Make sure you have permission to bind to the specified port
4. **Build failures**: Ensure Go 1.21+ is installed and dependencies are downloaded
5. **TUI display issues**: Ensure your terminal supports colors and UTF-8

### Debug Mode

The server logs all cache hits and misses. Check the console output for debugging information:

```
2024/01/15 10:30:00 Starting caching proxy server on port 3000
2024/01/15 10:30:00 Forwarding requests to: http://dummyjson.com
2024/01/15 10:30:00 Cache size: 0 entries
2024/01/15 10:30:05 Cache MISS for /products
2024/01/15 10:30:05 Cached response for /products
2024/01/15 10:30:10 Cache HIT for /products
```

## CLI Commands

### Main Command

```bash
caching-proxy --port <port> --origin <url>
```

### TUI Mode

```bash
caching-proxy --port <port> --origin <url> --tui
```

### Subcommands

```bash
caching-proxy clear-cache    # Clear the cache
caching-proxy help          # Show help
caching-proxy completion    # Generate shell completion
```

### Flags

- `-p, --port`: Port number (1-65535)
- `-o, --origin`: Origin server URL
- `-t, --tui`: Enable beautiful terminal user interface
- `-h, --help`: Show help information

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

### Development Setup

```bash
# Install dependencies
make deps

# Run tests
make test

# Build
make build

# Clean
make clean
```

## Example Output

### TUI Interface

```
🚀 Caching Proxy Server

Status: Running
Port: 3000 | Origin: http://dummyjson.com | Cache Size: 5 entries

Controls: ↑/↓ Navigate | c Clear Cache | r Refresh | q Quit

Recent Requests:
Time                Method  Path                           Status  Cache   Response Time
15:04:05           GET     /products                      200     HIT     1.2ms
15:04:03           GET     /users                         200     MISS    45.8ms
15:04:01           GET     /products                      200     MISS    52.1ms

Press 'q' to quit
```

## Quick Reference

| Command                | Description               |
| ---------------------- | ------------------------- |
| `make build`           | Build the project         |
| `make run`             | Run with example settings |
| `make run-tui`         | Run with beautiful TUI    |
| `make test`            | Run unit tests            |
| `make clear-cache`     | Clear the cache           |
| `./example.sh`         | Run automated testing     |
| `./demo-tui.sh`        | Demo TUI functionality    |
| `caching-proxy --help` | Show CLI help             |

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the test examples
3. Check the console logs for debugging information
4. Open an issue in the repository
