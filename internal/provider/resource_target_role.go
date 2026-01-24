// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceTargetRole creates and returns a schema for the target-role association resource.
func resourceTargetRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTargetRoleCreate,
		ReadContext:   resourceTargetRoleRead,
		DeleteContext: resourceTargetRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the target to assign the role to",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the role to assign",
			},
		},
	}
}

// resourceTargetRoleCreate handles the creation of a new target-role association in Warpgate.
func resourceTargetRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	targetID := d.Get("target_id").(string)
	roleID := d.Get("role_id").(string)

	err := c.AddTargetRole(ctx, targetID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign role to target: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", targetID, roleID))

	return diags
}

// resourceTargetRoleRead retrieves the target-role association data from Warpgate and
// updates the Terraform state accordingly.
func resourceTargetRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format: %s (expected target_id:role_id)", id)
	}

	targetID := parts[0]
	roleID := parts[1]

	if err := d.Set("target_id", targetID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set target_id: %w", err))
	}

	if err := d.Set("role_id", roleID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role_id: %w", err))
	}

	// Check if the role is still assigned to the target
	roles, err := c.GetTargetRoles(ctx, targetID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get target roles: %w", err))
	}

	found := false
	for _, role := range roles {
		if role.ID == roleID {
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
	}

	return diags
}

// resourceTargetRoleDelete removes a target-role association from Warpgate.
func resourceTargetRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	targetID := d.Get("target_id").(string)
	roleID := d.Get("role_id").(string)

	err := c.DeleteTargetRole(ctx, targetID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to remove role from target: %w", err))
	}

	d.SetId("")

	return diags
}
