package mcpcontract

import (
	"sync"
	"testing"
	"time"
)

func TestBoundedTimezoneCacheEvictsRoundRobin(t *testing.T) {
	cache := &boundedTimezoneCache{
		values:   make(map[string]*time.Location),
		capacity: 2,
	}
	first := time.FixedZone("first", 0)
	second := time.FixedZone("second", 3600)
	third := time.FixedZone("third", 7200)

	cache.put("first", first)
	cache.put("second", second)
	cache.put("third", third)

	if _, exists := cache.values["first"]; exists {
		t.Fatal("first insertion was retained after eviction")
	}
	if cache.values["second"] != second || cache.values["third"] != third {
		t.Fatalf("cache values = %v/%v, want second/third", cache.values["second"], cache.values["third"])
	}
	if cache.order[0] != "third" || cache.order[1] != "second" || cache.next != 1 {
		t.Fatalf("order = %v, next = %d, want [third second]/1", cache.order, cache.next)
	}
}

func TestLoadTimezoneConcurrentRequestsAreSafe(t *testing.T) {
	zones := []string{"Asia/Shanghai", "Asia/Tokyo", "Europe/London", "America/New_York"}
	var wg sync.WaitGroup
	for _, zone := range zones {
		for range 8 {
			wg.Add(1)
			go func(zone string) {
				defer wg.Done()
				if _, err := loadTimezone(zone); err != nil {
					t.Errorf("loadTimezone(%s): %v", zone, err)
				}
			}(zone)
		}
	}
	wg.Wait()
}

func TestLoadTimezoneCachesImmutableLocation(t *testing.T) {
	first, err := loadTimezone("Asia/Shanghai")
	if err != nil {
		t.Fatalf("loadTimezone: %v", err)
	}
	second, err := loadTimezone("Asia/Shanghai")
	if err != nil {
		t.Fatalf("second loadTimezone: %v", err)
	}
	if first != second {
		t.Fatal("repeated timezone load returned distinct immutable locations")
	}
}
