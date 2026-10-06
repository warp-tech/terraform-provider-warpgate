// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceUserAdminRole creates and returns a schema for the user-admin-role
// association resource.
func resourceUserAdminRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserAdminRoleCreate,
		ReadContext:   resourceUserAdminRoleRead,
		DeleteContext: resourceUserAdminRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user to assign the admin role to",
			},
			"admin_role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the admin role to assign",
			},
		},
	}
}

// resourceUserAdminRoleCreate handles the creation of a new user-admin-role
// association in Warpgate.
func resourceUserAdminRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	roleID := d.Get("admin_role_id").(string)

	err := c.AddUserAdminRole(ctx, userID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign admin role to user: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, roleID))

	return diags
}

// resourceUserAdminRoleRead checks that the admin role is still assigned to
// the user and updates the Terraform state accordingly.
func resourceUserAdminRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format: %s (expected user_id:admin_role_id)", id)
	}

	userID := parts[0]
	roleID := parts[1]

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set user_id: %w", err))
	}

	if err := d.Set("admin_role_id", roleID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set admin_role_id: %w", err))
	}

	// A missing user comes back as no roles, which also removes the assignment
	roles, err := c.GetUserAdminRoles(ctx, userID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get user admin roles: %w", err))
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

// resourceUserAdminRoleDelete removes a user-admin-role association from
// Warpgate.
func resourceUserAdminRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	roleID := d.Get("admin_role_id").(string)

	err := c.DeleteUserAdminRole(ctx, userID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to remove admin role from user: %w", err))
	}

	d.SetId("")

	return diags
}
