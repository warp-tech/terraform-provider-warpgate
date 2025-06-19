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
			"credential_policy": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The credential policy for the user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ssh": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"mysql": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"postgres": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
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

	// Handle credential policy if it's set
	if v, ok := d.GetOk("credential_policy"); ok {
		updateReq := &client.UserUpdateRequest{
			Username:         username,
			Description:      description,
			CredentialPolicy: expandCredentialPolicy(v.([]any)),
		}

		_, err := c.UpdateUser(ctx, user.ID, updateReq)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to set credential policy: %w", err))
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

	if user.CredentialPolicy != nil {
		if err := d.Set("credential_policy", flattenCredentialPolicy(user.CredentialPolicy)); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set credential_policy: %w", err))
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
	username := d.Get("username").(string)
	description := d.Get("description").(string)

	var credentialPolicy *client.UserRequireCredentialsPolicy
	if v, ok := d.GetOk("credential_policy"); ok {
		credentialPolicy = expandCredentialPolicy(v.([]any))
	}

	req := &client.UserUpdateRequest{
		Username:         username,
		Description:      description,
		CredentialPolicy: credentialPolicy,
	}

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

// expandCredentialPolicy converts a Terraform schema representation of credential policy
// to the Warpgate API client structure.
func expandCredentialPolicy(policyList []any) *client.UserRequireCredentialsPolicy {
	if len(policyList) == 0 {
		return nil
	}

	policyMap := policyList[0].(map[string]any)
	policy := &client.UserRequireCredentialsPolicy{}

	if v, ok := policyMap["http"]; ok && v != nil {
		policy.HTTP = expandCredentialKindList(v.([]any))
	}

	if v, ok := policyMap["ssh"]; ok && v != nil {
		policy.SSH = expandCredentialKindList(v.([]any))
	}

	if v, ok := policyMap["mysql"]; ok && v != nil {
		policy.MySQL = expandCredentialKindList(v.([]any))
	}

	if v, ok := policyMap["postgres"]; ok && v != nil {
		policy.Postgres = expandCredentialKindList(v.([]any))
	}

	return policy
}

// expandCredentialKindList converts a list of credential kinds from Terraform schema format
// to the Warpgate API client format.
func expandCredentialKindList(list []any) []client.CredentialKind {
	if len(list) == 0 {
		return nil
	}

	result := make([]client.CredentialKind, len(list))
	for i, v := range list {
		result[i] = client.CredentialKind(v.(string))
	}
	return result
}

// flattenCredentialPolicy converts a Warpgate API credential policy structure
// to the Terraform schema representation.
func flattenCredentialPolicy(policy *client.UserRequireCredentialsPolicy) []any {
	if policy == nil {
		return nil
	}

	result := make(map[string]any)

	if policy.HTTP != nil {
		result["http"] = flattenCredentialKindList(policy.HTTP)
	}

	if policy.SSH != nil {
		result["ssh"] = flattenCredentialKindList(policy.SSH)
	}

	if policy.MySQL != nil {
		result["mysql"] = flattenCredentialKindList(policy.MySQL)
	}

	if policy.Postgres != nil {
		result["postgres"] = flattenCredentialKindList(policy.Postgres)
	}

	return []any{result}
}

// flattenCredentialKindList converts a list of credential kinds from Warpgate API format
// to the Terraform schema format.
func flattenCredentialKindList(list []client.CredentialKind) []any {
	if len(list) == 0 {
		return nil
	}

	result := make([]any, len(list))
	for i, v := range list {
		result[i] = string(v)
	}
	return result
}

// validateUserConfig validates the user configuration in a Terraform resource diff,
// ensuring that credential policies are correctly formatted.
func validateUserConfig(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if v, ok := d.GetOk("credential_policy"); ok {
		credPolicies, ok := v.([]any)
		if !ok || len(credPolicies) == 0 {
			return nil
		}

		policy, ok := credPolicies[0].(map[string]any)
		if !ok {
			return fmt.Errorf("credential_policy must be a map")
		}

		// Valid credential kinds
		validKinds := map[string]bool{
			"Password":        true,
			"PublicKey":       true,
			"Totp":            true,
			"Sso":             true,
			"WebUserApproval": true,
		}

		// Validate each field
		for key, val := range policy {
			// Validate only for known keys
			if key != "http" && key != "ssh" && key != "mysql" && key != "postgres" {
				return fmt.Errorf("unknown credential policy key: %s", key)
			}

			// Ensure the value is a list
			valueList, ok := val.([]any)
			if !ok {
				return fmt.Errorf("credential_policy.%s must be a list", key)
			}

			// Validate each credential kind in the list
			for i, kind := range valueList {
				kindStr, ok := kind.(string)
				if !ok || !validKinds[kindStr] {
					return fmt.Errorf("credential_policy.%s[%d]: %s is not a valid credential kind", key, i, kindStr)
				}
			}
		}
	}

	return nil
}
