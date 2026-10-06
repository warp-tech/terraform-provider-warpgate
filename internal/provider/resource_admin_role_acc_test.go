package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAdminRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "warpgate_admin_role" "r" {
  name               = "acctest-security"
  description        = "Security team"
  sessions_view      = true
  sessions_terminate = true
  approve_sessions   = true
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "name", "acctest-security"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "description", "Security team"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "sessions_view", "true"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "sessions_terminate", "true"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "approve_sessions", "true"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "config_edit", "false"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "admin_roles_manage", "false"),
					testAccCheckAdminRolePermissions(t, "warpgate_admin_role.r", map[string]bool{
						"sessions_view":      true,
						"sessions_terminate": true,
						"approve_sessions":   true,
					}),
				),
			},
			{
				// Dropping an attribute from the config revokes the permission
				Config: `resource "warpgate_admin_role" "r" {
  name            = "acctest-security"
  sessions_view   = true
  recordings_view = true
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "description", ""),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "sessions_terminate", "false"),
					resource.TestCheckResourceAttr("warpgate_admin_role.r", "recordings_view", "true"),
					testAccCheckAdminRolePermissions(t, "warpgate_admin_role.r", map[string]bool{
						"sessions_view":   true,
						"recordings_view": true,
					}),
				),
			},
			{
				ResourceName:      "warpgate_admin_role.r",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserAdminRole(t *testing.T) {
	config := `resource "warpgate_user" "u" {
  username = "acctest-admin-role-user"
}

resource "warpgate_admin_role" "r" {
  name          = "acctest-assigned"
  sessions_view = true
}

resource "warpgate_user_admin_role" "a" {
  user_id       = warpgate_user.u.id
  admin_role_id = warpgate_admin_role.r.id
}`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("warpgate_user_admin_role.a", "user_id", "warpgate_user.u", "id"),
					resource.TestCheckResourceAttrPair("warpgate_user_admin_role.a", "admin_role_id", "warpgate_admin_role.r", "id"),
					testAccCheckUserHasAdminRole(t, "warpgate_user.u", "warpgate_admin_role.r", true),
				),
			},
			{
				ResourceName:      "warpgate_user_admin_role.a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Removing the assignment leaves the user and the role in place
				Config: `resource "warpgate_user" "u" {
  username = "acctest-admin-role-user"
}

resource "warpgate_admin_role" "r" {
  name          = "acctest-assigned"
  sessions_view = true
}`,
				Check: testAccCheckUserHasAdminRole(t, "warpgate_user.u", "warpgate_admin_role.r", false),
			},
		},
	})
}

func TestAccAdminRoleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "warpgate_admin_role" "r" {
  name          = "acctest-lookup"
  tickets_delete = true
}

data "warpgate_admin_role" "by_name" {
  name = warpgate_admin_role.r.name
}

data "warpgate_admin_role" "by_id" {
  id = warpgate_admin_role.r.id
}

# The built-in role ships with every instance
data "warpgate_admin_role" "builtin" {
  name = "warpgate:admin"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.warpgate_admin_role.by_name", "id", "warpgate_admin_role.r", "id"),
					resource.TestCheckResourceAttr("data.warpgate_admin_role.by_name", "tickets_delete", "true"),
					resource.TestCheckResourceAttr("data.warpgate_admin_role.by_name", "config_edit", "false"),
					resource.TestCheckResourceAttr("data.warpgate_admin_role.by_id", "name", "acctest-lookup"),
					resource.TestCheckResourceAttr("data.warpgate_admin_role.builtin", "admin_roles_manage", "true"),
					resource.TestCheckResourceAttr("data.warpgate_admin_role.builtin", "config_edit", "true"),
				),
			},
		},
	})
}

// testAccCheckAdminRolePermissions asserts that Warpgate holds exactly the
// given granted permissions for the role, with every other permission off.
func testAccCheckAdminRolePermissions(t *testing.T, resourceName string, granted map[string]bool) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", resourceName)
		}

		role, err := testAccClient(t).GetAdminRole(context.Background(), rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read admin role %s: %w", rs.Primary.ID, err)
		}
		if role == nil {
			return fmt.Errorf("admin role %s does not exist in Warpgate", rs.Primary.ID)
		}

		for _, attribute := range adminRolePermissionAttributes {
			want := granted[attribute.key]
			got := *attribute.field(&role.AdminRolePermissions)
			if got != want {
				return fmt.Errorf("permission %s: Warpgate has %t, want %t", attribute.key, got, want)
			}
		}

		return nil
	}
}

// testAccCheckUserHasAdminRole asserts whether Warpgate lists the admin role
// among the user's assignments.
func testAccCheckUserHasAdminRole(t *testing.T, userResource, roleResource string, want bool) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		user, ok := s.RootModule().Resources[userResource]
		if !ok {
			return fmt.Errorf("%s not found in state", userResource)
		}
		role, ok := s.RootModule().Resources[roleResource]
		if !ok {
			return fmt.Errorf("%s not found in state", roleResource)
		}

		roles, err := testAccClient(t).GetUserAdminRoles(context.Background(), user.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to read admin roles of user %s: %w", user.Primary.ID, err)
		}

		got := false
		for _, r := range roles {
			if r.ID == role.Primary.ID {
				got = true
			}
		}

		if got != want {
			return fmt.Errorf("user %s has admin role %s: %t, want %t", user.Primary.ID, role.Primary.ID, got, want)
		}

		return nil
	}
}
