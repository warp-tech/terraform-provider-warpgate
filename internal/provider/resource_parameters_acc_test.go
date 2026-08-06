package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// warpgate_parameters is a singleton, so these tests mutate instance-wide state
// and must not run in parallel with each other.

const testAccParametersRecordingsDisk = `
resource "warpgate_parameters" "test" {
  allow_own_credential_management = true

  recordings_enable = true
  recordings_storage {
    disk {
      path = "/tmp/warpgate-acctest-recordings"
    }
  }
}
`

const testAccParametersRecordingsS3Static = `
resource "warpgate_parameters" "test" {
  allow_own_credential_management = true

  recordings_enable = true
  recordings_storage {
    s3 {
      bucket     = "warpgate-acctest"
      region     = "us-east-1"
      endpoint   = "http://localhost:19000"
      path_style = true
      prefix     = "sessions/"

      static_credentials {
        access_key_id     = "acctest-key"
        secret_access_key = "acctest-secret"
      }
    }
  }
}
`

const testAccParametersRecordingsS3Auto = `
resource "warpgate_parameters" "test" {
  allow_own_credential_management = true

  recordings_enable = true
  recordings_storage {
    s3 {
      bucket     = "warpgate-acctest"
      region     = "us-east-1"
      path_style = true

      auto_credentials {}
    }
  }
}
`

const testAccParameters027Fields = `
resource "warpgate_parameters" "test" {
  allow_own_credential_management = true

  banner                    = "Authorized access only"
  web_clients_enabled       = true
  ssh_host_key_verification = "AutoAccept"
  web_auth_max_age_seconds  = 28800
}
`

// A fresh 0.27 database seeds recordings_enable = false with the upstream
// default path, ignoring warpgate.yaml, so turning recording on is only
// possible through these parameters.
func TestAccParametersRecordingsDisk(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParametersRecordingsDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_enable", "true"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.disk.0.path", "/tmp/warpgate-acctest-recordings"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.#", "0"),
					testAccCheckRecordingsStorage(t, func(t *testing.T, params recordingsStorage) error {
						if !params.enabled {
							return fmt.Errorf("expected recording to be enabled in Warpgate")
						}

						if params.kind != "Disk" {
							return fmt.Errorf("expected kind Disk in Warpgate, got %q", params.kind)
						}

						if params.path != "/tmp/warpgate-acctest-recordings" {
							return fmt.Errorf("expected the configured path in Warpgate, got %q", params.path)
						}

						return nil
					}),
				),
			},
		},
	})
}

// The API redacts secret_access_key on read, so a read path that echoed the
// response would wipe it from state. The framework plans again after each
// step and fails on a non-empty plan, which is what catches that here.
func TestAccParametersRecordingsS3(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Start on Disk so the next step exercises the union swap.
				Config: testAccParametersRecordingsDisk,
			},
			{
				Config: testAccParametersRecordingsS3Static,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.disk.#", "0"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.0.bucket", "warpgate-acctest"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.0.path_style", "true"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.0.prefix", "sessions/"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.0.static_credentials.0.secret_access_key", "acctest-secret"),
					testAccCheckRecordingsStorage(t, func(t *testing.T, params recordingsStorage) error {
						if params.kind != "S3" {
							return fmt.Errorf("expected kind S3 in Warpgate, got %q", params.kind)
						}

						if params.path != "" {
							return fmt.Errorf("expected no Disk path on an S3 config, got %q", params.path)
						}

						// Required by the API and legitimately false and "", so
						// they prove the fields were sent rather than dropped.
						if !params.pathStyle {
							return fmt.Errorf("expected path_style to have been sent")
						}

						if params.credentialsMode != "Static" {
							return fmt.Errorf("expected Static credentials, got %q", params.credentialsMode)
						}

						return nil
					}),
				),
			},
			{
				// Re-planning the same config must be a no-op even though
				// Warpgate never returns the secret.
				Config:   testAccParametersRecordingsS3Static,
				PlanOnly: true,
			},
			{
				Config: testAccParametersRecordingsS3Auto,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.0.auto_credentials.#", "1"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "recordings_storage.0.s3.0.static_credentials.#", "0"),
					testAccCheckRecordingsStorage(t, func(t *testing.T, params recordingsStorage) error {
						if params.credentialsMode != "Auto" {
							return fmt.Errorf("expected Auto credentials, got %q", params.credentialsMode)
						}

						return nil
					}),
				),
			},
		},
	})
}

func TestAccParameters027Fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParameters027Fields,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_parameters.test", "banner", "Authorized access only"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "ssh_host_key_verification", "AutoAccept"),
					resource.TestCheckResourceAttr("warpgate_parameters.test", "web_auth_max_age_seconds", "28800"),
					func(*terraform.State) error {
						params := testAccParameters(t)

						if params.Banner != "Authorized access only" {
							return fmt.Errorf("expected the banner in Warpgate, got %q", params.Banner)
						}

						if params.SSHHostKeyVerification != "AutoAccept" {
							return fmt.Errorf("expected AutoAccept in Warpgate, got %q", params.SSHHostKeyVerification)
						}

						if params.WebAuthMaxAgeSeconds != 28800 {
							return fmt.Errorf("expected 28800 in Warpgate, got %d", params.WebAuthMaxAgeSeconds)
						}

						return nil
					},
				),
			},
		},
	})
}

type recordingsStorage struct {
	enabled         bool
	kind            string
	path            string
	pathStyle       bool
	credentialsMode string
}

// testAccCheckRecordingsStorage asserts against what Warpgate actually stored,
// not against Terraform state.
func testAccCheckRecordingsStorage(t *testing.T, check func(*testing.T, recordingsStorage) error) resource.TestCheckFunc {
	t.Helper()

	return func(*terraform.State) error {
		params := testAccParameters(t)
		storage := params.RecordingsStorage

		got := recordingsStorage{
			enabled:   params.RecordingsEnable,
			kind:      storage.Kind,
			path:      stringValue(storage.Path),
			pathStyle: boolValue(storage.PathStyle),
		}
		if storage.Credentials != nil {
			got.credentialsMode = storage.Credentials.Mode
		}

		return check(t, got)
	}
}
