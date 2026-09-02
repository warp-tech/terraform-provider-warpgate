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
  }
}
`

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
