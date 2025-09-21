package inmemory

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type Options struct {
	MaxIdle   int
	MaxActive int
	Timeout   int
	Wait      bool
	AuthKey   string
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

type Inmemory struct {
	Options Options
	mu      sync.RWMutex
	cache   map[string]*cacheItem
}

var (
	instance *Inmemory
	mu       sync.Mutex
)

// GetInstance returns a singleton instance of Inmemory cache
func GetInstance(options Options) *Inmemory {
	if instance == nil {
		mu.Lock()
		defer mu.Unlock()
		if instance == nil {
			instance = &Inmemory{
				Options: options,
				cache:   make(map[string]*cacheItem),
			}
			// Start cleanup goroutine
			go instance.startCleanup()
		}
	}
	return instance
}

// New is kept for backward compatibility but now returns singleton
func New(options Options) *Inmemory {
	return GetInstance(options)
}

func (i *Inmemory) Get(key string) (string, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	item, ok := i.cache[key]
	if !ok {
		return "", errors.New("key not found")
	}

	// Check if expired
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return "", errors.New("key expired")
	}

	return item.value, nil
}

func (i *Inmemory) SetWithExpire(key string, value interface{}, expire int) (string, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	item := &cacheItem{
		value: value.(string),
	}

	// Set expiration if provided
	if expire > 0 {
		item.expiresAt = time.Now().Add(time.Duration(expire) * time.Second)
	}

	i.cache[key] = item
	return "OK", nil
}

func (i *Inmemory) Del(key string) (int64, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	keys := strings.Split(key, " ")

	count := int64(0)
	for _, eachKey := range keys {
		if _, ok := i.cache[eachKey]; ok {
			delete(i.cache, eachKey)
			count++
		}
	}

	return count, nil
}

// startCleanup runs a background goroutine to clean expired items
func (i *Inmemory) startCleanup() {
	ticker := time.NewTicker(1 * time.Minute) // Clean every minute
	defer ticker.Stop()

	for range ticker.C {
		i.cleanupExpired()
	}
}

// cleanupExpired removes expired items from cache
func (i *Inmemory) cleanupExpired() {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	for key, item := range i.cache {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			delete(i.cache, key)
		}
	}
}

// GetStats returns cache statistics
func (i *Inmemory) GetStats() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return map[string]interface{}{
		"total_keys": len(i.cache),
		"max_idle":   i.Options.MaxIdle,
		"max_active": i.Options.MaxActive,
	}
}

// Clear removes all items from cache
func (i *Inmemory) Clear() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.cache = make(map[string]*cacheItem)
	return nil
}
