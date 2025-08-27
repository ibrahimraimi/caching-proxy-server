package main

import (
	"net/http"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := NewCache()

	entry := &CacheEntry{
		Body:       []byte("test response"),
		Headers:    http.Header{"Content-Type": []string{"text/plain"}},
		StatusCode: 200,
		Timestamp:  time.Now(),
		TTL:        1 * time.Minute,
	}

	cache.Set("test-key", entry)

	if retrieved, hit := cache.Get("test-key"); !hit {
		t.Error("Expected cache hit, got miss")
	} else if string(retrieved.Body) != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", string(retrieved.Body))
	}

	if _, hit := cache.Get("non-existent-key"); hit {
		t.Error("Expected cache miss, got hit")
	}

	if size := cache.Size(); size != 1 {
		t.Errorf("Expected cache size 1, got %d", size)
	}

	cache.Clear()
	if size := cache.Size(); size != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", size)
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewCache()

	entry := &CacheEntry{
		Body:       []byte("expired response"),
		Headers:    http.Header{},
		StatusCode: 200,
		Timestamp:  time.Now().Add(-2 * time.Minute), // 2 minutes ago
		TTL:        1 * time.Minute,                  // 1 minute TTL
	}

	cache.Set("expired-key", entry)

	if _, hit := cache.Get("expired-key"); hit {
		t.Error("Expected cache miss for expired entry, got hit")
	}

	if size := cache.Size(); size != 0 {
		t.Errorf("Expected cache size 0 after expiration, got %d", size)
	}
}

func TestProxyServerCreation(t *testing.T) {
	proxy, err := NewProxyServer("http://example.com", 8080)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if proxy.port != 8080 {
		t.Errorf("Expected port 8080, got %d", proxy.port)
	}
	if proxy.origin.String() != "http://example.com" {
		t.Errorf("Expected origin 'http://example.com', got '%s'", proxy.origin.String())
	}

	_, err = NewProxyServer("://invalid-url", 8080)
	if err == nil {
		t.Error("Expected error for invalid URL, got none")
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	proxy, _ := NewProxyServer("http://example.com", 8080)

	req, _ := http.NewRequest("GET", "http://localhost:8080/test", nil)
	req.Header.Set("User-Agent", "test-agent")

	key1 := proxy.generateCacheKey(req)
	key2 := proxy.generateCacheKey(req)

	if key1 != key2 {
		t.Error("Expected same cache key for same request")
	}

	req2, _ := http.NewRequest("POST", "http://localhost:8080/test", nil)
	req2.Header.Set("User-Agent", "test-agent")
	key3 := proxy.generateCacheKey(req2)

	if key1 == key3 {
		t.Error("Expected different cache keys for different requests")
	}
}
