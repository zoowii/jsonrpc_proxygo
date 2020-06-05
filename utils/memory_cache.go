package utils

/**
 * some code of memory_cache is copied from https://github.com/patrickmn/go-cache/blob/master/cache.go
 */

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Item struct {
	Object     interface{}
	Expiration int64
}

// Returns true if the item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

func (v Item) ObjectAsInt64() (intValue int64, err error) {
	switch v.Object.(type) {
	case int:
		intValue = int64(v.Object.(int))
	case int8:
		intValue = int64(v.Object.(int8))
	case int16:
		intValue = int64(v.Object.(int16))
	case int32:
		intValue = int64(v.Object.(int32))
	case int64:
		intValue = v.Object.(int64)
	case uint:
		intValue = int64(v.Object.(uint))
	case uintptr:
		intValue = int64(v.Object.(uintptr))
	case uint8:
		intValue = int64(v.Object.(uint8))
	case uint16:
		intValue = int64(v.Object.(uint16))
	case uint32:
		intValue = int64(v.Object.(uint32))
	case uint64:
		intValue = int64(v.Object.(uint64))
	case float32:
		intValue = int64(v.Object.(float32))
	case float64:
		intValue = int64(v.Object.(float64))
	default:
		err = errors.New("the value is not an integer")
	}
	return
}

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type MemoryCache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]Item),
	}
}

// Add an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *MemoryCache) Set(k string, x interface{}, d time.Duration) {
	// "Inlining" of set
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
}

func (c *MemoryCache) set(k string, x interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		Object:     x,
		Expiration: e,
	}
}

// Add an item to the cache, replacing any existing item, using the default
// expiration.
func (c *MemoryCache) SetDefault(k string, x interface{}) {
	c.Set(k, x, DefaultExpiration)
}

// Set a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (c *MemoryCache) Replace(k string, x interface{}, d time.Duration) error {
	c.mu.Lock()
	_, found := c.get(k)
	if !found {
		c.mu.Unlock()
		return fmt.Errorf("Item %s doesn't exist", k)
	}
	c.set(k, x, d)
	c.mu.Unlock()
	return nil
}

// Get an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *MemoryCache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			c.mu.RUnlock()
			return nil, false
		}
	}
	c.mu.RUnlock()
	return item.Object, true
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (c *MemoryCache) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	c.mu.RLock()
	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		c.mu.RUnlock()
		return nil, time.Time{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			c.mu.RUnlock()
			return nil, time.Time{}, false
		}

		// Return the item and the expiration time
		c.mu.RUnlock()
		return item.Object, time.Unix(0, item.Expiration), true
	}

	// If expiration <= 0 (i.e. no expiration time set) then return the item
	// and a zeroed time.Time
	c.mu.RUnlock()
	return item.Object, time.Time{}, true
}

func (c *MemoryCache) get(k string) (interface{}, bool) {
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	if item.Expired() {
		return nil, false
	}
	return item.Object, true
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (c *MemoryCache) Increment(k string, n int64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return fmt.Errorf("Item %s not found", k)
	}
	switch v.Object.(type) {
	case int:
		v.Object = v.Object.(int) + int(n)
	case int8:
		v.Object = v.Object.(int8) + int8(n)
	case int16:
		v.Object = v.Object.(int16) + int16(n)
	case int32:
		v.Object = v.Object.(int32) + int32(n)
	case int64:
		v.Object = v.Object.(int64) + n
	case uint:
		v.Object = v.Object.(uint) + uint(n)
	case uintptr:
		v.Object = v.Object.(uintptr) + uintptr(n)
	case uint8:
		v.Object = v.Object.(uint8) + uint8(n)
	case uint16:
		v.Object = v.Object.(uint16) + uint16(n)
	case uint32:
		v.Object = v.Object.(uint32) + uint32(n)
	case uint64:
		v.Object = v.Object.(uint64) + uint64(n)
	case float32:
		v.Object = v.Object.(float32) + float32(n)
	case float64:
		v.Object = v.Object.(float64) + float64(n)
	default:
		c.mu.Unlock()
		return fmt.Errorf("The value for %s is not an integer", k)
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

// Delete all items from the cache.
func (c *MemoryCache) Flush() {
	c.mu.Lock()
	c.items = map[string]Item{}
	c.mu.Unlock()
}

func (c *MemoryCache) DumpItems() (result map[string]Item, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result = make(map[string]Item)
	for k, v := range c.items {
		result[k] = v
	}
	return
}
