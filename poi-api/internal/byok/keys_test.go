package byok_test

import (
	"context"
	"testing"

	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/pkg/types"
)

func TestWithProviderKey_RoundTrip(t *testing.T) {
	ctx := byok.WithProviderKey(context.Background(), types.ProviderFoursquare, "fsq-key-789")
	if got := byok.GetProviderKey(ctx, types.ProviderFoursquare); got != "fsq-key-789" {
		t.Errorf("GetProviderKey = %q, want %q", got, "fsq-key-789")
	}
}

func TestGetProviderKey_Missing(t *testing.T) {
	if got := byok.GetProviderKey(context.Background(), types.ProviderFoursquare); got != "" {
		t.Errorf("expected empty string for missing key, got %q", got)
	}
}

func TestWithProviderKey_Isolation(t *testing.T) {
	ctx := byok.WithProviderKey(context.Background(), types.ProviderBaidu, "baidu-key")
	ctx = byok.WithProviderKey(ctx, types.ProviderAmap, "amap-key")

	if got := byok.GetProviderKey(ctx, types.ProviderBaidu); got != "baidu-key" {
		t.Errorf("Baidu key = %q, want %q", got, "baidu-key")
	}
	if got := byok.GetProviderKey(ctx, types.ProviderAmap); got != "amap-key" {
		t.Errorf("Amap key = %q, want %q", got, "amap-key")
	}
	if got := byok.GetProviderKey(ctx, types.ProviderKakao); got != "" {
		t.Errorf("Kakao key should be empty, got %q", got)
	}
}
