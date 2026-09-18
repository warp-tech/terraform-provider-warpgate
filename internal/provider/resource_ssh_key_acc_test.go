package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Warpgate echoes a generated key's type back as the OpenSSH algorithm name and
// stores a public key without its comment, so both resources only settle if the
// provider compares those on their normalized form. A leftover diff here means
// every plan would replace the key or rewrite the credential.
const testAccSSHKeyAndCredential = `
resource "warpgate_ssh_key" "generated" {
  label = "acctest-generated"
  kind  = "Ed25519"
}

resource "warpgate_user" "u" {
  username = "acctest-key-user"
}

resource "warpgate_public_key_credential" "u" {
  user_id    = warpgate_user.u.id
  label      = "laptop"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ3bKQmMpPLVrTEQJ1bGnqLXQ1NfFHSy7lI6BoY7DDBM acctest@example.com"
}
`

func TestAccSSHKeyAndPublicKeyCredentialAreStable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyAndCredential,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("warpgate_ssh_key.generated", "public_key"),
					resource.TestCheckResourceAttr("warpgate_public_key_credential.u", "label", "laptop"),
				),
			},
			{
				// A second apply of the same config must be a no-op.
				Config:   testAccSSHKeyAndCredential,
				PlanOnly: true,
			},
		},
	})
}
