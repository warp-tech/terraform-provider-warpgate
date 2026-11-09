package provider

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func stringPtr(v string) *string { return &v }

func boolPtr(v bool) *bool { return &v }

func TestOpenTargetsInNewTabValidation(t *testing.T) {
	for _, mode := range []string{"DefaultOn", "DefaultOff", "ForcedOn", "ForcedOff"} {
		diags := resourceParameters().Validate(terraform.NewResourceConfigRaw(map[string]any{
			"allow_own_credential_management": true,
			"open_targets_in_new_tab":         mode,
		}))
		if diags.HasError() {
			t.Fatalf("expected %q to validate, got %+v", mode, diags)
		}
	}

	diags := resourceParameters().Validate(terraform.NewResourceConfigRaw(map[string]any{
		"allow_own_credential_management": true,
		"open_targets_in_new_tab":         "Always",
	}))
	if !diags.HasError() {
		t.Fatal("expected an error for an invalid open_targets_in_new_tab value")
	}
}

func TestExpandRecordingsStorageDisk(t *testing.T) {
	cfg := buildRecordingsStorage(map[string]any{
		"disk": []any{
			map[string]any{"path": "/var/lib/warpgate/recordings"},
		},
	})
	if cfg == nil {
		t.Fatal("buildRecordingsStorage returned nil")
	}

	if cfg.Kind != "Disk" {
		t.Fatalf("expected kind Disk, got %q", cfg.Kind)
	}

	if cfg.Path == nil || *cfg.Path != "/var/lib/warpgate/recordings" {
		t.Fatalf("expected the configured path, got %v", cfg.Path)
	}

	// Sending an S3 field alongside Disk is rejected by the API.
	if cfg.Bucket != nil || cfg.Region != nil || cfg.PathStyle != nil || cfg.Prefix != nil || cfg.Credentials != nil {
		t.Fatalf("expected no S3 fields on a Disk config, got %+v", cfg)
	}
}

func TestExpandRecordingsStorageS3StaticCredentials(t *testing.T) {
	cfg := buildRecordingsStorage(map[string]any{
		"s3": []any{
			map[string]any{
				"bucket":     "recordings",
				"region":     "us-east-1",
				"endpoint":   "http://minio:9000",
				"path_style": true,
				"prefix":     "",
				"static_credentials": []any{
					map[string]any{
						"access_key_id":     "AKIA",
						"secret_access_key": "shhh",
					},
				},
			},
		},
	})
	if cfg == nil {
		t.Fatal("buildRecordingsStorage returned nil")
	}

	if cfg.Kind != "S3" {
		t.Fatalf("expected kind S3, got %q", cfg.Kind)
	}

	if cfg.Path != nil {
		t.Fatalf("expected no Disk path on an S3 config, got %v", cfg.Path)
	}

	// path_style and prefix are required by the API and legitimately false and
	// "", so they must be sent rather than dropped as empty values.
	if cfg.PathStyle == nil || !*cfg.PathStyle {
		t.Fatalf("expected path_style true, got %v", cfg.PathStyle)
	}

	if cfg.Prefix == nil || *cfg.Prefix != "" {
		t.Fatalf("expected an empty prefix to be sent, got %v", cfg.Prefix)
	}

	if cfg.Credentials == nil || cfg.Credentials.Mode != "Static" {
		t.Fatalf("expected Static credentials, got %+v", cfg.Credentials)
	}

	if cfg.Credentials.SecretAccessKey == nil || *cfg.Credentials.SecretAccessKey != "shhh" {
		t.Fatalf("expected the configured secret, got %v", cfg.Credentials.SecretAccessKey)
	}
}

func TestExpandRecordingsStorageS3OmitsUnsetSecret(t *testing.T) {
	cfg := buildRecordingsStorage(map[string]any{
		"s3": []any{
			map[string]any{
				"bucket": "recordings",
				"region": "us-east-1",
				"static_credentials": []any{
					map[string]any{"access_key_id": "AKIA"},
				},
			},
		},
	})
	if cfg == nil || cfg.Credentials == nil {
		t.Fatal("expected S3 credentials")
	}

	// Omitting the secret is how the caller says "keep the stored one".
	if cfg.Credentials.SecretAccessKey != nil {
		t.Fatalf("expected no secret to be sent, got %q", *cfg.Credentials.SecretAccessKey)
	}
}

func TestExpandRecordingsStorageS3AutoCredentials(t *testing.T) {
	cfg := buildRecordingsStorage(map[string]any{
		"s3": []any{
			map[string]any{
				"bucket":           "recordings",
				"region":           "us-east-1",
				"auto_credentials": []any{map[string]any{}},
			},
		},
	})
	if cfg == nil || cfg.Credentials == nil {
		t.Fatal("expected S3 credentials")
	}

	if cfg.Credentials.Mode != "Auto" {
		t.Fatalf("expected Auto credentials, got %q", cfg.Credentials.Mode)
	}

	if cfg.Credentials.AccessKeyID != nil {
		t.Fatalf("expected no access key id, got %q", *cfg.Credentials.AccessKeyID)
	}
}

func TestExpandRecordingsStorageUnsetIsNil(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{})

	// A parameters resource that does not manage storage must not send the
	// field at all: the update is partial, so an empty value would overwrite.
	if cfg := expandRecordingsStorage(d); cfg != nil {
		t.Fatalf("expected nil for unconfigured storage, got %+v", cfg)
	}
}

func TestFlattenRecordingsStorageDisk(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{})

	got := flattenRecordingsStorage(d, client.RecordingsStorageConfig{
		Kind: "Disk",
		Path: stringPtr("/var/lib/warpgate/recordings"),
	})

	storage := got[0].(map[string]any)

	disk, ok := storage["disk"].([]any)
	if !ok || len(disk) != 1 {
		t.Fatalf("expected one disk block, got %+v", storage)
	}

	if path := disk[0].(map[string]any)["path"]; path != "/var/lib/warpgate/recordings" {
		t.Fatalf("expected the API path, got %v", path)
	}

	if _, ok := storage["s3"]; ok {
		t.Fatalf("expected no s3 block on a Disk config, got %+v", storage)
	}
}

func TestFlattenRecordingsStorageS3KeepsConfiguredSecret(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{
		"recordings_storage": []any{
			map[string]any{
				"s3": []any{
					map[string]any{
						"bucket": "recordings",
						"region": "us-east-1",
						"static_credentials": []any{
							map[string]any{
								"access_key_id":     "AKIA",
								"secret_access_key": "shhh",
							},
						},
					},
				},
			},
		},
	})

	// The API redacts the secret, so a read that echoed the response would wipe
	// it from state and show a spurious diff on every plan.
	got := flattenRecordingsStorage(d, client.RecordingsStorageConfig{
		Kind:      "S3",
		Bucket:    stringPtr("recordings"),
		Region:    stringPtr("us-east-1"),
		PathStyle: boolPtr(false),
		Prefix:    stringPtr(""),
		Credentials: &client.S3Credentials{
			Mode:        "Static",
			AccessKeyID: stringPtr("AKIA"),
		},
	})

	s3 := got[0].(map[string]any)["s3"].([]any)[0].(map[string]any)
	static := s3["static_credentials"].([]any)[0].(map[string]any)

	if static["secret_access_key"] != "shhh" {
		t.Fatalf("expected the configured secret to survive a read, got %v", static["secret_access_key"])
	}
}

func TestFlattenRecordingsStorageUnknownKind(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{})

	if got := flattenRecordingsStorage(d, client.RecordingsStorageConfig{Kind: "Tape"}); got != nil {
		t.Fatalf("expected nil for an unknown kind, got %+v", got)
	}
}

func validateParameters(t *testing.T, storage []any) diag.Diagnostics {
	t.Helper()

	raw := map[string]any{"allow_own_credential_management": true}
	if storage != nil {
		raw["recordings_storage"] = storage
	}

	return resourceParameters().Validate(terraform.NewResourceConfigRaw(raw))
}

func TestValidateRecordingsStorageRequiresExactlyOneKind(t *testing.T) {
	// Neither disk nor s3: caught at validate rather than as an opaque
	// apply-time API error.
	if diags := validateParameters(t, []any{map[string]any{}}); !diags.HasError() {
		t.Fatal("expected an error for a recordings_storage block with no storage kind")
	}

	if diags := validateParameters(t, []any{
		map[string]any{
			"disk": []any{map[string]any{"path": "/rec"}},
			"s3": []any{map[string]any{
				"bucket":           "b",
				"region":           "r",
				"auto_credentials": []any{map[string]any{}},
			}},
		},
	}); !diags.HasError() {
		t.Fatal("expected an error for both disk and s3")
	}

	if diags := validateParameters(t, []any{
		map[string]any{"disk": []any{map[string]any{"path": "/rec"}}},
	}); diags.HasError() {
		t.Fatalf("expected a disk-only config to validate, got %+v", diags)
	}
}

func TestValidateS3RequiresExactlyOneCredentialsMode(t *testing.T) {
	s3 := func(creds map[string]any) []any {
		block := map[string]any{"bucket": "b", "region": "r"}
		for k, v := range creds {
			block[k] = v
		}

		return []any{map[string]any{"s3": []any{block}}}
	}

	// The API requires the credentials field, so an s3 block without one
	// would be rejected at apply time with a bare HTTP 400.
	if diags := validateParameters(t, s3(nil)); !diags.HasError() {
		t.Fatal("expected an error for an s3 block with no credentials")
	}

	if diags := validateParameters(t, s3(map[string]any{
		"auto_credentials": []any{map[string]any{}},
		"static_credentials": []any{
			map[string]any{"access_key_id": "AKIA"},
		},
	})); !diags.HasError() {
		t.Fatal("expected an error for both credential modes")
	}

	if diags := validateParameters(t, s3(map[string]any{
		"auto_credentials": []any{map[string]any{}},
	})); diags.HasError() {
		t.Fatalf("expected an auto_credentials config to validate, got %+v", diags)
	}
}

func TestValidateAbsentRecordingsStorage(t *testing.T) {
	if diags := validateParameters(t, nil); diags.HasError() {
		t.Fatalf("expected a config without recordings_storage to validate, got %+v", diags)
	}
}

func TestMfaEnforcementValidation(t *testing.T) {
	for _, mode := range []string{"Off", "Enroll", "Require"} {
		diags := resourceParameters().Validate(terraform.NewResourceConfigRaw(map[string]any{
			"allow_own_credential_management": true,
			"mfa_enforcement":                 mode,
		}))
		if diags.HasError() {
			t.Fatalf("expected %q to validate, got %+v", mode, diags)
		}
	}

	diags := resourceParameters().Validate(terraform.NewResourceConfigRaw(map[string]any{
		"allow_own_credential_management": true,
		"mfa_enforcement":                 "Always",
	}))
	if !diags.HasError() {
		t.Fatal("expected an error for an invalid mfa_enforcement value")
	}
}

func TestExpandDefaultCredentialPolicyUnsetIsNil(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{})

	// A partial update must not clear the policy Warpgate already has.
	if policy := expandDefaultCredentialPolicy(d); policy != nil {
		t.Fatalf("expected nil for an unconfigured policy, got %+v", policy)
	}
}

// One list per protocol Warpgate knows: a protocol missing from the mapping
// would silently vanish on the way to or from the API.
func TestCredentialPolicyRoundTripCoversEveryProtocol(t *testing.T) {
	raw := map[string]any{}
	for _, protocol := range credentialPolicyProtocols {
		raw[protocol] = []any{"Password", "Totp"}
	}

	policy := expandCredentialPolicy([]any{raw})
	for protocol, list := range credentialPolicyLists(policy) {
		if len(*list) != 2 {
			t.Fatalf("expected %s to carry two kinds, got %v", protocol, *list)
		}
	}

	back := flattenCredentialPolicy(policy)[0].(map[string]any)
	for _, protocol := range credentialPolicyProtocols {
		if got := back[protocol].([]any); len(got) != 2 || got[0] != "Password" {
			t.Fatalf("expected %s to round-trip, got %v", protocol, got)
		}
	}

	if err := validateCredentialPolicy("default_credential_policy", []any{map[string]any{"telnet": []any{"Password"}}}); err == nil {
		t.Fatal("expected an unknown protocol key to be rejected")
	}
}

// The block is Optional+Computed, so before Warpgate has reported a policy back
// the plan carries a one-element list with nothing in it. Warpgate versions that
// do not seed a new user's policy leave it that way, and rejecting it there made
// every plan fail with "credential_policy must be a map".
func TestValidateCredentialPolicyAcceptsAnUnreadComputedBlock(t *testing.T) {
	if err := validateCredentialPolicy("credential_policy", []any{nil}); err != nil {
		t.Fatalf("expected an unread computed block to pass validation, got %v", err)
	}
}

func TestExpandRecordingsStorageAzureCredentialModes(t *testing.T) {
	for _, tc := range []struct {
		block string
		value map[string]any
		want  client.AzureCredentials
	}{
		{"managed_identity", nil, client.AzureCredentials{Mode: "ManagedIdentity"}},
		{"managed_identity", map[string]any{"client_id": "id"}, client.AzureCredentials{Mode: "ManagedIdentity", ClientID: stringPtr("id")}},
		{"workload_identity", nil, client.AzureCredentials{Mode: "WorkloadIdentity"}},
		{
			"service_principal",
			map[string]any{"tenant_id": "t", "client_id": "c", "client_secret": "shhh"},
			client.AzureCredentials{Mode: "ServicePrincipal", TenantID: stringPtr("t"), ClientID: stringPtr("c"), ClientSecret: stringPtr("shhh")},
		},
		// client_secret is required by the API; "" tells Warpgate to keep the stored one.
		{
			"service_principal",
			map[string]any{"tenant_id": "t", "client_id": "c"},
			client.AzureCredentials{Mode: "ServicePrincipal", TenantID: stringPtr("t"), ClientID: stringPtr("c"), ClientSecret: stringPtr("")},
		},
		{"developer_tools", nil, client.AzureCredentials{Mode: "DeveloperTools"}},
	} {
		t.Run(tc.want.Mode, func(t *testing.T) {
			var element any
			if tc.value != nil {
				element = tc.value
			}

			cfg := buildRecordingsStorage(map[string]any{
				"azure": []any{
					map[string]any{
						"account":   "acct",
						"container": "recordings",
						"prefix":    "",
						tc.block:    []any{element},
					},
				},
			})
			if cfg == nil || cfg.Kind != "Azure" {
				t.Fatalf("expected an Azure config, got %+v", cfg)
			}

			if stringValue(cfg.Account) != "acct" || stringValue(cfg.Container) != "recordings" {
				t.Fatalf("expected account and container, got %+v", cfg)
			}

			// prefix is required by the API, so "" must be sent.
			if cfg.Prefix == nil || *cfg.Prefix != "" {
				t.Fatalf("expected an empty prefix to be sent, got %v", cfg.Prefix)
			}

			if cfg.Endpoint != nil || cfg.Bucket != nil || cfg.Credentials != nil || cfg.Path != nil {
				t.Fatalf("expected no other backend's fields, got %+v", cfg)
			}

			got := cfg.AzureCredentials
			if got == nil {
				t.Fatal("expected Azure credentials")
			}

			for name, pair := range map[string][2]*string{
				"tenant_id":     {got.TenantID, tc.want.TenantID},
				"client_id":     {got.ClientID, tc.want.ClientID},
				"client_secret": {got.ClientSecret, tc.want.ClientSecret},
			} {
				if (pair[0] == nil) != (pair[1] == nil) || stringValue(pair[0]) != stringValue(pair[1]) {
					t.Fatalf("%s: got %v, want %v", name, pair[0], pair[1])
				}
			}

			if got.Mode != tc.want.Mode {
				t.Fatalf("expected mode %q, got %q", tc.want.Mode, got.Mode)
			}
		})
	}
}

func TestFlattenRecordingsStorageAzureCredentialModes(t *testing.T) {
	for _, tc := range []struct {
		creds client.AzureCredentials
		block string
		want  map[string]any
	}{
		{client.AzureCredentials{Mode: "ManagedIdentity", ClientID: stringPtr("id")}, "managed_identity", map[string]any{"client_id": "id"}},
		{client.AzureCredentials{Mode: "WorkloadIdentity"}, "workload_identity", map[string]any{}},
		{client.AzureCredentials{Mode: "DeveloperTools"}, "developer_tools", map[string]any{}},
		{
			client.AzureCredentials{Mode: "ServicePrincipal", TenantID: stringPtr("t"), ClientID: stringPtr("c"), ClientSecret: stringPtr("")},
			"service_principal",
			map[string]any{"tenant_id": "t", "client_id": "c", "client_secret": ""},
		},
	} {
		t.Run(tc.creds.Mode, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{})

			got := flattenRecordingsStorage(d, client.RecordingsStorageConfig{
				Kind:                 "Azure",
				Account:              stringPtr("acct"),
				Container:            stringPtr("recordings"),
				Prefix:               stringPtr("sessions/"),
				ServeThroughWarpgate: boolPtr(true),
				AzureCredentials:     &tc.creds,
			})

			storage := got[0].(map[string]any)
			if _, ok := storage["s3"]; ok {
				t.Fatalf("expected no s3 block on an Azure config, got %+v", storage)
			}

			azure := storage["azure"].([]any)[0].(map[string]any)
			if azure["account"] != "acct" || azure["container"] != "recordings" || azure["prefix"] != "sessions/" || azure["serve_through_warpgate"] != true {
				t.Fatalf("unexpected azure block %+v", azure)
			}

			block, ok := azure[tc.block].([]any)
			if !ok || len(block) != 1 {
				t.Fatalf("expected one %s block, got %+v", tc.block, azure)
			}

			for k, v := range tc.want {
				if block[0].(map[string]any)[k] != v {
					t.Fatalf("%s.%s: got %v, want %v", tc.block, k, block[0].(map[string]any)[k], v)
				}
			}

			if diags := validateParameters(t, []any{map[string]any{"azure": []any{azure}}}); diags.HasError() {
				t.Fatalf("expected the flattened block to validate, got %+v", diags)
			}
		})
	}
}

func TestFlattenRecordingsStorageKeepsSecretsFromState(t *testing.T) {
	// The API redacts secrets to "" (or omits them, on older servers); either
	// way a read that echoed the response would wipe them from state.
	for _, redacted := range []*string{nil, stringPtr("")} {
		s3 := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{
			"recordings_storage": []any{map[string]any{"s3": []any{map[string]any{
				"bucket": "b",
				"region": "r",
				"static_credentials": []any{map[string]any{
					"access_key_id":     "AKIA",
					"secret_access_key": "shhh",
				}},
			}}}},
		})

		got := flattenRecordingsStorage(s3, client.RecordingsStorageConfig{
			Kind:        "S3",
			Credentials: &client.S3Credentials{Mode: "Static", AccessKeyID: stringPtr("AKIA"), SecretAccessKey: redacted},
		})
		static := got[0].(map[string]any)["s3"].([]any)[0].(map[string]any)["static_credentials"].([]any)[0].(map[string]any)
		if static["secret_access_key"] != "shhh" {
			t.Fatalf("S3 (redacted=%v): expected the stored secret, got %v", redacted, static["secret_access_key"])
		}

		azure := schema.TestResourceDataRaw(t, resourceParameters().Schema, map[string]any{
			"recordings_storage": []any{map[string]any{"azure": []any{map[string]any{
				"account":   "acct",
				"container": "c",
				"service_principal": []any{map[string]any{
					"tenant_id":     "t",
					"client_id":     "c",
					"client_secret": "shhh",
				}},
			}}}},
		})

		got = flattenRecordingsStorage(azure, client.RecordingsStorageConfig{
			Kind: "Azure",
			AzureCredentials: &client.AzureCredentials{
				Mode:         "ServicePrincipal",
				TenantID:     stringPtr("t"),
				ClientID:     stringPtr("c"),
				ClientSecret: redacted,
			},
		})
		principal := got[0].(map[string]any)["azure"].([]any)[0].(map[string]any)["service_principal"].([]any)[0].(map[string]any)
		if principal["client_secret"] != "shhh" {
			t.Fatalf("Azure (redacted=%v): expected the stored secret, got %v", redacted, principal["client_secret"])
		}
	}
}

// storageResourceData builds ResourceData whose raw config carries the
// backend block, which TestResourceDataRaw leaves null.
func storageResourceData(t *testing.T, backend string, serve cty.Value) *schema.ResourceData {
	t.Helper()

	attrs := map[string]string{
		"recordings_storage.#":                                      "1",
		"recordings_storage.0." + backend + ".#":                    "1",
		"recordings_storage.0." + backend + ".0.prefix":             "",
		"recordings_storage.0." + backend + ".0.bucket":             "b",
		"recordings_storage.0." + backend + ".0.region":             "r",
		"recordings_storage.0." + backend + ".0.account":            "acct",
		"recordings_storage.0." + backend + ".0.container":          "c",
		"recordings_storage.0." + backend + ".0.auto_credentials.#": "1",
		"recordings_storage.0." + backend + ".0.developer_tools.#":  "1",
	}

	block := map[string]cty.Value{}
	if !serve.IsNull() {
		block["serve_through_warpgate"] = serve
	}

	return resourceParameters().Data(&terraform.InstanceState{
		ID:         "parameters",
		Attributes: attrs,
		RawConfig: cty.ObjectVal(map[string]cty.Value{
			"recordings_storage": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
				backend: cty.ListVal([]cty.Value{cty.ObjectVal(block)}),
			})}),
		}),
	})
}

func TestExpandRecordingsStorageServeThroughWarpgate(t *testing.T) {
	for _, backend := range []string{"s3", "azure"} {
		// Unset must not be sent: the server default differs per backend.
		if cfg := expandRecordingsStorage(storageResourceData(t, backend, cty.NullVal(cty.Bool))); cfg == nil || cfg.ServeThroughWarpgate != nil {
			t.Fatalf("%s: expected serve_through_warpgate to be omitted, got %+v", backend, cfg)
		}

		for _, want := range []bool{true, false} {
			cfg := expandRecordingsStorage(storageResourceData(t, backend, cty.BoolVal(want)))
			if cfg == nil || cfg.ServeThroughWarpgate == nil || *cfg.ServeThroughWarpgate != want {
				t.Fatalf("%s: expected serve_through_warpgate %v, got %+v", backend, want, cfg)
			}
		}
	}
}

func TestValidateRecordingsStorageRejectsAzureWithAnotherBackend(t *testing.T) {
	azure := []any{map[string]any{
		"account":         "acct",
		"container":       "c",
		"developer_tools": []any{map[string]any{}},
	}}

	for name, other := range map[string]any{
		"disk": []any{map[string]any{"path": "/rec"}},
		"s3": []any{map[string]any{
			"bucket":           "b",
			"region":           "r",
			"auto_credentials": []any{map[string]any{}},
		}},
	} {
		if diags := validateParameters(t, []any{map[string]any{"azure": azure, name: other}}); !diags.HasError() {
			t.Fatalf("expected an error for both azure and %s", name)
		}
	}

	if diags := validateParameters(t, []any{map[string]any{"azure": azure}}); diags.HasError() {
		t.Fatalf("expected an azure-only config to validate, got %+v", diags)
	}
}

func TestValidateAzureRequiresExactlyOneCredentialsMode(t *testing.T) {
	azure := func(creds map[string]any) []any {
		block := map[string]any{"account": "acct", "container": "c"}
		for k, v := range creds {
			block[k] = v
		}

		return []any{map[string]any{"azure": []any{block}}}
	}

	if diags := validateParameters(t, azure(nil)); !diags.HasError() {
		t.Fatal("expected an error for an azure block with no credentials")
	}

	if diags := validateParameters(t, azure(map[string]any{
		"workload_identity": []any{map[string]any{}},
		"developer_tools":   []any{map[string]any{}},
	})); !diags.HasError() {
		t.Fatal("expected an error for two credential modes")
	}
}
