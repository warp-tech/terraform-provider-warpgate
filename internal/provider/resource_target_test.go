package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func TestBuildRDPTargetOptions(t *testing.T) {
	opts, err := buildRDPTargetOptions(map[string]any{
		"host":              "rdp.example.com",
		"port":              3389,
		"username":          "admin",
		"password":          "secret",
		"domain":            "WORKGROUP",
		"verify_tls":        false,
		"tls_security":      "Tls12",
		"compression":       "lossless",
		"interactive_logon": true,
	})
	if err != nil {
		t.Fatalf("buildRDPTargetOptions returned error: %v", err)
	}

	if opts.Compression != "lossless" || !opts.InteractiveLogon {
		t.Fatalf("expected compression lossless and interactive logon, got %+v", opts)
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
		Kind:             "Rdp",
		Host:             "rdp.example.com",
		Port:             3389,
		Username:         "admin",
		Domain:           "WORKGROUP",
		VerifyTLS:        false,
		TLSSecurity:      "Tls12",
		Compression:      "remotefx",
		InteractiveLogon: false,
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

	if got := opts["compression"]; got != "remotefx" {
		t.Fatalf("expected compression remotefx, got %v", got)
	}

	if got := opts["interactive_logon"]; got != false {
		t.Fatalf("expected interactive_logon false, got %v", got)
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

	auth, ok := opts.Auth.(*client.DatabaseTargetPasswordAuth)
	if !ok {
		t.Fatalf("expected DatabaseTargetPasswordAuth, got %T", opts.Auth)
	}

	if auth.Kind != "Password" || auth.Password != "secret" {
		t.Fatalf("expected a Password auth carrying the password, got %+v", auth)
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
		Auth: &client.DatabaseTargetPasswordAuth{
			Kind:     "Password",
			Password: "secret",
		},
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

	if got := opts["password"]; got != "secret" {
		t.Fatalf("expected the password out of the auth block, got %v", got)
	}
}

func TestBuildMysqlTargetOptionsIamRole(t *testing.T) {
	opts, err := buildMysqlTargetOptions(map[string]any{
		"host":          "mysql.example.com",
		"port":          3306,
		"username":      "app",
		"iam_role_auth": []any{map[string]any{}},
		"tls": []any{
			map[string]any{"mode": "Required", "verify": true},
		},
	})
	if err != nil {
		t.Fatalf("buildMysqlTargetOptions returned error: %v", err)
	}

	if _, ok := opts.Auth.(*client.DatabaseTargetIamRoleAuth); !ok {
		t.Fatalf("expected DatabaseTargetIamRoleAuth, got %T", opts.Auth)
	}
}

func TestSetTargetOptionsWithMysqlIamRole(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	err := setTargetOptions(d, &client.TargetMySQLOptions{
		Kind:     "MySql",
		Host:     "mysql.example.com",
		Port:     3306,
		Username: "app",
		Auth:     &client.DatabaseTargetIamRoleAuth{Kind: "IamRole"},
		TLS:      client.TLS{Mode: client.TLSModeRequired, Verify: true},
	})
	if err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	opts := d.Get("mysql_options").([]any)[0].(map[string]any)
	if got := opts["iam_role_auth"].([]any); len(got) != 1 {
		t.Fatalf("expected one iam_role_auth block, got %v", got)
	}

	if got := opts["password"]; got != "" {
		t.Fatalf("expected no password for an IAM target, got %v", got)
	}
}

// The API requires the headers map, and Go marshals a nil map as null, which
// it rejects: an HTTP target with no headers must still send `{}`.
func TestBuildHTTPTargetOptionsSendsEmptyHeaders(t *testing.T) {
	opts, err := buildHTTPTargetOptions(map[string]any{
		"url": "http://app.internal",
		"tls": []any{map[string]any{"mode": "Disabled", "verify": false}},
	})
	if err != nil {
		t.Fatalf("buildHTTPTargetOptions returned error: %v", err)
	}

	body, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), `"headers":{}`) {
		t.Fatalf("expected an empty headers object on the wire, got %s", body)
	}
}

// Warpgate refuses a target write that leaves an approval gate unstated, so
// each flag must be present in the body even when the config never set it.
func TestBuildTargetDataRequestStatesEveryApprovalGate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{
		"name": "gated",
		"ssh_options": []any{
			map[string]any{
				"host":            "ssh.example.com",
				"port":            22,
				"username":        "root",
				"public_key_auth": []any{map[string]any{}},
			},
		},
	})

	req, err := buildTargetDataRequest(d)
	if err != nil {
		t.Fatalf("buildTargetDataRequest returned error: %v", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	for _, gate := range []string{"require_approval", "ticket_requests_disabled", "ticket_require_approval"} {
		if !strings.Contains(string(body), `"`+gate+`":false`) {
			t.Fatalf("expected %s to be sent explicitly, got %s", gate, body)
		}
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

func TestVNCTargetOptionsRoundTrip(t *testing.T) {
	opts := buildVNCTargetOptions(map[string]any{"host": "vnc.example.com", "port": 5901, "password": "secret"})
	if opts.Kind != "Vnc" || opts.Host != "vnc.example.com" || opts.Port != 5901 {
		t.Fatalf("unexpected options %+v", opts)
	}

	if auth, ok := opts.Auth.(*client.VncTargetPasswordAuth); !ok || auth.Kind != "Password" || auth.Password != "secret" {
		t.Fatalf("expected password auth, got %#v", opts.Auth)
	}

	if auth, ok := buildVNCTargetOptions(map[string]any{"host": "h", "port": 5900, "password": ""}).Auth.(*client.VncTargetNoneAuth); !ok || auth.Kind != "None" {
		t.Fatalf("expected None auth for an empty password")
	}

	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	if err := setTargetOptions(d, opts); err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	block := d.Get("vnc_options").([]any)[0].(map[string]any)
	if block["host"] != "vnc.example.com" || block["port"] != 5901 || block["password"] != "secret" {
		t.Fatalf("unexpected vnc_options block %v", block)
	}
}

func TestKubernetesTargetOptionsIamRoleRoundTrip(t *testing.T) {
	opts, err := buildKubernetesTargetOptions(map[string]any{
		"cluster_url":   "https://k8s.example.com",
		"tls":           []any{map[string]any{"mode": "Required", "verify": true}},
		"iam_role_auth": []any{map[string]any{}},
	})
	if err != nil {
		t.Fatalf("buildKubernetesTargetOptions returned error: %v", err)
	}

	if auth, ok := opts.Auth.(*client.KubernetesTargetIamRoleAuth); !ok || auth.Kind != "IamRole" {
		t.Fatalf("expected IamRole auth, got %#v", opts.Auth)
	}

	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	if err := setTargetOptions(d, opts); err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	block := d.Get("kubernetes_options").([]any)[0].(map[string]any)
	if got := block["iam_role_auth"].([]any); len(got) != 1 {
		t.Fatalf("expected an iam_role_auth block, got %v", block)
	}
}

func TestPostgresTargetOptionsIdleTimeoutRoundTrip(t *testing.T) {
	opts, err := buildPostgresTargetOptions(map[string]any{
		"host":         "pg.example.com",
		"port":         5432,
		"username":     "app",
		"password":     "secret",
		"idle_timeout": "30m",
		"tls":          []any{map[string]any{"mode": "Preferred", "verify": false}},
	})
	if err != nil {
		t.Fatalf("buildPostgresTargetOptions returned error: %v", err)
	}

	if opts.IdleTimeout != "30m" {
		t.Fatalf("expected idle_timeout 30m, got %q", opts.IdleTimeout)
	}

	d := schema.TestResourceDataRaw(t, resourceTarget().Schema, map[string]any{})
	if err := setTargetOptions(d, opts); err != nil {
		t.Fatalf("setTargetOptions returned error: %v", err)
	}

	if got := d.Get("postgres_options.0.idle_timeout"); got != "30m" {
		t.Fatalf("expected idle_timeout 30m in state, got %v", got)
	}
}
