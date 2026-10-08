package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProviderHeadersConvertsTheMap(t *testing.T) {
	got, err := providerHeaders(map[string]any{"CF-Access-Client-Id": "client-id"})
	if err != nil {
		t.Fatal(err)
	}
	if got["CF-Access-Client-Id"] != "client-id" || len(got) != 1 {
		t.Fatalf("unexpected headers: %v", got)
	}
}

func TestProviderHeadersRejectsHeadersTheClientNeverSends(t *testing.T) {
	for _, name := range []string{"Host", "host", "Content-Length", "transfer-encoding", "Trailer"} {
		if _, err := providerHeaders(map[string]any{name: "x"}); err == nil {
			t.Errorf("expected %q to be rejected", name)
		}
	}
}

func TestConfigureReadsHeaders(t *testing.T) {
	p := New("test")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"host":    "https://warpgate.example.com",
		"headers": map[string]any{"Host": "other.example.com"},
	})
	if _, diags := configure()(context.Background(), d); !diags.HasError() {
		t.Fatal("expected configure to reject a Host header")
	}

	d = schema.TestResourceDataRaw(t, p.Schema, map[string]any{
		"host":    "https://warpgate.example.com",
		"headers": map[string]any{"CF-Access-Client-Id": "client-id"},
	})
	if _, diags := configure()(context.Background(), d); diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
}
