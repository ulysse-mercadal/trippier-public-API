package byok_test

import (
	"context"
	"testing"

	"github.com/trippier/poi-api/internal/byok"
	"github.com/trippier/poi-api/pkg/types"
)

func TestTicketmasterAPIKey(t *testing.T) {
	ctx := context.WithValue(context.Background(), byok.TicketmasterKey, "tm-key-123")
	if got := byok.TicketmasterAPIKey(ctx); got != "tm-key-123" {
		t.Errorf("TicketmasterAPIKey = %q, want %q", got, "tm-key-123")
	}
}

func TestTicketmasterAPIKey_Missing(t *testing.T) {
	if got := byok.TicketmasterAPIKey(context.Background()); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestEventbriteToken(t *testing.T) {
	ctx := context.WithValue(context.Background(), byok.EventbriteKey, "eb-token-456")
	if got := byok.EventbriteToken(ctx); got != "eb-token-456" {
		t.Errorf("EventbriteToken = %q, want %q", got, "eb-token-456")
	}
}

func TestEventbriteToken_Missing(t *testing.T) {
	if got := byok.EventbriteToken(context.Background()); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestWithProviderKey_GetProviderKey(t *testing.T) {
	id := types.ProviderFoursquare
	ctx := byok.WithProviderKey(context.Background(), id, "fsq-key-789")
	if got := byok.GetProviderKey(ctx, id); got != "fsq-key-789" {
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

func TestGenericKeyDoesNotConflictWithLegacy(t *testing.T) {
	// Storing a key via the generic slot must not be retrievable via legacy accessors.
	ctx := byok.WithProviderKey(context.Background(), types.ProviderTicketmaster, "generic-tm")
	if got := byok.TicketmasterAPIKey(ctx); got != "" {
		t.Errorf("legacy TicketmasterAPIKey should be empty when only generic key is set, got %q", got)
	}
}
