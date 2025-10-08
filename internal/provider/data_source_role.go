// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// dataSourceRole creates and returns a schema for the role data source.
func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the role",
				ConflictsWith: []string{},
				AtLeastOneOf:  []string{"id", "name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of the role",
				ConflictsWith: []string{},
				AtLeastOneOf:  []string{"id", "name"},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the role",
			},
		},
	}
}

// dataSourceRoleRead retrieves role data from Warpgate by ID or name filter and populates
// the Terraform state.
func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics
	var role *client.Role

	id, idOk := d.GetOk("id")
	name, nameOk := d.GetOk("name")

	if !idOk && !nameOk {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if nameStr, ok := name.(string); ok && nameStr != "" {
		roles, err := c.GetRoles(ctx, nameStr)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to search roles: %w", err))
		}

		for i := range roles {
			if roles[i].Name == nameStr {
				role = &roles[i]
				break
			}
		}

		if role == nil {
			return diag.Errorf("role with name %s not found", nameStr)
		}
	} else {
		idStr := id.(string)
		var err error
		role, err = c.GetRole(ctx, idStr)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to read role: %w", err))
		}

		if role == nil {
			return diag.Errorf("role with ID %s not found", idStr)
		}
	}

	d.SetId(role.ID)
	if err := d.Set("id", role.ID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set id: %w", err))
	}
	if err := d.Set("name", role.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %w", err))
	}

	if err := d.Set("description", role.Description); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set description: %w", err))
	}

	return diags
}
