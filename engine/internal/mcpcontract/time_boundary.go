package mcpcontract

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// loadTimezone loads an explicit IANA timezone. Fixed offsets, Local, and bare
// abbreviations are not accepted; timezone identity and offset are separate.
func loadTimezone(value string) (*time.Location, error) {
	if value == "" {
		return nil, errors.New("timezone is required")
	}

	if strings.EqualFold(value, "Local") {
		return nil, fmt.Errorf("timezone %q is not allowed", value)
	}

	if cached, ok := timezoneCache.get(value); ok {
		return cached, nil
	}
	loc, err := time.LoadLocation(value)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", value, err)
	}
	timezoneCache.put(value, loc)

	return loc, nil
}

type boundedTimezoneCache struct {
	mu       sync.RWMutex
	values   map[string]*time.Location
	order    []string
	next     int
	capacity int
}

func (c *boundedTimezoneCache) get(name string) (*time.Location, bool) {
	c.mu.RLock()
	location, exists := c.values[name]
	c.mu.RUnlock()
	return location, exists
}

func (c *boundedTimezoneCache) put(name string, location *time.Location) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.values[name]; exists {
		return
	}
	if len(c.order) < c.capacity {
		c.order = append(c.order, name)
	} else {
		delete(c.values, c.order[c.next])
		c.order[c.next] = name
		c.next = (c.next + 1) % c.capacity
	}
	c.values[name] = location
}

var timezoneCache = &boundedTimezoneCache{
	values:   make(map[string]*time.Location, 256),
	capacity: 1024,
}
