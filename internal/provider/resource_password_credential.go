package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePasswordCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordCredentialCreate,
		ReadContext:   resourcePasswordCredentialRead,
		DeleteContext: resourcePasswordCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user to add the password credential to",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "The password for authentication",
			},
		},
	}
}

func resourcePasswordCredentialCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	userID := d.Get("user_id").(string)
	password := d.Get("password").(string)

	cred, err := c.AddPasswordCredential(ctx, userID, password)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to add password credential: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, cred.ID))

	return diags
}

func resourcePasswordCredentialRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Implementation would check if credential still exists
	// TODO: We can't read back the password but can verify the credential exists
	var diags diag.Diagnostics
	return diags
}

func resourcePasswordCredentialDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	// Parse the ID to get user_id and credential_id
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format: %s (expected user_id:credential_id)", d.Id())
	}

	userID := parts[0]
	credID := parts[1]

	err := c.DeletePasswordCredential(ctx, userID, credID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete password credential: %w", err))
	}

	d.SetId("")

	return diags
}
