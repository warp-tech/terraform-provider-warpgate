package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// resourceUserRole creates and returns a schema for user-role association resource.
func resourceUserRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserRoleCreate,
		ReadContext:   resourceUserRoleRead,
		UpdateContext: resourceUserRoleUpdate,
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
			"expiry": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Expiry configuration for the role assignment",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expires_at": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "When the role assignment expires (RFC3339 timestamp)",
							ValidateFunc: func(i interface{}, k string) (ws []string, errors []error) {
								v := i.(string)
								if v == "" {
									return nil, nil
								}
								if _, err := time.Parse(time.RFC3339, v); err != nil {
									errors = append(errors, fmt.Errorf("%q: invalid RFC3339 timestamp: %s", k, err))
								}
								return ws, errors
							},
						},
					},
				},
			},
		},
	}
}

// resourceUserRoleCreate handles creation of a new user-role association in Warpgate.
func resourceUserRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	roleID := d.Get("role_id").(string)

	var expiresAt *string = nil
	if expiryList, ok := d.GetOk("expiry"); ok && len(expiryList.([]interface{})) > 0 {
		if expiry, ok := expiryList.([]interface{})[0].(map[string]interface{}); ok {
			if val, ok := expiry["expires_at"].(string); ok && val != "" {
				expiresAt = &val
			}
		}
	}

	assignment, err := c.AddUserRole(ctx, userID, roleID, expiresAt)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign role to user: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, roleID))
	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set user_id: %w", err))
	}
	if err := d.Set("role_id", roleID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role_id: %w", err))
	}
	if assignment.ExpiresAt != nil {
		if err := d.Set("expiry", []interface{}{
			map[string]interface{}{
				"expires_at": *assignment.ExpiresAt,
			},
		}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set expiry: %w", err))
		}
	}

	return diags
}

// resourceUserRoleRead retrieves user-role association data from Warpgate and
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

	// Get user role assignment details
	assignment, err := c.GetUserRole(ctx, userID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get user role: %w", err))
	}

	if assignment == nil {
		d.SetId("")
		return diags
	}

	// Set expiry info if it exists
	if assignment.ExpiresAt != nil {
		if err := d.Set("expiry", []interface{}{
			map[string]interface{}{
				"expires_at": *assignment.ExpiresAt,
			},
		}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set expiry: %w", err))
		}
	} else {
		if err := d.Set("expiry", []interface{}{}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set expiry: %w", err))
		}
	}

	return diags
}

// resourceUserRoleUpdate handles updating an existing user-role association in Warpgate.
func resourceUserRoleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	var expiresAt *string = nil
	if expiryList, ok := d.GetOk("expiry"); ok && len(expiryList.([]interface{})) > 0 {
		if expiry, ok := expiryList.([]interface{})[0].(map[string]interface{}); ok {
			if val, ok := expiry["expires_at"].(string); ok {
				expiresAt = &val
			}
		}
	}

	assignment, err := c.UpdateUserRoleExpiry(ctx, userID, roleID, expiresAt)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update user role expiry: %w", err))
	}

	if err := d.Set("user_id", userID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set user_id: %w", err))
	}

	if err := d.Set("role_id", roleID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role_id: %w", err))
	}

	if assignment.ExpiresAt != nil {
		if err := d.Set("expiry", []interface{}{
			map[string]interface{}{
				"expires_at": *assignment.ExpiresAt,
			},
		}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set expiry: %w", err))
		}
	} else {
		if err := d.Set("expiry", []interface{}{}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set expiry: %w", err))
		}
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
