package client

import (
	"encoding/json"
	"testing"
)

func strPtr(v string) *string { return &v }

// S3 and Azure share the "credentials" key, so each must land in its own field.
func TestRecordingsStorageCredentialsRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  RecordingsStorageConfig
		want string
	}{
		{
			name: "S3",
			cfg: RecordingsStorageConfig{
				Kind:        "S3",
				Bucket:      strPtr("b"),
				Credentials: &S3Credentials{Mode: "Static", AccessKeyID: strPtr("AKIA")},
			},
			want: `{"kind":"S3","bucket":"b","credentials":{"mode":"Static","access_key_id":"AKIA"}}`,
		},
		{
			name: "Azure",
			cfg: RecordingsStorageConfig{
				Kind:    "Azure",
				Account: strPtr("acct"),
				AzureCredentials: &AzureCredentials{
					Mode:         "ServicePrincipal",
					TenantID:     strPtr("t"),
					ClientID:     strPtr("c"),
					ClientSecret: strPtr(""),
				},
			},
			want: `{"kind":"Azure","account":"acct","credentials":{"mode":"ServicePrincipal","tenant_id":"t","client_id":"c","client_secret":""}}`,
		},
		{
			name: "Disk",
			cfg:  RecordingsStorageConfig{Kind: "Disk", Path: strPtr("/rec")},
			want: `{"kind":"Disk","path":"/rec"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(&tc.cfg)
			if err != nil {
				t.Fatal(err)
			}

			if string(got) != tc.want {
				t.Fatalf("marshal:\n got %s\nwant %s", got, tc.want)
			}

			var back RecordingsStorageConfig
			if err := json.Unmarshal(got, &back); err != nil {
				t.Fatal(err)
			}

			again, _ := json.Marshal(back)
			if string(again) != tc.want {
				t.Fatalf("round trip:\n got %s\nwant %s", again, tc.want)
			}
		})
	}
}

func TestRecordingsStorageUnmarshalRoutesCredentialsByKind(t *testing.T) {
	var s3 RecordingsStorageConfig
	if err := json.Unmarshal([]byte(`{"kind":"S3","credentials":{"mode":"Static","access_key_id":"AKIA","secret_access_key":""},"serve_through_warpgate":false}`), &s3); err != nil {
		t.Fatal(err)
	}

	if s3.Credentials == nil || s3.Credentials.Mode != "Static" || s3.AzureCredentials != nil {
		t.Fatalf("expected S3 credentials only, got %+v / %+v", s3.Credentials, s3.AzureCredentials)
	}

	if s3.ServeThroughWarpgate == nil || *s3.ServeThroughWarpgate {
		t.Fatalf("expected serve_through_warpgate false, got %v", s3.ServeThroughWarpgate)
	}

	var azure RecordingsStorageConfig
	if err := json.Unmarshal([]byte(`{"kind":"Azure","account":"acct","container":"c","prefix":"","credentials":{"mode":"ManagedIdentity","client_id":"id"}}`), &azure); err != nil {
		t.Fatal(err)
	}

	if azure.AzureCredentials == nil || azure.AzureCredentials.Mode != "ManagedIdentity" || azure.Credentials != nil {
		t.Fatalf("expected Azure credentials only, got %+v / %+v", azure.Credentials, azure.AzureCredentials)
	}

	if azure.AzureCredentials.ClientID == nil || *azure.AzureCredentials.ClientID != "id" {
		t.Fatalf("expected client_id, got %v", azure.AzureCredentials.ClientID)
	}
}
