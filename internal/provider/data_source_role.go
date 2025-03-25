// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceRole creates and returns a schema for the role data source.
func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the role",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the role",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the role",
			},
		},
	}
}

// dataSourceRoleRead retrieves role data from Warpgate by ID and populates
// the Terraform state.
func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Get("id").(string)

	role, err := c.GetRole(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read role: %w", err))
	}

	if role == nil {
		return diag.Errorf("role with ID %s not found", id)
	}

	d.SetId(role.ID)
	if err := d.Set("name", role.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %w", err))
	}

	if err := d.Set("description", role.Description); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set description: %w", err))
	}

	return diags
}
