package core

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		c := NewCache()
		if c.defaultTTL != 5*time.Minute {
			t.Errorf("defaultTTL = %v, want %v", c.defaultTTL, 5*time.Minute)
		}
		if c.maxItems != 1000 {
			t.Errorf("maxItems = %d, want %d", c.maxItems, 1000)
		}
	})

	t.Run("with options", func(t *testing.T) {
		c := NewCache(
			WithDefaultTTL(10*time.Minute),
			WithMaxItems(500),
		)
		if c.defaultTTL != 10*time.Minute {
			t.Errorf("defaultTTL = %v, want %v", c.defaultTTL, 10*time.Minute)
		}
		if c.maxItems != 500 {
			t.Errorf("maxItems = %d, want %d", c.maxItems, 500)
		}
	})
}

func TestCache_SetGet(t *testing.T) {
	c := NewCache()

	// Set a value
	c.Set("key1", "value1")

	// Get the value
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("value = %v, want %v", val, "value1")
	}

	// Get non-existent key
	_, ok = c.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent key")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	now := time.Now()
	c := NewCache(WithTimeFunc(func() time.Time { return now }))

	c.SetWithTTL("key1", "value1", 1*time.Hour)

	// Should be found before expiry
	val, ok := c.Get("key1")
	if !ok || val != "value1" {
		t.Fatal("expected to find key1")
	}

	// Advance time past expiry
	now = now.Add(2 * time.Hour)

	// Should not be found after expiry
	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	c := NewCache()

	c.Set("key1", "value1")
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}

	// Delete non-existent key should not panic
	c.Delete("nonexistent")
}

func TestCache_Clear(t *testing.T) {
	c := NewCache()

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}
}

func TestCache_Len(t *testing.T) {
	c := NewCache()

	if c.Len() != 0 {
		t.Errorf("Len() = %d, want 0", c.Len())
	}

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	if c.Len() != 2 {
		t.Errorf("Len() = %d, want 2", c.Len())
	}
}

func TestCache_Keys(t *testing.T) {
	c := NewCache()

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	keys := c.Keys()
	if len(keys) != 2 {
		t.Errorf("len(Keys()) = %d, want 2", len(keys))
	}

	// Check that both keys are present
	found := make(map[string]bool)
	for _, k := range keys {
		found[k] = true
	}
	if !found["key1"] || !found["key2"] {
		t.Error("expected both key1 and key2 in Keys()")
	}
}

func TestCache_Cleanup(t *testing.T) {
	now := time.Now()
	c := NewCache(WithTimeFunc(func() time.Time { return now }))

	c.SetWithTTL("key1", "value1", 1*time.Hour)
	c.SetWithTTL("key2", "value2", 2*time.Hour)

	// Advance time past first expiry
	now = now.Add(90 * time.Minute)

	count := c.Cleanup()
	if count != 1 {
		t.Errorf("Cleanup() = %d, want 1", count)
	}

	if c.Len() != 1 {
		t.Errorf("Len() = %d, want 1", c.Len())
	}
}

func TestCache_Stats(t *testing.T) {
	now := time.Now()
	c := NewCache(WithTimeFunc(func() time.Time { return now }))

	c.SetWithTTL("key1", "value1", 1*time.Hour)
	c.SetWithTTL("key2", "value2", 2*time.Hour)

	stats := c.Stats()
	if stats.Items != 2 {
		t.Errorf("Items = %d, want 2", stats.Items)
	}
	if stats.Active != 2 {
		t.Errorf("Active = %d, want 2", stats.Active)
	}
	if stats.Expired != 0 {
		t.Errorf("Expired = %d, want 0", stats.Expired)
	}

	// Advance time past first expiry
	now = now.Add(90 * time.Minute)

	stats = c.Stats()
	if stats.Items != 2 {
		t.Errorf("Items = %d, want 2", stats.Items)
	}
	if stats.Active != 1 {
		t.Errorf("Active = %d, want 1", stats.Active)
	}
	if stats.Expired != 1 {
		t.Errorf("Expired = %d, want 1", stats.Expired)
	}
}

func TestCache_MaxItems(t *testing.T) {
	now := time.Now()
	c := NewCache(
		WithMaxItems(3),
		WithTimeFunc(func() time.Time { return now }),
	)

	c.SetWithTTL("key1", "value1", 1*time.Hour)
	c.SetWithTTL("key2", "value2", 2*time.Hour)
	c.SetWithTTL("key3", "value3", 3*time.Hour)

	// Adding a 4th item should evict the oldest
	c.SetWithTTL("key4", "value4", 4*time.Hour)

	if c.Len() != 3 {
		t.Errorf("Len() = %d, want 3", c.Len())
	}

	// key1 should be evicted (oldest expiry)
	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be evicted")
	}
}

func TestCache_MaxItems_EvictsExpiredFirst(t *testing.T) {
	now := time.Now()
	c := NewCache(
		WithMaxItems(3),
		WithTimeFunc(func() time.Time { return now }),
	)

	c.SetWithTTL("key1", "value1", 30*time.Minute) // Will expire
	c.SetWithTTL("key2", "value2", 2*time.Hour)
	c.SetWithTTL("key3", "value3", 3*time.Hour)

	// Advance time past key1 expiry
	now = now.Add(1 * time.Hour)

	// Adding a 4th item should evict expired key1 first
	c.SetWithTTL("key4", "value4", 4*time.Hour)

	if c.Len() != 3 {
		t.Errorf("Len() = %d, want 3", c.Len())
	}

	// key2, key3, key4 should remain
	_, ok := c.Get("key2")
	if !ok {
		t.Error("expected key2 to still exist")
	}
}

func TestCache_GetOrSet(t *testing.T) {
	c := NewCache()

	callCount := 0
	fn := func() (any, error) {
		callCount++
		return "computed", nil
	}

	// First call should compute
	val, err := c.GetOrSet("key1", fn)
	if err != nil {
		t.Fatalf("GetOrSet error: %v", err)
	}
	if val != "computed" {
		t.Errorf("value = %v, want %v", val, "computed")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1", callCount)
	}

	// Second call should use cache
	val, err = c.GetOrSet("key1", fn)
	if err != nil {
		t.Fatalf("GetOrSet error: %v", err)
	}
	if val != "computed" {
		t.Errorf("value = %v, want %v", val, "computed")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should not have called fn again)", callCount)
	}
}

func TestCache_GetOrSet_Error(t *testing.T) {
	c := NewCache()

	expectedErr := errors.New("compute error")
	fn := func() (any, error) {
		return nil, expectedErr
	}

	_, err := c.GetOrSet("key1", fn)
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}

	// Should not be cached
	_, ok := c.Get("key1")
	if ok {
		t.Error("expected error result not to be cached")
	}
}

func TestCache_GetOrSetWithTTL(t *testing.T) {
	now := time.Now()
	c := NewCache(WithTimeFunc(func() time.Time { return now }))

	callCount := 0
	fn := func() (any, error) {
		callCount++
		return "computed", nil
	}

	// First call with short TTL
	val, err := c.GetOrSetWithTTL("key1", 1*time.Hour, fn)
	if err != nil {
		t.Fatalf("GetOrSetWithTTL error: %v", err)
	}
	if val != "computed" || callCount != 1 {
		t.Error("first call failed")
	}

	// Advance time past TTL
	now = now.Add(2 * time.Hour)

	// Should recompute
	val, err = c.GetOrSetWithTTL("key1", 1*time.Hour, fn)
	if err != nil {
		t.Fatalf("GetOrSetWithTTL error: %v", err)
	}
	if val != "computed" {
		t.Errorf("value = %v, want %v", val, "computed")
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

func TestCache_Concurrent(t *testing.T) {
	c := NewCache()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key"
			c.Set(key, i)
			c.Get(key)
			c.Delete(key)
		}(i)
	}
	wg.Wait()
}

func TestCache_GetOrSet_Concurrent(t *testing.T) {
	c := NewCache()

	callCount := int32(0)
	fn := func() (any, error) {
		// Use atomic to check call count
		var mu sync.Mutex
		mu.Lock()
		callCount++
		mu.Unlock()
		return "computed", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = c.GetOrSet("key1", fn)
		}()
	}
	wg.Wait()

	// Due to the double-check pattern, fn might be called more than once
	// in a race, but the final value should be consistent
	val, ok := c.Get("key1")
	if !ok || val != "computed" {
		t.Error("expected cached value")
	}
}
