package highrise

import (
	"context"
	"strings"
	"sync"
	"time"
)

type slidingWindow struct {
	limit  int
	window time.Duration
	count  int
	start  time.Time
}

func newSlidingWindow(limit int, windowSec float64) *slidingWindow {
	return &slidingWindow{
		limit:  limit,
		window: time.Duration(windowSec * float64(time.Second)),
		start:  time.Now(),
	}
}

func (w *slidingWindow) acquire() time.Duration {
	now := time.Now()
	if now.Sub(w.start) >= w.window {
		w.count = 0
		w.start = now
	}
	if w.count < w.limit {
		w.count++
		return 0
	}
	return w.window - now.Sub(w.start)
}

type rateLimiter struct {
	mu     sync.Mutex
	limits map[string]*slidingWindow
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		limits: make(map[string]*slidingWindow),
	}
}

func (rl *rateLimiter) apply(raw map[string][]any) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for key, val := range raw {
		if len(val) < 2 {
			continue
		}
		limit, ok1 := val[0].(float64)
		window, ok2 := val[1].(float64)
		if !ok1 || !ok2 {
			continue
		}
		rl.limits[key] = newSlidingWindow(int(limit), window)
	}
}

func (rl *rateLimiter) acquire(ctx context.Context, actionType string) error {
	key := strings.TrimSuffix(actionType, "Request")

	rl.mu.Lock()
	var maxWait time.Duration
	for name, sw := range rl.limits {
		if name != "global_bot" && !strings.EqualFold(name, key) {
			continue
		}
		if wait := sw.acquire(); wait > maxWait {
			maxWait = wait
		}
	}
	rl.mu.Unlock()

	if maxWait > 0 {
		timer := time.NewTimer(maxWait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}
