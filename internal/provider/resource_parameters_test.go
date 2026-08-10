package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func stringPtr(v string) *string { return &v }

func boolPtr(v bool) *bool { return &v }

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
