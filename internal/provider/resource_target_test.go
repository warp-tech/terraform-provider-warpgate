package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

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

func TestBuildSSHTargetOptionsWithKeyIDAndJumpHost(t *testing.T) {
	opts, err := buildSSHTargetOptions(map[string]any{
		"host":                 "ssh.example.com",
		"port":                 22,
		"username":             "root",
		"allow_insecure_algos": true,
		"jump_host":            "11111111-2222-3333-4444-555555555555",
		"public_key_auth": []any{
			map[string]any{
				"key_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			},
		},
	})
	if err != nil {
		t.Fatalf("buildSSHTargetOptions returned error: %v", err)
	}

	if opts.Host != "ssh.example.com" || opts.Port != 22 || opts.Username != "root" {
		t.Fatalf("unexpected target options: %+v", opts)
	}
	if opts.JumpHost != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("expected jump host ID, got %q", opts.JumpHost)
	}

	pkAuth, ok := opts.Auth.(*client.SSHTargetPublicKeyAuth)
	if !ok {
		t.Fatalf("expected SSHTargetPublicKeyAuth, got %T", opts.Auth)
	}
	if pkAuth.KeyID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("expected key ID aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee, got %q", pkAuth.KeyID)
	}
}

func TestSetTargetOptionsWithSSHKeyIDAndJumpHost(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	err := setTargetOptions(d, &client.TargetSSHOptions{
		Kind:               "Ssh",
		Host:               "ssh.example.com",
		Port:               22,
		Username:           "root",
		AllowInsecureAlgos: false,
		JumpHost:           "11111111-2222-3333-4444-555555555555",
		Auth: &client.SSHTargetPublicKeyAuth{
			Kind:  "PublicKey",
			KeyID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
	})
	if err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	sshOptsList := d.Get("ssh_options").([]any)
	if len(sshOptsList) != 1 {
		t.Fatalf("expected 1 ssh_options block, got %d", len(sshOptsList))
	}

	sshOpts := sshOptsList[0].(map[string]any)
	if got := sshOpts["jump_host"]; got != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("expected jump_host 11111111-2222-3333-4444-555555555555, got %v", got)
	}

	pkAuthList := sshOpts["public_key_auth"].([]any)
	if len(pkAuthList) != 1 {
		t.Fatalf("expected 1 public_key_auth block, got %d", len(pkAuthList))
	}
	pkAuth := pkAuthList[0].(map[string]any)
	if got := pkAuth["key_id"]; got != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("expected key_id aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee, got %v", got)
	}
}

