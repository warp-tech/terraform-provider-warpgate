package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePublicKeyCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePublicKeyCredentialCreate,
		ReadContext:   resourcePublicKeyCredentialRead,
		UpdateContext: resourcePublicKeyCredentialUpdate,
		DeleteContext: resourcePublicKeyCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the user to add the public key credential to",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A label for the public key",
			},
			"public_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The OpenSSH public key",
			},
			"date_added": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the key was added",
			},
			"last_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the key was last used",
			},
		},
	}
}

func resourcePublicKeyCredentialCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	userID := d.Get("user_id").(string)
	label := d.Get("label").(string)
	publicKey := d.Get("public_key").(string)

	cred, err := c.AddPublicKeyCredential(ctx, userID, label, publicKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to add public key credential: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", userID, cred.ID))

	return resourcePublicKeyCredentialRead(ctx, d, meta)
}

func resourcePublicKeyCredentialRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	creds, err := c.GetPublicKeyCredentials(ctx, userID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get public key credentials: %w", err))
	}

	// Find the specific credential
	found := false
	for _, cred := range creds {
		if cred.ID == credID {
			found = true

			if err := d.Set("user_id", userID); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set user_id: %w", err))
			}

			if err := d.Set("label", cred.Label); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set label: %w", err))
			}

			if err := d.Set("public_key", cred.OpensshPublicKey); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set public_key: %w", err))
			}

			if err := d.Set("date_added", cred.DateAdded); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set date_added: %w", err))
			}

			if cred.LastUsed != "" {
				if err := d.Set("last_used", cred.LastUsed); err != nil {
					return diag.FromErr(fmt.Errorf("failed to set last_used: %w", err))
				}
			}

			break
		}
	}

	if !found {
		d.SetId("")
	}

	return diags
}

func resourcePublicKeyCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	// Parse the ID to get user_id and credential_id
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format: %s (expected user_id:credential_id)", d.Id())
	}

	userID := parts[0]
	credID := parts[1]

	label := d.Get("label").(string)
	publicKey := d.Get("public_key").(string)

	_, err := c.UpdatePublicKeyCredential(ctx, userID, credID, label, publicKey)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update public key credential: %w", err))
	}

	return resourcePublicKeyCredentialRead(ctx, d, meta)
}

func resourcePublicKeyCredentialDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	err := c.DeletePublicKeyCredential(ctx, userID, credID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete public key credential: %w", err))
	}

	d.SetId("")

	return diags
}
