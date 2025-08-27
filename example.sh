#!/bin/bash

# Example script demonstrating the caching proxy server
# This script shows how to build, run, and test the caching proxy

echo "🚀 Building the caching proxy server..."
make build

if [ $? -ne 0 ]; then
    echo "❌ Build failed. Please check the error messages above."
    exit 1
fi

echo "✅ Build successful!"

echo ""
echo "🔧 Starting the caching proxy server on port 3000..."
echo "   Forwarding requests to: http://dummyjson.com"
echo ""

# Start the server in the background
./caching-proxy --port 3000 --origin http://dummyjson.com &
SERVER_PID=$!

# Wait a moment for the server to start
sleep 2

echo "🧪 Testing the proxy server..."
echo ""

# Test 1: First request (should be a cache MISS)
echo "📡 Test 1: First request to /products (Cache MISS expected)"
curl -s -I http://localhost:3000/products | grep "X-Cache"
echo ""

# Test 2: Second request (should be a cache HIT)
echo "📡 Test 2: Second request to /products (Cache HIT expected)"
curl -s -I http://localhost:3000/products | grep "X-Cache"
echo ""

# Test 3: Different endpoint
echo "📡 Test 3: Request to /users (Cache MISS expected)"
curl -s -I http://localhost:3000/users | grep "X-Cache"
echo ""

# Test 4: Same endpoint again (should be a cache HIT)
echo "📡 Test 4: Second request to /users (Cache HIT expected)"
curl -s -I http://localhost:3000/users | grep "X-Cache"
echo ""

echo "🧹 Cleaning up..."
# Stop the server
kill $SERVER_PID 2>/dev/null

echo ""
echo "✨ Example completed successfully!"
echo ""
echo "💡 To run the server manually:"
echo "   ./caching-proxy --port 3000 --origin http://dummyjson.com"
echo ""
echo "💡 To clear the cache:"
echo "   ./caching-proxy clear-cache"
echo ""
echo "💡 To see all available commands:"
echo "   make help"
