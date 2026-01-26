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
		Description: "Retrieves the SSH host keys for the Warpgate server. These are the server's own SSH keys that clients use to verify the server identity.",
		Schema: map[string]*schema.Schema{
			"keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of SSH host keys",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of SSH key (e.g., 'Ed25519', 'RSA')",
						},
						"public_key_base64": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public key in base64 format",
						},
					},
				},
			},
		},
	}
}

// flattenSSHOwnKeys converts a slice of SSH keys from the Warpgate API format
// to the Terraform schema representation.
func flattenSSHOwnKeys(keys []client.SSHOwnKey) []any {
	if len(keys) == 0 {
		return nil
	}

	result := make([]any, len(keys))
	for i, key := range keys {
		result[i] = map[string]any{
			"kind":              key.Kind,
			"public_key_base64": key.PublicKeyBase64,
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

	// Use a static ID since this data source always returns the same server keys
	d.SetId("ssh-own-keys")

	if err := d.Set("keys", flattenSSHOwnKeys(keys)); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set keys: %w", err))
	}

	return diags
}
