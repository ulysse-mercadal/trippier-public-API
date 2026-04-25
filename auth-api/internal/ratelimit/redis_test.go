package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	rl "github.com/trippier/auth-api/internal/ratelimit"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestSetTokens_And_GetUsage(t *testing.T) {
	rdb := newTestRedis(t)
	ctx := context.Background()

	userID := "abc123"
	if err := rl.SetTokens(ctx, rdb, userID, 100, time.Minute); err != nil {
		t.Fatalf("SetTokens: %v", err)
	}

	remaining, ttlSecs, err := rl.GetUsage(ctx, rdb, userID)
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if remaining != 100 {
		t.Errorf("remaining = %d, want 100", remaining)
	}
	if ttlSecs <= 0 {
		t.Errorf("ttl should be positive, got %d", ttlSecs)
	}
}

func TestGetUsage_NotFound(t *testing.T) {
	rdb := newTestRedis(t)
	ctx := context.Background()

	remaining, _, err := rl.GetUsage(ctx, rdb, "nonexistent")
	if err != nil {
		t.Fatalf("GetUsage: %v", err)
	}
	if remaining != -1 {
		t.Errorf("expected -1 for missing key, got %d", remaining)
	}
}

func TestDeduct_Success(t *testing.T) {
	rdb := newTestRedis(t)
	ctx := context.Background()

	userID := "key1"
	_ = rl.SetTokens(ctx, rdb, userID, 50, time.Minute)

	remaining, ttl, notFound, insufficient, err := rl.Deduct(ctx, rdb, userID, 10)
	if err != nil {
		t.Fatalf("Deduct: %v", err)
	}
	if notFound || insufficient {
		t.Fatalf("notFound=%v insufficient=%v, want both false", notFound, insufficient)
	}
	if remaining != 40 {
		t.Errorf("remaining = %d, want 40", remaining)
	}
	if ttl <= 0 {
		t.Errorf("ttl should be positive, got %d", ttl)
	}
}

func TestDeduct_Insufficient(t *testing.T) {
	rdb := newTestRedis(t)
	ctx := context.Background()

	userID := "key2"
	_ = rl.SetTokens(ctx, rdb, userID, 5, time.Minute)

	_, _, _, insufficient, err := rl.Deduct(ctx, rdb, userID, 10)
	if err != nil {
		t.Fatalf("Deduct: %v", err)
	}
	if !insufficient {
		t.Error("expected insufficient=true when cost > remaining")
	}
}

func TestDeduct_NotFound(t *testing.T) {
	rdb := newTestRedis(t)
	ctx := context.Background()

	_, _, notFound, _, err := rl.Deduct(ctx, rdb, "ghost-key", 1)
	if err != nil {
		t.Fatalf("Deduct: %v", err)
	}
	if !notFound {
		t.Error("expected notFound=true for missing key")
	}
}

func TestDeduct_Idempotent_Remaining(t *testing.T) {
	rdb := newTestRedis(t)
	ctx := context.Background()

	userID := "key3"
	_ = rl.SetTokens(ctx, rdb, userID, 100, time.Minute)

	for i := 0; i < 5; i++ {
		remaining, _, _, _, err := rl.Deduct(ctx, rdb, userID, 10)
		if err != nil {
			t.Fatalf("deduct %d: %v", i, err)
		}
		want := 100 - (i+1)*10
		if remaining != want {
			t.Errorf("step %d: remaining = %d, want %d", i, remaining, want)
		}
	}
}
