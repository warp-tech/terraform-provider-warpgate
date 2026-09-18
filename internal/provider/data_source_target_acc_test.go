package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Looking a target up by id and by name are separate code paths in the data
// source, and only the id one carries the fetched target into state itself.
const testAccTargetDataSource = `
resource "warpgate_target" "vnc" {
  name = "acctest-ds-vnc"

  vnc_options {
    host     = "desktop.example.com"
    port     = 5901
    password = "AccTestPassword123!"
  }
}

data "warpgate_target" "by_id" {
  id = warpgate_target.vnc.id
}

data "warpgate_target" "by_name" {
  name = warpgate_target.vnc.name
}
`

func TestAccTargetDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckTargetDestroyed(t, "warpgate_target.vnc"),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.warpgate_target.by_id", "name", "acctest-ds-vnc"),
					resource.TestCheckResourceAttr("data.warpgate_target.by_id", "vnc_options.0.host", "desktop.example.com"),
					resource.TestCheckResourceAttr("data.warpgate_target.by_id", "vnc_options.0.port", "5901"),
					resource.TestCheckResourceAttr("data.warpgate_target.by_name", "vnc_options.0.host", "desktop.example.com"),
				),
			},
		},
	})
}
