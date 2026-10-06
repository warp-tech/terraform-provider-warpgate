package provider

import (
	"reflect"
	"testing"

	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// The attribute table must name every permission the API has, under the API's
// own spelling, or a permission would silently never reach Warpgate.
func TestAdminRolePermissionAttributesCoverAPI(t *testing.T) {
	apiFields := map[string]bool{}
	permissionsType := reflect.TypeOf(client.AdminRolePermissions{})
	for i := 0; i < permissionsType.NumField(); i++ {
		apiFields[permissionsType.Field(i).Tag.Get("json")] = true
	}

	seen := map[string]bool{}
	var probe client.AdminRolePermissions
	for _, attribute := range adminRolePermissionAttributes {
		if seen[attribute.key] {
			t.Errorf("attribute %q is listed twice", attribute.key)
		}
		seen[attribute.key] = true

		if !apiFields[attribute.key] {
			t.Errorf("attribute %q has no field in client.AdminRolePermissions", attribute.key)
		}

		// Each accessor must point at a distinct field
		*attribute.field(&probe) = true
	}

	for field := range apiFields {
		if !seen[field] {
			t.Errorf("API permission %q is missing from adminRolePermissionAttributes", field)
		}
	}

	all := reflect.ValueOf(probe)
	for i := 0; i < all.NumField(); i++ {
		if !all.Field(i).Bool() {
			t.Errorf("no attribute writes client.AdminRolePermissions.%s", permissionsType.Field(i).Name)
		}
	}
}
