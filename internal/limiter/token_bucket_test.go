package limiter_test

import (
	"context"
	"testing"
	"time"

	"rate-limiter/internal/limiter"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTokenBucketLimiter_Allow(t *testing.T) {
	// Start miniredis
	s := miniredis.RunT(t)

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	defer rdb.Close()

	// Initialize Limiter: Capacity 5, Rate 1 token/sec
	// We use a rate of 10 to make tests faster if we used smaller units,
	// but here we just sleep.
	capacity := 5
	rate := 1.0
	l := limiter.NewTokenBucketLimiter(rdb, capacity, rate)
	ctx := context.Background()
	key := "test_user_123"

	// 1. Consume 5 tokens (all should succeed)
	for i := 0; i < capacity; i++ {
		allowed, err := l.Allow(ctx, key)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if !allowed {
			t.Fatalf("iteration %d: request should be allowed", i)
		}
	}

	// 2. Consume 1 more (should fail)
	allowed, err := l.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("request exceeding capacity should be denied")
	}

	// 3. Wait for 2.1 seconds (should replenish 2 tokens)
	// We need enough time to be sure. 1.1s is risky if execution is slow.
	// But `now` in script is based on `time.Now()`.
	time.Sleep(2100 * time.Millisecond)

	// 4. Consume 1 (should succeed)
	allowed, err = l.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("request should be allowed after refill")
	}

	// 5. Consume 1 more (should succeed, since we waited 2s and got 2 tokens)
	allowed, err = l.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("second request should be allowed after refill")
	}

	// 6. Consume 1 more (should fail)
	allowed, err = l.Allow(ctx, key)
	if allowed {
		t.Fatal("request should be denied after consuming refilled tokens")
	}
}
