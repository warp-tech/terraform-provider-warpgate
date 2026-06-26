package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestBuildTargetDataRequestIncludesRateLimit(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{
		"name":                     "limited-target",
		"description":              "target with a bandwidth limit",
		"group_id":                 "2f474c68-9980-46ff-b2d3-c0b60f8b1193",
		rateLimitBytesPerSecondKey: 0,
		"ssh_options": []any{
			map[string]any{
				"host":                 "example.com",
				"port":                 22,
				"username":             "admin",
				"allow_insecure_algos": false,
				"public_key_auth":      []any{map[string]any{}},
			},
		},
	})

	req, err := buildTargetDataRequest(d)
	if err != nil {
		t.Fatalf("unexpected error building target request: %v", err)
	}

	if req.RateLimitBytesPerSecond == nil {
		t.Fatal("expected rate limit to be set")
	}

	if got, want := *req.RateLimitBytesPerSecond, 0; got != want {
		t.Fatalf("unexpected rate limit: got %d, want %d", got, want)
	}

	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("unexpected error marshaling target request: %v", err)
	}

	if !strings.Contains(string(payload), `"rate_limit_bytes_per_second":0`) {
		t.Fatalf("expected explicit zero rate limit in JSON payload, got %s", payload)
	}
}

func TestBuildUserUpdateRequestIncludesRateLimit(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]any{
		"username":                 "limited-user",
		"description":              "user with a bandwidth limit",
		rateLimitBytesPerSecondKey: 4096,
		"allowed_ip_ranges":        []any{"10.0.0.0/8"},
	})

	req := buildUserUpdateRequest(d)
	if req.RateLimitBytesPerSecond == nil {
		t.Fatal("expected rate limit to be set")
	}

	if got, want := *req.RateLimitBytesPerSecond, 4096; got != want {
		t.Fatalf("unexpected rate limit: got %d, want %d", got, want)
	}

	if req.AllowedIPRanges == nil || len(*req.AllowedIPRanges) != 1 || (*req.AllowedIPRanges)[0] != "10.0.0.0/8" {
		t.Fatalf("unexpected allowed IP ranges: %#v", req.AllowedIPRanges)
	}
}

func TestBuildUserUpdateRequestOmitsUnsetRateLimit(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]any{
		"username":    "unlimited-user",
		"description": "user without a bandwidth limit",
	})

	req := buildUserUpdateRequest(d)
	if req.RateLimitBytesPerSecond != nil {
		t.Fatalf("expected rate limit to be nil, got %d", *req.RateLimitBytesPerSecond)
	}
}
