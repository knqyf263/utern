package cache

import (
	"sync"
)

var (
	// Cache represents cache client
	Cache = NewCache()
)

// EventCache represents cache client
type EventCache struct {
	mapData *sync.Map
}

// NewCache returns cache client
func NewCache() *EventCache {
	return &EventCache{mapData: new(sync.Map)}
}

// Store saves key and value
func (c *EventCache) Store(logGroupName string, key *string, value *int64) {
	v, ok := c.mapData.Load(logGroupName)
	var cache map[string]*int64
	if !ok {
		cache = map[string]*int64{}
		c.mapData.Store(logGroupName, cache)
	} else {
		cache = v.(map[string]*int64)
	}
	cache[*key] = value
}

// Load gets value
func (c *EventCache) Load(logGroupName string, key *string) (ok bool) {
	v, ok := c.mapData.Load(logGroupName)
	if !ok {
		return false
	}
	cache := v.(map[string]*int64)
	_, ok = cache[*key]
	return ok
}

// Expire deletes cache older than the given timestamp
func (c *EventCache) Expire(logGroupName string, lastSeen *int64) {
	v, ok := c.mapData.Load(logGroupName)
	if !ok {
		return
	}
	cache := v.(map[string]*int64)
	for eventID, timestamp := range cache {
		if *timestamp < *lastSeen {
			delete(cache, eventID)
		}
	}
}
