package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func TestAccSecretBackend(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "warpgate_secret_backend" "b" {
  name          = "acctest-bao"
  backend_type  = "openbao"
  address       = "https://bao.acctest.invalid:8200"
  allowed_paths = ["ssh-warpgate/sign/deploy", "secret/warpgate"]

  token_auth {
    token = "s.acctest-token"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "name", "acctest-bao"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "backend_type", "openbao"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "allowed_paths.#", "2"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "allowed_paths.0", "ssh-warpgate/sign/deploy"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "token_auth.0.token", "s.acctest-token"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "tls_skip_verify", "false"),
					testAccCheckSecretBackendAuth(t, "warpgate_secret_backend.b", client.SecretBackendAuth{Method: client.SecretBackendAuthToken}),
				),
			},
			{
				Config: `resource "warpgate_secret_backend" "b" {
  name            = "acctest-bao"
  backend_type    = "vault"
  address         = "https://vault.acctest.invalid:8200"
  namespace       = "infra"
  tls_skip_verify = true

  approle_auth {
    role_id   = "acctest-role-id"
    secret_id = "acctest-secret-id"
    mount     = "approle-warpgate"
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "backend_type", "vault"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "namespace", "infra"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "tls_skip_verify", "true"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "allowed_paths.#", "0"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "token_auth.#", "0"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "approle_auth.0.role_id", "acctest-role-id"),
					resource.TestCheckResourceAttr("warpgate_secret_backend.b", "approle_auth.0.secret_id", "acctest-secret-id"),
					testAccCheckSecretBackendAuth(t, "warpgate_secret_backend.b", client.SecretBackendAuth{
						Method: client.SecretBackendAuthAppRole,
						RoleID: "acctest-role-id",
						Mount:  "approle-warpgate",
					}),
				),
			},
			{
				ResourceName:            "warpgate_secret_backend.b",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"approle_auth.0.secret_id"},
			},
		},
	})
}

func TestAccTargetSSHCertificateAuth(t *testing.T) {
	config := func(mount, role string) string {
		return fmt.Sprintf(`resource "warpgate_secret_backend" "b" {
  name          = "acctest-ssh-ca"
  backend_type  = "openbao"
  address       = "https://bao.acctest.invalid:8200"
  allowed_paths = ["%[1]s/sign/%[2]s"]

  token_auth {
    token = "s.acctest-token"
  }
}

resource "warpgate_target" "t" {
  name = "acctest-ssh-certificate"

  ssh_options {
    host     = "10.0.0.10"
    port     = 22
    username = "deploy"

    certificate_auth {
      backend = warpgate_secret_backend.b.name
      mount   = "%[1]s"
      role    = "%[2]s"
    }
  }
}`, mount, role)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("ssh-warpgate", "deploy"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.t", "ssh_options.0.certificate_auth.0.backend", "acctest-ssh-ca"),
					resource.TestCheckResourceAttr("warpgate_target.t", "ssh_options.0.certificate_auth.0.mount", "ssh-warpgate"),
					resource.TestCheckResourceAttr("warpgate_target.t", "ssh_options.0.certificate_auth.0.role", "deploy"),
					resource.TestCheckResourceAttr("warpgate_target.t", "ssh_options.0.public_key_auth.#", "0"),
				),
			},
			{
				Config: config("ssh", "warpgate"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.t", "ssh_options.0.certificate_auth.0.mount", "ssh"),
					resource.TestCheckResourceAttr("warpgate_target.t", "ssh_options.0.certificate_auth.0.role", "warpgate"),
				),
			},
			{
				ResourceName:      "warpgate_target.t",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSecretBackendAuth(t *testing.T, name string, want client.SecretBackendAuth) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource %s not found in state", name)
		}

		backend, err := testAccClient(t).GetSecretBackend(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}
		if backend == nil {
			return fmt.Errorf("secret backend %s not found in Warpgate", rs.Primary.ID)
		}
		if backend.Auth != want {
			return fmt.Errorf("expected auth %+v in Warpgate, got %+v", want, backend.Auth)
		}

		return nil
	}
}
