// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// dataSourceUser creates and returns a schema for the user data source.
func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the user",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The username of the user",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the user",
			},
			"credential_policy": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The credential policy for the user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ssh": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"mysql": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"postgres": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"sso_credentials": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The SSO credentials associated with the user",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the SSO credential",
						},
						"provider": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SSO provider name",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email address associated with the SSO provider",
						},
					},
				},
			},
		},
	}
}

// flattenSsoCredentials converts a slice of SSO credentials from the Warpgate API format
// to the Terraform schema representation.
func flattenSsoCredentials(credentials []client.SsoCredential) []any {
	if len(credentials) == 0 {
		return nil
	}

	result := make([]any, len(credentials))
	for i, cred := range credentials {
		result[i] = map[string]any{
			"id":       cred.ID,
			"provider": cred.Provider,
			"email":    cred.Email,
		}
	}
	return result
}

// dataSourceUserRead retrieves user data from Warpgate by ID and populates
// the Terraform state.
func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Get("id").(string)

	user, err := c.GetUser(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read user: %w", err))
	}

	if user == nil {
		return diag.Errorf("user with ID %s not found", id)
	}

	d.SetId(user.ID)
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

	// Get SSO credentials for the user
	ssoCredentials, err := c.GetSsoCredentials(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read SSO credentials: %w", err))
	}

	if err := d.Set("sso_credentials", flattenSsoCredentials(ssoCredentials)); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set sso_credentials: %w", err))
	}

	return diags
}
