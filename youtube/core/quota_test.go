package core

import (
	"sync"
	"testing"
	"time"
)

func TestNewQuotaTracker(t *testing.T) {
	qt := NewQuotaTracker(10000)

	if qt.Limit() != 10000 {
		t.Errorf("Limit() = %d, want 10000", qt.Limit())
	}
	if qt.Used() != 0 {
		t.Errorf("Used() = %d, want 0", qt.Used())
	}
	if qt.Remaining() != 10000 {
		t.Errorf("Remaining() = %d, want 10000", qt.Remaining())
	}
}

func TestQuotaTracker_Add(t *testing.T) {
	qt := NewQuotaTracker(10000)

	// Add known operation
	used := qt.Add("videos.list", 1)
	if used != 1 {
		t.Errorf("Add() returned %d, want 1", used)
	}

	// Add multiple
	used = qt.Add("liveChatMessages.list", 5)
	// 1 + (5 * 5 quota cost) = 26
	if used != 26 {
		t.Errorf("Add() returned %d, want 26", used)
	}

	// Add unknown operation (defaults to cost 1)
	used = qt.Add("unknown.operation", 1)
	if used != 27 {
		t.Errorf("Add() with unknown op returned %d, want 27", used)
	}
}

func TestQuotaTracker_AddCost(t *testing.T) {
	qt := NewQuotaTracker(10000)

	qt.AddCost(100)
	if qt.Used() != 100 {
		t.Errorf("Used() = %d, want 100", qt.Used())
	}

	qt.AddCost(50)
	if qt.Used() != 150 {
		t.Errorf("Used() = %d, want 150", qt.Used())
	}
}

func TestQuotaTracker_Remaining(t *testing.T) {
	qt := NewQuotaTracker(100)

	qt.AddCost(30)
	if qt.Remaining() != 70 {
		t.Errorf("Remaining() = %d, want 70", qt.Remaining())
	}

	qt.AddCost(80)
	if qt.Remaining() != 0 {
		t.Errorf("Remaining() = %d, want 0 (not negative)", qt.Remaining())
	}
}

func TestQuotaTracker_IsExhausted(t *testing.T) {
	qt := NewQuotaTracker(100)

	if qt.IsExhausted() {
		t.Error("IsExhausted() should be false initially")
	}

	qt.AddCost(99)
	if qt.IsExhausted() {
		t.Error("IsExhausted() should be false at 99/100")
	}

	qt.AddCost(1)
	if !qt.IsExhausted() {
		t.Error("IsExhausted() should be true at 100/100")
	}

	qt.AddCost(1)
	if !qt.IsExhausted() {
		t.Error("IsExhausted() should be true when over limit")
	}
}

func TestQuotaTracker_Reset(t *testing.T) {
	qt := NewQuotaTracker(100)
	qt.AddCost(50)

	qt.Reset()

	if qt.Used() != 0 {
		t.Errorf("Used() after Reset() = %d, want 0", qt.Used())
	}
}

func TestQuotaTracker_ResetAt(t *testing.T) {
	qt := NewQuotaTracker(100)
	resetAt := qt.ResetAt()

	if resetAt.IsZero() {
		t.Error("ResetAt() should not be zero")
	}

	// Should be in the future
	if !resetAt.After(time.Now()) {
		t.Error("ResetAt() should be in the future")
	}
}

func TestQuotaTracker_OnUsageChange(t *testing.T) {
	qt := NewQuotaTracker(100)

	var lastUsed, lastLimit int
	unsub := qt.OnUsageChange(func(used, limit int) {
		lastUsed = used
		lastLimit = limit
	})

	qt.AddCost(25)
	if lastUsed != 25 || lastLimit != 100 {
		t.Errorf("Handler got (%d, %d), want (25, 100)", lastUsed, lastLimit)
	}

	qt.AddCost(10)
	if lastUsed != 35 {
		t.Errorf("Handler got used=%d, want 35", lastUsed)
	}

	// Unsubscribe
	unsub()

	// Handler should not be called after unsubscribe
	// (though it may still be in slice as nil)
	qt.AddCost(5)
	// No panic expected
}

func TestQuotaTracker_Concurrent(t *testing.T) {
	qt := NewQuotaTracker(100000)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				qt.AddCost(1)
				_ = qt.Used()
				_ = qt.Remaining()
			}
		}()
	}
	wg.Wait()

	if qt.Used() != 10000 {
		t.Errorf("Used() = %d, want 10000", qt.Used())
	}
}

func TestQuotaCosts(t *testing.T) {
	// Verify some known costs
	tests := []struct {
		op   string
		cost int
	}{
		{"videos.list", 1},
		{"search.list", 100},
		{"liveChatMessages.list", 5},
		{"liveChatMessages.insert", 50},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			cost, ok := QuotaCosts[tt.op]
			if !ok {
				t.Errorf("QuotaCosts missing %q", tt.op)
				return
			}
			if cost != tt.cost {
				t.Errorf("QuotaCosts[%q] = %d, want %d", tt.op, cost, tt.cost)
			}
		})
	}
}

func TestEstimateCost(t *testing.T) {
	ops := map[string]int{
		"videos.list":  10, // 10 * 1 = 10
		"search.list":  2,  // 2 * 100 = 200
		"unknown.op":   5,  // 5 * 1 = 5
	}

	got := EstimateCost(ops)
	want := 215

	if got != want {
		t.Errorf("EstimateCost() = %d, want %d", got, want)
	}
}

func TestNextPacificMidnight(t *testing.T) {
	midnight := nextPacificMidnight()

	// Should be in the future
	if !midnight.After(time.Now()) {
		t.Error("nextPacificMidnight() should be in the future")
	}

	// Should be at midnight (00:00:00)
	loc, _ := time.LoadLocation("America/Los_Angeles")
	if loc != nil {
		inPacific := midnight.In(loc)
		if inPacific.Hour() != 0 || inPacific.Minute() != 0 || inPacific.Second() != 0 {
			t.Errorf("nextPacificMidnight() not at midnight: %v", inPacific)
		}
	}
}
