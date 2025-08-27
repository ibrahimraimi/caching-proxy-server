#!/bin/bash

# Demo script for the TUI functionality of the caching proxy server
# This script demonstrates the beautiful terminal interface

echo "🎨 Caching Proxy Server - TUI Demo"
echo "=================================="
echo ""

echo "🚀 Building the project..."
make build

if [ $? -ne 0 ]; then
    echo "❌ Build failed. Please check the error messages above."
    exit 1
fi

echo "✅ Build successful!"
echo ""

echo "🎯 Starting the caching proxy server with TUI..."
echo "   Port: 3000"
echo "   Origin: http://dummyjson.com"
echo "   Interface: Beautiful TUI (Bubble Tea)"
echo ""

echo "📱 TUI Features:"
echo "   • Real-time request monitoring"
echo "   • Live cache statistics"
echo "   • Beautiful color-coded interface"
echo "   • Interactive keyboard controls"
echo "   • Responsive design"
echo ""

echo "⌨️  TUI Controls:"
echo "   ↑/↓ - Navigate through requests"
echo "   c   - Clear cache"
echo "   r   - Refresh data"
echo "   q   - Quit"
echo ""

echo "🧪 To test the TUI:"
echo "   1. The TUI will start automatically"
echo "   2. Open another terminal and run:"
echo "      curl http://localhost:3000/products"
echo "      curl http://localhost:3000/users"
echo "      curl http://localhost:3000/products  # Should show cache HIT"
echo "   3. Watch the beautiful interface update in real-time!"
echo ""

echo "🎬 Starting TUI in 3 seconds..."
sleep 3

# Start the server with TUI
./caching-proxy --port 3000 --origin http://dummyjson.com --tui
