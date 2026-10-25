package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func secretBackendResourceData(t *testing.T, raw map[string]any) *schema.ResourceData {
	t.Helper()
	base := map[string]any{
		"name":         "openbao",
		"backend_type": "openbao",
		"address":      "https://bao.example.com:8200",
	}
	for key, value := range raw {
		base[key] = value
	}
	return schema.TestResourceDataRaw(t, resourceSecretBackend().Schema, base)
}

func TestSecretBackendRequestWithTokenAuth(t *testing.T) {
	d := secretBackendResourceData(t, map[string]any{
		"token_auth": []any{map[string]any{"token": "s.secret"}},
	})

	req, err := secretBackendRequestFromResourceData(d)
	if err != nil {
		t.Fatalf("secretBackendRequestFromResourceData returned error: %v", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`"auth":{"method":"Token","token":"s.secret"}`,
		`"allowed_paths":[]`,
		`"tls_skip_verify":false`,
	} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("expected %s on the wire, got %s", want, body)
		}
	}
	if strings.Contains(string(body), `"namespace"`) {
		t.Fatalf("an unset namespace must not be sent, got %s", body)
	}
}

func TestSecretBackendRequestWithAppRoleAuth(t *testing.T) {
	d := secretBackendResourceData(t, map[string]any{
		"namespace":     "infra",
		"allowed_paths": []any{"ssh-warpgate/sign/deploy", "secret/warpgate"},
		"approle_auth": []any{map[string]any{
			"role_id":   "role-id",
			"secret_id": "secret-id",
			"mount":     "approle-warpgate",
		}},
	})

	req, err := secretBackendRequestFromResourceData(d)
	if err != nil {
		t.Fatalf("secretBackendRequestFromResourceData returned error: %v", err)
	}

	want := client.SecretBackendAuth{
		Method:   client.SecretBackendAuthAppRole,
		RoleID:   "role-id",
		SecretID: "secret-id",
		Mount:    "approle-warpgate",
	}
	if req.Auth != want {
		t.Fatalf("expected %+v, got %+v", want, req.Auth)
	}
	if req.Namespace != "infra" {
		t.Fatalf("expected namespace infra, got %q", req.Namespace)
	}
	if strings.Join(req.AllowedPaths, ",") != "ssh-warpgate/sign/deploy,secret/warpgate" {
		t.Fatalf("allowed paths lost their order: %v", req.AllowedPaths)
	}
}

func TestSecretBackendRequestWithKubernetesAuth(t *testing.T) {
	d := secretBackendResourceData(t, map[string]any{
		"kubernetes_auth": []any{map[string]any{"role": "warpgate"}},
	})

	req, err := secretBackendRequestFromResourceData(d)
	if err != nil {
		t.Fatalf("secretBackendRequestFromResourceData returned error: %v", err)
	}

	body, err := json.Marshal(req.Auth)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"method":"Kubernetes","role":"warpgate"}` {
		t.Fatalf("unexpected auth on the wire: %s", body)
	}
}

func TestSetSecretBackendDataKeepsSecretsFromState(t *testing.T) {
	tests := []struct {
		name  string
		raw   map[string]any
		auth  client.SecretBackendAuth
		key   string
		value string
	}{
		{
			name: "token",
			raw:  map[string]any{"token_auth": []any{map[string]any{"token": "s.secret"}}},
			auth: client.SecretBackendAuth{Method: client.SecretBackendAuthToken},
			key:  "token_auth.0.token", value: "s.secret",
		},
		{
			name: "approle",
			raw: map[string]any{"approle_auth": []any{map[string]any{
				"role_id":   "role-id",
				"secret_id": "secret-id",
			}}},
			auth: client.SecretBackendAuth{Method: client.SecretBackendAuthAppRole, RoleID: "role-id"},
			key:  "approle_auth.0.secret_id", value: "secret-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := secretBackendResourceData(t, tt.raw)
			err := setSecretBackendData(d, &client.SecretBackend{
				Name:         "openbao",
				BackendType:  "openbao",
				Address:      "https://bao.example.com:8200",
				Auth:         tt.auth,
				AllowedPaths: []string{"ssh-warpgate/sign/deploy"},
			})
			if err != nil {
				t.Fatalf("setSecretBackendData returned error: %v", err)
			}

			if got := d.Get(tt.key).(string); got != tt.value {
				t.Fatalf("expected %s to keep %q, got %q", tt.key, tt.value, got)
			}
			if got := d.Get("allowed_paths.0").(string); got != "ssh-warpgate/sign/deploy" {
				t.Fatalf("expected allowed path, got %q", got)
			}
		})
	}
}

func TestSetSecretBackendDataSwitchesAuthBlock(t *testing.T) {
	d := secretBackendResourceData(t, map[string]any{
		"token_auth": []any{map[string]any{"token": "s.secret"}},
	})

	err := setSecretBackendData(d, &client.SecretBackend{
		Name:        "openbao",
		BackendType: "openbao",
		Address:     "https://bao.example.com:8200",
		Auth:        client.SecretBackendAuth{Method: client.SecretBackendAuthKubernetes, Role: "warpgate", Mount: "k8s"},
	})
	if err != nil {
		t.Fatalf("setSecretBackendData returned error: %v", err)
	}

	if got := d.Get("token_auth").([]any); len(got) != 0 {
		t.Fatalf("expected token_auth to be cleared, got %v", got)
	}
	if got := d.Get("kubernetes_auth.0.mount").(string); got != "k8s" {
		t.Fatalf("expected kubernetes mount k8s, got %q", got)
	}
}

func TestSetSecretBackendDataRejectsUnknownAuthMethod(t *testing.T) {
	d := secretBackendResourceData(t, map[string]any{
		"token_auth": []any{map[string]any{"token": "s.secret"}},
	})

	err := setSecretBackendData(d, &client.SecretBackend{
		Name: "openbao",
		Auth: client.SecretBackendAuth{Method: "Jwt"},
	})
	if err == nil || !strings.Contains(err.Error(), "Jwt") {
		t.Fatalf("expected an error naming the unknown method, got %v", err)
	}
}

func TestSecretBackendSchemaValidation(t *testing.T) {
	s := resourceSecretBackend().Schema

	for value, valid := range map[string]bool{
		"openbao":      true,
		"vault.prod_1": true,
		"has space":    false,
		"slash/name":   false,
		"":             false,
	} {
		_, errs := s["name"].ValidateFunc(value, "name")
		if (len(errs) == 0) != valid {
			t.Fatalf("name %q: expected valid=%v, got errors %v", value, valid, errs)
		}
	}

	if _, errs := s["backend_type"].ValidateFunc("consul", "backend_type"); len(errs) == 0 {
		t.Fatal("expected backend_type consul to be rejected")
	}
}
