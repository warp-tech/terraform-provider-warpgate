package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func TestBuildRDPTargetOptions(t *testing.T) {
	opts, err := buildRDPTargetOptions(map[string]any{
		"host":         "rdp.example.com",
		"port":         3389,
		"username":     "admin",
		"password":     "secret",
		"domain":       "WORKGROUP",
		"verify_tls":   false,
		"tls_security": "Tls12",
	})
	if err != nil {
		t.Fatalf("buildRDPTargetOptions returned error: %v", err)
	}

	if opts.Host != "rdp.example.com" {
		t.Fatalf("expected host rdp.example.com, got %q", opts.Host)
	}

	if opts.Port != 3389 {
		t.Fatalf("expected port 3389, got %d", opts.Port)
	}

	if opts.Username != "admin" {
		t.Fatalf("expected username admin, got %q", opts.Username)
	}

	if opts.Domain != "WORKGROUP" {
		t.Fatalf("expected domain WORKGROUP, got %q", opts.Domain)
	}

	if opts.VerifyTLS {
		t.Fatal("expected verify_tls to be false")
	}

	if opts.TLSSecurity != "Tls12" {
		t.Fatalf("expected tls_security Tls12, got %q", opts.TLSSecurity)
	}

	passwordAuth, ok := opts.Auth.(*client.RDPTargetPasswordAuth)
	if !ok {
		t.Fatalf("expected RDPTargetPasswordAuth, got %T", opts.Auth)
	}

	if passwordAuth.Password != "secret" {
		t.Fatalf("expected password secret, got %q", passwordAuth.Password)
	}
}

func TestSetTargetOptionsWithRDP(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	err := setTargetOptions(d, &client.TargetRDPOptions{
		Kind:        "Rdp",
		Host:        "rdp.example.com",
		Port:        3389,
		Username:    "admin",
		Domain:      "WORKGROUP",
		VerifyTLS:   false,
		TLSSecurity: "Tls12",
		Auth: &client.RDPTargetPasswordAuth{
			Kind:     "password",
			Password: "secret",
		},
	})
	if err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	rdpOptions := d.Get("rdp_options").([]any)
	if len(rdpOptions) != 1 {
		t.Fatalf("expected one rdp_options block, got %d", len(rdpOptions))
	}

	opts := rdpOptions[0].(map[string]any)
	if got := opts["host"]; got != "rdp.example.com" {
		t.Fatalf("expected host rdp.example.com, got %v", got)
	}

	if got := opts["port"]; got != 3389 {
		t.Fatalf("expected port 3389, got %v", got)
	}

	if got := opts["username"]; got != "admin" {
		t.Fatalf("expected username admin, got %v", got)
	}

	if got := opts["domain"]; got != "WORKGROUP" {
		t.Fatalf("expected domain WORKGROUP, got %v", got)
	}

	if got := opts["password"]; got != "secret" {
		t.Fatalf("expected password secret, got %v", got)
	}

	if got := opts["verify_tls"]; got != false {
		t.Fatalf("expected verify_tls false, got %v", got)
	}

	if got := opts["tls_security"]; got != "Tls12" {
		t.Fatalf("expected tls_security Tls12, got %v", got)
	}
}

func TestBuildPostgresTargetOptionsWithProtocolVersion(t *testing.T) {
	opts, err := buildPostgresTargetOptions(map[string]any{
		"host":                  "postgres.example.com",
		"port":                  5432,
		"username":              "admin",
		"default_database_name": "app_db",
		"protocol_version":      "3.0",
		"password":              "secret",
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

	if opts.DefaultDatabaseName != "app_db" {
		t.Fatalf("expected default database name app_db, got %q", opts.DefaultDatabaseName)
	}
}

func TestSetTargetOptionsWithPostgresProtocolVersion(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	err := setTargetOptions(d, &client.TargetPostgresOptions{
		Kind:                "Postgres",
		Host:                "postgres.example.com",
		Port:                5432,
		Username:            "admin",
		DefaultDatabaseName: "app_db",
		ProtocolVersion:     "3.2",
		Password:            "secret",
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
	if got := opts["default_database_name"]; got != "app_db" {
		t.Fatalf("expected default database name app_db, got %v", got)
	}

	if got := opts["protocol_version"]; got != "3.2" {
		t.Fatalf("expected protocol version 3.2, got %v", got)
	}
}
