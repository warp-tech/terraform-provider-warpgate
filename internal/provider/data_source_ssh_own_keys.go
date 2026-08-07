// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// dataSourceSSHOwnKeys creates and returns a schema for the SSH own keys data source.
func dataSourceSSHOwnKeys() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSHOwnKeysRead,
		Description: "Retrieves the SSH keys for the Warpgate server.",
		Schema: map[string]*schema.Schema{
			"keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of SSH keys",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the SSH key",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The label of the SSH key",
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of SSH key (e.g., 'Ed25519', 'RSA')",
						},
						"public_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public key string",
						},
						"public_key_base64": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public key in base64 format",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this key is set as default",
						},
					},
				},
			},
		},
	}
}

// flattenSSHOwnKeys converts a slice of SSH keys from the Warpgate API format
// to the Terraform schema representation.
func flattenSSHOwnKeys(keys []client.SSHClientKey) []any {
	if len(keys) == 0 {
		return nil
	}

	result := make([]any, len(keys))
	for i, key := range keys {
		result[i] = map[string]any{
			"id":                key.ID,
			"label":             key.Label,
			"kind":              key.Kind,
			"public_key":        key.PublicKey,
			"public_key_base64": key.PublicKeyBase64,
			"is_default":        key.IsDefault,
		}
	}
	return result
}

// dataSourceSSHOwnKeysRead retrieves SSH host keys from Warpgate and populates
// the Terraform state.
func dataSourceSSHOwnKeysRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	keys, err := c.GetSSHOwnKeys(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read SSH own keys: %w", err))
	}

	// Use a static ID since this data source always returns the server keys list
	d.SetId("ssh-own-keys")

	if err := d.Set("keys", flattenSSHOwnKeys(keys)); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set keys: %w", err))
	}

	return diags
}
