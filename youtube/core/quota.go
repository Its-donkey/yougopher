package core

import (
	"sync"
	"time"
)

// QuotaCosts contains the quota cost for each YouTube API operation.
// See: https://developers.google.com/youtube/v3/determine_quota_cost
var QuotaCosts = map[string]int{
	// Live Chat
	"liveChatMessages.list":     5,
	"liveChatMessages.insert":   50,
	"liveChatMessages.delete":   50,
	"liveChatBans.insert":       50,
	"liveChatBans.delete":       50,
	"liveChatModerators.insert": 50,
	"liveChatModerators.delete": 50,

	// Data API - Read
	"videos.list":        1,
	"channels.list":      1,
	"playlists.list":     1,
	"playlistItems.list": 1,
	"subscriptions.list": 1,
	"comments.list":      1,
	"commentThreads.list": 1,

	// Data API - Search (expensive!)
	"search.list": 100,

	// Data API - Write
	"videos.insert":        1600, // Video upload
	"videos.update":        50,
	"videos.delete":        50,
	"playlists.insert":     50,
	"playlists.update":     50,
	"playlists.delete":     50,
	"playlistItems.insert": 50,
	"playlistItems.update": 50,
	"playlistItems.delete": 50,
	"comments.insert":      50,
	"comments.update":      50,
	"comments.delete":      50,
	"subscriptions.insert": 50,
	"subscriptions.delete": 50,

	// Live Streaming
	"liveBroadcasts.list":   1,
	"liveBroadcasts.insert": 50,
	"liveBroadcasts.update": 50,
	"liveBroadcasts.delete": 50,
	"liveStreams.list":      1,
	"liveStreams.insert":    50,
	"liveStreams.update":    50,
	"liveStreams.delete":    50,
}

// DefaultDailyQuota is the default daily quota for YouTube Data API projects.
const DefaultDailyQuota = 10000

// QuotaTracker tracks YouTube API quota usage.
// It is safe for concurrent use.
type QuotaTracker struct {
	mu            sync.RWMutex
	used          int
	limit         int
	resetAt       time.Time
	handlers      map[uint64]func(used, limit int)
	nextHandlerID uint64
}

// NewQuotaTracker creates a new QuotaTracker with the specified daily limit.
func NewQuotaTracker(limit int) *QuotaTracker {
	return &QuotaTracker{
		limit:    limit,
		resetAt:  nextPacificMidnight(),
		handlers: make(map[uint64]func(used, limit int)),
	}
}

// Add records quota usage for an operation.
// Returns the total used quota after this operation.
func (q *QuotaTracker) Add(operation string, count int) int {
	cost, ok := QuotaCosts[operation]
	if !ok {
		cost = 1 // Default cost for unknown operations
	}

	q.mu.Lock()
	q.checkReset()
	q.used += cost * count
	used, limit := q.used, q.limit
	// Snapshot handlers to call outside lock (prevents deadlock)
	handlers := make([]func(int, int), 0, len(q.handlers))
	for _, h := range q.handlers {
		handlers = append(handlers, h)
	}
	q.mu.Unlock()

	// Notify handlers outside lock
	for _, h := range handlers {
		h(used, limit)
	}

	return used
}

// AddCost records a specific quota cost.
// Returns the total used quota after this operation.
func (q *QuotaTracker) AddCost(cost int) int {
	q.mu.Lock()
	q.checkReset()
	q.used += cost
	used, limit := q.used, q.limit
	// Snapshot handlers to call outside lock (prevents deadlock)
	handlers := make([]func(int, int), 0, len(q.handlers))
	for _, h := range q.handlers {
		handlers = append(handlers, h)
	}
	q.mu.Unlock()

	// Notify handlers outside lock
	for _, h := range handlers {
		h(used, limit)
	}

	return used
}

// Used returns the current quota usage.
func (q *QuotaTracker) Used() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.used
}

// Limit returns the daily quota limit.
func (q *QuotaTracker) Limit() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.limit
}

// Remaining returns the remaining quota for today.
func (q *QuotaTracker) Remaining() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.checkReset()
	remaining := q.limit - q.used
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ResetAt returns when the quota will reset (Pacific midnight).
func (q *QuotaTracker) ResetAt() time.Time {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.resetAt
}

// IsExhausted returns true if the quota has been exhausted.
func (q *QuotaTracker) IsExhausted() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.used >= q.limit
}

// Reset manually resets the quota counter.
func (q *QuotaTracker) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.used = 0
	q.resetAt = nextPacificMidnight()
}

// OnUsageChange registers a callback for quota usage changes.
// Returns an unsubscribe function.
func (q *QuotaTracker) OnUsageChange(fn func(used, limit int)) func() {
	q.mu.Lock()
	defer q.mu.Unlock()

	id := q.nextHandlerID
	q.nextHandlerID++
	q.handlers[id] = fn

	return func() {
		q.mu.Lock()
		defer q.mu.Unlock()
		delete(q.handlers, id)
	}
}

// checkReset resets the counter if we've passed midnight Pacific.
// Must be called with lock held.
func (q *QuotaTracker) checkReset() {
	now := time.Now()
	if now.After(q.resetAt) {
		q.used = 0
		q.resetAt = nextPacificMidnight()
	}
}

// nextPacificMidnight returns the next midnight in Pacific Time.
// YouTube API quotas reset at midnight Pacific Time (America/Los_Angeles).
func nextPacificMidnight() time.Time {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		// Fallback if timezone data unavailable.
		// Use approximate DST rules for US Pacific Time:
		// DST is observed from second Sunday of March to first Sunday of November.
		// This is an approximation; for accurate timing, ensure tzdata is available.
		now := time.Now().UTC()
		month := now.Month()
		// PDT (UTC-7) roughly March-November, PST (UTC-8) otherwise
		offset := -8 * 60 * 60 // PST
		if month >= time.March && month < time.November {
			offset = -7 * 60 * 60 // PDT (approximate)
		}
		loc = time.FixedZone("Pacific", offset)
	}

	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	return midnight
}

// EstimateCost calculates the quota cost for a batch of operations.
func EstimateCost(operations map[string]int) int {
	total := 0
	for op, count := range operations {
		cost, ok := QuotaCosts[op]
		if !ok {
			cost = 1
		}
		total += cost * count
	}
	return total
}
