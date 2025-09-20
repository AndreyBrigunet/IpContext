package cache

import (
	"sync"
	"time"
)

// Item represents a cached item with expiration
type Item struct {
	Value      interface{}
	Expiration int64
}

// IsExpired checks if the item has expired
func (item Item) IsExpired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Cache represents an in-memory cache with TTL
type Cache struct {
	mu    sync.RWMutex
	items map[string]Item
	ttl   time.Duration
}

// New creates a new cache with the specified TTL
func New(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]Item),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go c.cleanup()
	
	return c
}

// Set adds an item to the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	expiration := time.Now().Add(c.ttl).UnixNano()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	
	if item.IsExpired() {
		return nil, false
	}
	
	return item.Value, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item)
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// cleanup removes expired items periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for key, item := range c.items {
				if item.IsExpired() {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		}
	}
}
