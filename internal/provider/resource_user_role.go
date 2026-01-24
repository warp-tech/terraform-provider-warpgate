// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceUserRole creates and returns a schema for the user-role association resource.
func resourceUserRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserRoleCreate,
		ReadContext:   resourceUserRoleRead,
		DeleteContext: resourceUserRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user to assign the role to",
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

// resourceUserRoleCreate handles the creation of a new user-role association in Warpgate.
func resourceUserRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	roleID := d.Get("role_id").(string)

	err := c.AddUserRole(ctx, userID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign role to user: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, roleID))

	return diags
}

// resourceUserRoleRead retrieves the user-role association data from Warpgate and
// updates the Terraform state accordingly.
func resourceUserRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format: %s (expected user_id:role_id)", id)
	}

	userID := parts[0]
	roleID := parts[1]

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set user_id: %w", err))
	}

	if err := d.Set("role_id", roleID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role_id: %w", err))
	}

	// Check if the role is still assigned to the user
	roles, err := c.GetUserRoles(ctx, userID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get user roles: %w", err))
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

// resourceUserRoleDelete removes a user-role association from Warpgate.
func resourceUserRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	roleID := d.Get("role_id").(string)

	err := c.DeleteUserRole(ctx, userID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to remove role from user: %w", err))
	}

	d.SetId("")

	return diags
}
