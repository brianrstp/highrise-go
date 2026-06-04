package highrise

import (
	"context"
	"testing"
	"time"
)

func TestSlidingWindow_Acquire_UnderLimit(t *testing.T) {
	sw := newSlidingWindow(5, 1.0)
	for i := 0; i < 5; i++ {
		if wait := sw.acquire(); wait != 0 {
			t.Fatalf("expected no wait on attempt %d, got %v", i, wait)
		}
	}
}

func TestSlidingWindow_Acquire_OverLimit(t *testing.T) {
	sw := newSlidingWindow(2, 0.5)
	sw.acquire()
	sw.acquire()
	wait := sw.acquire()
	if wait <= 0 {
		t.Fatal("expected positive wait time when over limit")
	}
}

func TestSlidingWindow_Acquire_ResetsAfterWindow(t *testing.T) {
	sw := newSlidingWindow(1, 0.05)
	sw.acquire()
	time.Sleep(60 * time.Millisecond)
	if wait := sw.acquire(); wait != 0 {
		t.Fatal("expected window to reset")
	}
}

func TestSlidingWindow_ZeroLimit(t *testing.T) {
	sw := newSlidingWindow(0, 1.0)
	wait := sw.acquire()
	if wait <= 0 {
		t.Fatal("expected wait for zero limit")
	}
}

func TestRateLimiter_Apply(t *testing.T) {
	rl := newRateLimiter()
	raw := map[string][]any{
		"chat": {float64(10), float64(1.0)},
	}
	rl.apply(raw)

	if _, ok := rl.limits["chat"]; !ok {
		t.Fatal("rate limit not applied")
	}
}

func TestRateLimiter_Apply_Invalid(t *testing.T) {
	rl := newRateLimiter()
	raw := map[string][]any{
		"test": {float64(10)},           // missing window
		"test2": {"invalid", 1.0},       // invalid limit
	}
	rl.apply(raw)

	if len(rl.limits) != 0 {
		t.Fatal("expected no valid limits")
	}
}

func TestRateLimiter_Acquire_UnderLimit(t *testing.T) {
	rl := newRateLimiter()
	rl.apply(map[string][]any{
		"chat": {float64(5), float64(1.0)},
	})

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := rl.acquire(ctx, "ChatRequest"); err != nil {
			t.Fatalf("unexpected error on attempt %d: %v", i, err)
		}
	}
}

func TestRateLimiter_Acquire_Waits(t *testing.T) {
	rl := newRateLimiter()
	rl.apply(map[string][]any{
		"chat": {float64(1), float64(0.1)},
	})

	ctx := context.Background()
	rl.acquire(ctx, "ChatRequest")

	start := time.Now()
	err := rl.acquire(ctx, "ChatRequest")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 50*time.Millisecond {
		t.Fatal("expected rate limiter to wait")
	}
}

func TestRateLimiter_Acquire_CancelledContext(t *testing.T) {
	rl := newRateLimiter()
	rl.apply(map[string][]any{
		"chat": {float64(0), float64(1.0)},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.acquire(ctx, "ChatRequest")
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
}

func TestRateLimiter_Acquire_UnknownAction(t *testing.T) {
	rl := newRateLimiter()
	rl.apply(map[string][]any{
		"chat": {float64(5), float64(1.0)},
	})

	ctx := context.Background()
	if err := rl.acquire(ctx, "UnknownRequest"); err != nil {
		t.Fatal("unknown action should not be rate limited")
	}
}

func TestRateLimiter_GlobalLimit(t *testing.T) {
	rl := newRateLimiter()
	rl.apply(map[string][]any{
		"global_bot": {float64(1), float64(1.0)},
	})

	ctx := context.Background()
	rl.acquire(ctx, "ChatRequest")

	start := time.Now()
	rl.acquire(ctx, "AnyRequest")
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Fatal("expected global rate limit to apply")
	}
}
