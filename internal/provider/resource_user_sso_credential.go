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

// resourceUserSsoCredential creates and returns a schema for the user SSO credential resource.
func resourceUserSsoCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserSsoCredentialCreate,
		ReadContext:   resourceUserSsoCredentialRead,
		UpdateContext: resourceUserSsoCredentialUpdate,
		DeleteContext: resourceUserSsoCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserSsoCredentialImport,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The ID of the user to add the SSO credential to",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"sso_provider": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The SSO provider name (e.g., 'google', 'github', 'okta')",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The email address associated with the SSO provider",
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

// resourceUserSsoCredentialCreate handles the creation of a new SSO credential for a user.
func resourceUserSsoCredentialCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	userID := d.Get("user_id").(string)
	provider := d.Get("sso_provider").(string)
	email := d.Get("email").(string)

	// Verify user exists
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to verify user exists: %w", err))
	}
	if user == nil {
		return diag.Errorf("user with ID %s not found", userID)
	}

	credential, err := c.AddSsoCredential(ctx, userID, provider, email)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create SSO credential: %w", err))
	}

	d.SetId(credential.ID)

	return resourceUserSsoCredentialRead(ctx, d, meta)
}

// resourceUserSsoCredentialRead retrieves the SSO credential data from Warpgate and updates
// the Terraform state accordingly.
func resourceUserSsoCredentialRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	credentialID := d.Id()

	credentials, err := c.GetSsoCredentials(ctx, userID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read SSO credentials: %w", err))
	}

	// Find the specific credential
	var credential *client.SsoCredential
	for _, cred := range credentials {
		if cred.ID == credentialID {
			credential = &cred
			break
		}
	}

	// If the credential was not found, remove it from state
	if credential == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("sso_provider", credential.Provider); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set provider: %w", err))
	}

	if err := d.Set("email", credential.Email); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set email: %w", err))
	}

	return diags
}

// resourceUserSsoCredentialUpdate handles the update of an existing SSO credential.
func resourceUserSsoCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	userID := d.Get("user_id").(string)
	credentialID := d.Id()
	provider := d.Get("sso_provider").(string)
	email := d.Get("email").(string)

	_, err := c.UpdateSsoCredential(ctx, userID, credentialID, provider, email)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update SSO credential: %w", err))
	}

	return resourceUserSsoCredentialRead(ctx, d, meta)
}

// resourceUserSsoCredentialDelete removes an SSO credential from a user.
func resourceUserSsoCredentialDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	credentialID := d.Id()

	err := c.DeleteSsoCredential(ctx, userID, credentialID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete SSO credential: %w", err))
	}

	d.SetId("")

	return diags
}

// resourceUserSsoCredentialImport handles the import of an existing SSO credential.
// The import ID should be in the format "user_id:credential_id".
func resourceUserSsoCredentialImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	userID, credentialID, err := parseCompositeID(d.Id(), "user_id", "credential_id")
	if err != nil {
		return nil, err
	}

	d.SetId(credentialID)
	if err := d.Set("user_id", userID); err != nil {
		return nil, fmt.Errorf("failed to set user_id: %w", err)
	}

	return []*schema.ResourceData{d}, nil
}
