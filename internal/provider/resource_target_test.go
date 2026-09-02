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

func TestBuildSSHTargetOptionsWithEmptyPublicKeyAuthBlock(t *testing.T) {
	opts, err := buildSSHTargetOptions(map[string]any{
		"host":                 "ssh.example.com",
		"port":                 22,
		"username":             "root",
		"allow_insecure_algos": false,
		"public_key_auth":      []any{nil},
	})
	if err != nil {
		t.Fatalf("buildSSHTargetOptions returned error: %v", err)
	}

	pkAuth, ok := opts.Auth.(*client.SSHTargetPublicKeyAuth)
	if !ok {
		t.Fatalf("expected SSHTargetPublicKeyAuth, got %T", opts.Auth)
	}
	if pkAuth.KeyID != "" {
		t.Fatalf("expected empty key ID, got %q", pkAuth.KeyID)
	}
}
