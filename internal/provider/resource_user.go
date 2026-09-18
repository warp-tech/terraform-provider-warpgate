// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// resourceUser creates and returns a schema for the user resource.
func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The username of the user",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the user",
			},
			rateLimitBytesPerSecondKey: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Bandwidth limit in bytes per second",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"credential_policy": credentialPolicySchema("The credential policy for the user"),
			"allowed_ip_ranges": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of allowed IP ranges in CIDR notation. If set, only connections from these IP ranges will be allowed for this user.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		CustomizeDiff: validateUserConfig,
	}
}

// resourceUserCreate handles the creation of a new user in Warpgate based on
// the provided resource data.
func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	username := d.Get("username").(string)
	description := d.Get("description").(string)

	req := &client.UserCreateRequest{
		Username:    username,
		Description: description,
	}

	user, err := c.CreateUser(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create user: %w", err))
	}

	d.SetId(user.ID)

	if needsUserPostCreateUpdate(d) {
		updateReq := buildUserUpdateRequest(d)
		_, err := c.UpdateUser(ctx, user.ID, updateReq)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update user after creation: %w", err))
		}
	}

	return resourceUserRead(ctx, d, meta)
}

// resourceUserRead retrieves the user data from Warpgate and updates the
// Terraform state accordingly.
func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	user, err := c.GetUser(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read user: %w", err))
	}

	// If the user was not found, return nil to indicate that the resource no longer exists
	if user == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("username", user.Username); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set username: %w", err))
	}

	if err := d.Set("description", user.Description); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set description: %w", err))
	}

	if err := setOptionalInt(d, rateLimitBytesPerSecondKey, user.RateLimitBytesPerSecond); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set rate_limit_bytes_per_second: %w", err))
	}

	if user.CredentialPolicy != nil {
		if err := d.Set("credential_policy", flattenCredentialPolicy(user.CredentialPolicy)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set credential_policy: %w", err))
		}
	}

	if user.AllowedIPRanges != nil {
		if err := d.Set("allowed_ip_ranges", *user.AllowedIPRanges); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set allowed_ip_ranges: %w", err))
		}
	}

	return diags
}

// resourceUserUpdate handles the update of an existing user in Warpgate based on
// the provided resource data changes.
func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	id := d.Id()
	req := buildUserUpdateRequest(d)

	_, err := c.UpdateUser(ctx, id, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update user: %w", err))
	}

	return resourceUserRead(ctx, d, meta)
}

// resourceUserDelete removes a user from Warpgate based on the resource data.
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	err := c.DeleteUser(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete user: %w", err))
	}

	d.SetId("")

	return diags
}

func needsUserPostCreateUpdate(d *schema.ResourceData) bool {
	if _, ok := d.GetOk("credential_policy"); ok {
		return true
	}

	if configuredValueExists(d, rateLimitBytesPerSecondKey) {
		return true
	}

	_, ok := d.GetOk("allowed_ip_ranges")
	return ok
}

func buildUserUpdateRequest(d *schema.ResourceData) *client.UserUpdateRequest {
	req := &client.UserUpdateRequest{
		Username:                d.Get("username").(string),
		Description:             d.Get("description").(string),
		RateLimitBytesPerSecond: optionalIntPointer(d, rateLimitBytesPerSecondKey),
		AllowedIPRanges:         expandAllowedIPRanges(d),
	}

	if v, ok := d.GetOk("credential_policy"); ok {
		req.CredentialPolicy = expandCredentialPolicy(v.([]any))
	}

	return req
}

// expandAllowedIPRanges converts the allowed_ip_ranges from Terraform schema to the API format.
func expandAllowedIPRanges(d *schema.ResourceData) *[]string {
	v, ok := d.GetOk("allowed_ip_ranges")
	if !ok || v == nil {
		return nil
	}
	raw := v.([]any)
	ranges := make([]string, len(raw))
	for i, r := range raw {
		ranges[i] = r.(string)
	}
	return &ranges
}

// validateUserConfig validates the user configuration in a Terraform resource diff,
// ensuring that credential policies are correctly formatted.
func validateUserConfig(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if v, ok := d.GetOk("credential_policy"); ok {
		return validateCredentialPolicy("credential_policy", v)
	}

	return nil
}
