package cache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Cache struct {
	mu      *sync.Mutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) Cache {
	cache := Cache{
		mu:      &sync.Mutex{},
		entries: map[string]cacheEntry{},
	}
	go cache.reapLoop(interval)
	return cache
}
func (cache Cache) Add(key string, val []byte) {
	(*cache.mu).Lock()
	defer (*cache.mu).Unlock()

	cache.entries[key] = cacheEntry{
		val:       val,
		createdAt: time.Now(),
	}
}
func (cache Cache) Get(key string) ([]byte, bool) {
	(*cache.mu).Lock()
	defer (*cache.mu).Unlock()

	entry, exists := cache.entries[key]
	if exists {
		log.Println("Returning cached results")
		return entry.val, true
	}
	return []byte{}, false
}
func (cache Cache) HttpGet(url string) ([]byte, error) {
	if data, exists := cache.Get(url); exists {
		return data, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	cache.Add(url, body)

	return body, nil
}
func (cache Cache) Remove(key string) {
	(*cache.mu).Lock()
	defer (*cache.mu).Unlock()

	delete(cache.entries, key)
	log.Println("Cleared cache for", key)
}

func (cache Cache) reapLoop(interval time.Duration) {
	for {
		time.Sleep(interval)

		(*cache.mu).Lock()

		toRemove := []string{}
		for key, entry := range cache.entries {
			if time.Since(entry.createdAt) > interval {
				toRemove = append(toRemove, key)
			}
		}
		for _, key := range toRemove {
			delete(cache.entries, key)
		}

		(*cache.mu).Unlock()
	}
}
