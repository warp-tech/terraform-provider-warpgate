package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRoleIsDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "warpgate_role" "r" {
  name       = "acctest-default-role"
  is_default = true
}`,
				Check: resource.TestCheckResourceAttr("warpgate_role.r", "is_default", "true"),
			},
			{
				Config: `resource "warpgate_role" "r" {
  name = "acctest-default-role"
}`,
				Check: resource.TestCheckResourceAttr("warpgate_role.r", "is_default", "false"),
			},
		},
	})
}
