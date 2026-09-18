package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testAccTargetRDP = `
resource "warpgate_target" "rdp" {
  name = "acctest-rdp"

  rdp_options {
    host         = "windows.example.com"
    port         = 3389
    username     = "Administrator"
    password     = "AccTestPassword123!"
    domain       = "EXAMPLE"
    verify_tls   = false
    tls_security = "Tls12"
  }
}
`

const testAccTargetRDPUpdated = `
resource "warpgate_target" "rdp" {
  name = "acctest-rdp"

  rdp_options {
    host         = "windows2.example.com"
    port         = 13389
    username     = "Administrator"
    password     = "AccTestPassword456!"
    verify_tls   = true
    tls_security = "Tls12WithLegacyCiphers"
    compression       = "lossless"
    interactive_logon = true
  }
}
`

const testAccTargetMySQLPassword = `
resource "warpgate_target" "db" {
  name             = "acctest-mysql"
  require_approval = true

  mysql_options {
    host     = "db.example.com"
    port     = 3306
    username = "app"
    password = "AccTestPassword123!"
    tls {
      mode   = "Preferred"
      verify = false
    }
  }
}
`

// require_approval is computed as well as optional, like its ticket_* siblings:
// dropping it from the config keeps whatever the gate is set to, so switching
// it off has to be said explicitly.
const testAccTargetMySQLIamRole = `
resource "warpgate_target" "db" {
  name             = "acctest-mysql"
  require_approval = false

  mysql_options {
    host     = "db.example.com"
    port     = 3306
    username = "app"
    iam_role_auth {}
    tls {
      mode   = "Preferred"
      verify = false
    }
  }
}
`

const testAccTargetVNC = `
resource "warpgate_target" "vnc" {
  name = "acctest-vnc"

  vnc_options {
    host     = "desktop.example.com"
    password = "AccTestPassword123!"
  }
}
`

const testAccTargetVNCNoAuth = `
resource "warpgate_target" "vnc" {
  name = "acctest-vnc"

  vnc_options {
    host = "desktop2.example.com"
    port = 5901
  }
}
`

func TestAccTargetVNC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroyed(t, "warpgate_target.vnc"),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetVNC,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.vnc", "vnc_options.0.host", "desktop.example.com"),
					resource.TestCheckResourceAttr("warpgate_target.vnc", "vnc_options.0.port", "5900"),
					resource.TestCheckResourceAttr("warpgate_target.vnc", "vnc_options.0.password", "AccTestPassword123!"),
					testAccCheckTargetOptions(t, "warpgate_target.vnc", func(t *testing.T, options map[string]any) error {
						if kind := options["kind"]; kind != "Vnc" {
							return fmt.Errorf("expected kind Vnc in Warpgate, got %v", kind)
						}

						if kind := options["auth"].(map[string]any)["kind"]; kind != "Password" {
							return fmt.Errorf("expected auth kind Password in Warpgate, got %v", kind)
						}

						return nil
					}),
				),
			},
			{
				Config: testAccTargetVNCNoAuth,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.vnc", "vnc_options.0.host", "desktop2.example.com"),
					resource.TestCheckResourceAttr("warpgate_target.vnc", "vnc_options.0.port", "5901"),
					resource.TestCheckResourceAttr("warpgate_target.vnc", "vnc_options.0.password", ""),
					testAccCheckTargetOptions(t, "warpgate_target.vnc", func(t *testing.T, options map[string]any) error {
						if kind := options["auth"].(map[string]any)["kind"]; kind != "None" {
							return fmt.Errorf("expected auth kind None in Warpgate, got %v", kind)
						}

						return nil
					}),
				),
			},
		},
	})
}

// The RDP option names and the Rdp / Password / Tls12 enum spellings are only
// checked by Warpgate itself: a create that reaches the API with the wrong
// casing is rejected, and a read that mishandles the response shows up as a
// non-empty plan after the step.
func TestAccTargetRDP(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroyed(t, "warpgate_target.rdp"),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetRDP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.host", "windows.example.com"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.port", "3389"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.username", "Administrator"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.password", "AccTestPassword123!"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.domain", "EXAMPLE"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.verify_tls", "false"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.tls_security", "Tls12"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "ssh_options.#", "0"),
					testAccCheckTargetOptions(t, "warpgate_target.rdp", func(t *testing.T, options map[string]any) error {
						if kind := options["kind"]; kind != "Rdp" {
							return fmt.Errorf("expected kind Rdp in Warpgate, got %v", kind)
						}

						if host := options["host"]; host != "windows.example.com" {
							return fmt.Errorf("expected the configured host in Warpgate, got %v", host)
						}

						if tlsSecurity := options["tls_security"]; tlsSecurity != "Tls12" {
							return fmt.Errorf("expected tls_security Tls12 in Warpgate, got %v", tlsSecurity)
						}

						auth, ok := options["auth"].(map[string]any)
						if !ok {
							return fmt.Errorf("expected an auth object in Warpgate, got %v", options["auth"])
						}

						if kind := auth["kind"]; kind != "Password" {
							return fmt.Errorf("expected auth kind Password in Warpgate, got %v", kind)
						}

						return nil
					}),
				),
			},
			{
				Config: testAccTargetRDPUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.host", "windows2.example.com"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.port", "13389"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.password", "AccTestPassword456!"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.domain", ""),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.verify_tls", "true"),
					resource.TestCheckResourceAttr("warpgate_target.rdp", "rdp_options.0.tls_security", "Tls12WithLegacyCiphers"),
					testAccCheckTargetOptions(t, "warpgate_target.rdp", func(t *testing.T, options map[string]any) error {
						if tlsSecurity := options["tls_security"]; tlsSecurity != "Tls12WithLegacyCiphers" {
							return fmt.Errorf("expected tls_security Tls12WithLegacyCiphers in Warpgate, got %v", tlsSecurity)
						}

						if compression := options["compression"]; compression != "lossless" {
							return fmt.Errorf("expected compression lossless in Warpgate, got %v", compression)
						}

						if logon := options["interactive_logon"]; logon != true {
							return fmt.Errorf("expected interactive_logon true in Warpgate, got %v", logon)
						}

						if verify := options["verify_tls"]; verify != true {
							return fmt.Errorf("expected verify_tls true in Warpgate, got %v", verify)
						}

						return nil
					}),
				),
			},
			{
				ResourceName:      "warpgate_target.rdp",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// The database password travels inside the `auth` union; a provider that sent
// the retired flat field would be accepted and stored with an empty password.
func TestAccTargetMySQLAuth(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroyed(t, "warpgate_target.db"),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetMySQLPassword,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.db", "require_approval", "true"),
					resource.TestCheckResourceAttr("warpgate_target.db", "mysql_options.0.password", "AccTestPassword123!"),
					resource.TestCheckResourceAttr("warpgate_target.db", "mysql_options.0.iam_role_auth.#", "0"),
					testAccCheckTargetOptions(t, "warpgate_target.db", func(t *testing.T, options map[string]any) error {
						auth, _ := options["auth"].(map[string]any)
						if auth["kind"] != "Password" || auth["password"] != "AccTestPassword123!" {
							return fmt.Errorf("expected a Password auth carrying the password in Warpgate, got %v", options["auth"])
						}

						return nil
					}),
					testAccCheckTargetRequireApproval(t, "warpgate_target.db", true),
				),
			},
			{
				Config: testAccTargetMySQLIamRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_target.db", "require_approval", "false"),
					resource.TestCheckResourceAttr("warpgate_target.db", "mysql_options.0.iam_role_auth.#", "1"),
					resource.TestCheckResourceAttr("warpgate_target.db", "mysql_options.0.password", ""),
					testAccCheckTargetOptions(t, "warpgate_target.db", func(t *testing.T, options map[string]any) error {
						auth, _ := options["auth"].(map[string]any)
						if auth["kind"] != "IamRole" {
							return fmt.Errorf("expected IamRole auth in Warpgate, got %v", options["auth"])
						}

						return nil
					}),
					testAccCheckTargetRequireApproval(t, "warpgate_target.db", false),
				),
			},
			{
				ResourceName:      "warpgate_target.db",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTargetRequireApproval(t *testing.T, resourceName string, want bool) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", resourceName)
		}

		target, err := testAccClient(t).GetTarget(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read target %s: %w", rs.Primary.ID, err)
		}

		if target.RequireApproval != want {
			return fmt.Errorf("expected require_approval %v in Warpgate, got %v", want, target.RequireApproval)
		}

		return nil
	}
}

// testAccCheckTargetOptions asserts against the options Warpgate actually
// stored, not against Terraform state.
func testAccCheckTargetOptions(t *testing.T, resourceName string, check func(*testing.T, map[string]any) error) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", resourceName)
		}

		target, err := testAccClient(t).GetTarget(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read target %s: %w", rs.Primary.ID, err)
		}

		options, ok := target.Options.(map[string]any)
		if !ok {
			return fmt.Errorf("expected the target options to decode to an object, got %T", target.Options)
		}

		return check(t, options)
	}
}

func testAccCheckTargetDestroyed(t *testing.T, resourceName string) resource.TestCheckFunc {
	t.Helper()

	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return nil
		}

		// GetTarget reports a missing target as (nil, nil), not as an error.
		target, err := testAccClient(t).GetTarget(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to check target %s: %w", rs.Primary.ID, err)
		}

		if target != nil {
			return fmt.Errorf("target %s still exists in Warpgate", rs.Primary.ID)
		}

		return nil
	}
}
