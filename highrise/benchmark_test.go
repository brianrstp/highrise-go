package highrise

import (
	"context"
	"encoding/json"
	"testing"
)

func BenchmarkChatRequestMarshal(b *testing.B) {
	req := ChatRequest{
		Type:    "ChatRequest",
		Message: "hello world this is a test message for benchmarking purposes",
		RID:     "test-rid-12345",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}

func BenchmarkGetRoomUsersRequestMarshal(b *testing.B) {
	req := GetRoomUsersRequest{
		Type: "GetRoomUsersRequest",
		RID:  "test-rid-12345",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}

func BenchmarkChatEventUnmarshal(b *testing.B) {
	data := []byte(`{"_type":"ChatEvent","user":{"id":"u1","username":"alice","position":{"x":1,"y":2,"z":3,"facing":"Front"}},"message":"hello","whisper":false}`)
	var ev ChatEvent
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &ev)
	}
}

func BenchmarkSessionUserEntryUnmarshal(b *testing.B) {
	data := []byte(`{"user":{"id":"u1","username":"alice"},"position":{"x":0,"y":1,"z":2,"facing":"Front"}}`)
	var entry SessionUserEntry
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &entry)
	}
}

func BenchmarkRateLimiterAcquire(b *testing.B) {
	rl := newRateLimiter()
	rl.apply(map[string][]any{
		"global_bot": {float64(100000000), float64(1)},
	})
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.acquire(ctx, "ChatRequest")
	}
}

func BenchmarkMiddlewareChain(b *testing.B) {
	var sum int
	mw1 := func(next func()) { sum += 1; next() }
	mw2 := func(next func()) { sum += 2; next() }
	mw3 := func(next func()) { sum += 3; next() }
	mws := []Middleware{mw1, mw2, mw3}

	fn := func() {}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := len(mws) - 1; j >= 0; j-- {
			prev := fn
			fn = func() { mws[j](prev) }
		}
		fn()
		_ = sum
	}
}

func BenchmarkSlidingWindow(b *testing.B) {
	sw := newSlidingWindow(1000, 1.0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sw.acquire()
	}
}

func BenchmarkMessageMediaRequestMarshal(b *testing.B) {
	req := MessageMediaRequest{
		Type: "MessageMediaRequest",
		Media: MessageMedia{
			Type: "image",
			URL:  "https://example.com/image.png",
		},
		RID: "test-rid-12345",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}
