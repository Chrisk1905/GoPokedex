package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entries map[string]cacheEntry
	mutex   sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// constructor for Cache
func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]cacheEntry),
	}

	cache.reapLoop(interval)

	return cache
}

// adds a new entry to the cache
func (c *Cache) Add(key string, val []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.entries[key] = entry
}

// .Get() gets an entry from the cache.
// It should take a key (a string) and return a []byte and a bool. The bool should be true if the entry was found and false if it wasn't.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(interval * -1)
			c.mutex.Lock()
			for key, val := range c.entries {
				if val.createdAt.Before(cutoff) {
					delete(c.entries, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}
