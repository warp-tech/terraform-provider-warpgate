// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// dataSourceAdminRole creates and returns a schema for the admin role data
// source.
func dataSourceAdminRole() *schema.Resource {
	s := map[string]*schema.Schema{
		"id": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "The ID of the admin role",
			AtLeastOneOf: []string{"id", "name"},
		},
		"name": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "The name of the admin role",
			AtLeastOneOf: []string{"id", "name"},
		},
		"description": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The description of the admin role",
		},
	}
	for key, permission := range adminRolePermissionsSchema(true) {
		s[key] = permission
	}

	return &schema.Resource{
		ReadContext: dataSourceAdminRoleRead,
		Schema:      s,
	}
}

// dataSourceAdminRoleRead retrieves admin role data from Warpgate by ID or name
// and populates the Terraform state.
func dataSourceAdminRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics
	var role *client.AdminRole

	id, idOk := d.GetOk("id")
	name, nameOk := d.GetOk("name")

	if !idOk && !nameOk {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if nameStr, ok := name.(string); ok && nameStr != "" {
		roles, err := c.GetAdminRoles(ctx, nameStr)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to search admin roles: %w", err))
		}

		for i := range roles {
			if roles[i].Name == nameStr {
				role = &roles[i]
				break
			}
		}

		if role == nil {
			return diag.Errorf("admin role with name %s not found", nameStr)
		}
	} else {
		idStr := id.(string)
		var err error
		role, err = c.GetAdminRole(ctx, idStr)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to read admin role: %w", err))
		}

		if role == nil {
			return diag.Errorf("admin role with ID %s not found", idStr)
		}
	}

	d.SetId(role.ID)
	if err := d.Set("id", role.ID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set id: %w", err))
	}

	if err := setAdminRoleData(d, role); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
