package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func TestBuildPostgresTargetOptionsWithProtocolVersion(t *testing.T) {
	opts, err := buildPostgresTargetOptions(map[string]any{
		"host":             "postgres.example.com",
		"port":             5432,
		"username":         "admin",
		"protocol_version": "3.0",
		"password":         "secret",
		"tls": []any{
			map[string]any{
				"mode":   "Required",
				"verify": true,
			},
		},
	})
	if err != nil {
		t.Fatalf("buildPostgresTargetOptions returned error: %v", err)
	}

	if opts.ProtocolVersion != "3.0" {
		t.Fatalf("expected protocol version 3.0, got %q", opts.ProtocolVersion)
	}
}

func TestSetTargetOptionsWithPostgresProtocolVersion(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	err := setTargetOptions(d, &client.TargetPostgresOptions{
		Kind:            "Postgres",
		Host:            "postgres.example.com",
		Port:            5432,
		Username:        "admin",
		ProtocolVersion: "3.2",
		Password:        "secret",
		TLS: client.TLS{
			Mode:   client.TLSModeRequired,
			Verify: true,
		},
	})
	if err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	postgresOptions := d.Get("postgres_options").([]any)
	if len(postgresOptions) != 1 {
		t.Fatalf("expected one postgres_options block, got %d", len(postgresOptions))
	}

	opts := postgresOptions[0].(map[string]any)
	if got := opts["protocol_version"]; got != "3.2" {
		t.Fatalf("expected protocol version 3.2, got %v", got)
	}
}
