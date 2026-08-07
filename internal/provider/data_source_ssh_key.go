// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// dataSourceSSHKey creates and returns a schema for the warpgate_ssh_key data source.
func dataSourceSSHKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSHKeyRead,
		Description: "Retrieves details of a specific Warpgate SSH key by ID or label.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The ID of the SSH key",
				AtLeastOneOf:  []string{"id", "label"},
			},
			"label": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The label of the SSH key",
				AtLeastOneOf:  []string{"id", "label"},
			},
			"kind": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of SSH key (e.g. Ed25519, Rsa4096)",
			},
			"public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public key string",
			},
			"public_key_base64": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The base64 encoded public key body without algorithm prefix",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this key is set as default",
			},
		},
	}
}

func dataSourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics
	var key *client.SSHClientKey

	id, idOk := d.GetOk("id")
	label, labelOk := d.GetOk("label")

	if !idOk && !labelOk {
		return diag.Errorf("either 'id' or 'label' must be specified")
	}

	keys, err := c.GetSSHOwnKeys(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list SSH keys: %w", err))
	}

	if idOk {
		idStr := id.(string)
		for i := range keys {
			if keys[i].ID == idStr {
				key = &keys[i]
				break
			}
		}
		if key == nil {
			return diag.Errorf("SSH key with ID %s not found", idStr)
		}
	} else {
		labelStr := label.(string)
		for i := range keys {
			if keys[i].Label == labelStr {
				key = &keys[i]
				break
			}
		}
		if key == nil {
			return diag.Errorf("SSH key with label %s not found", labelStr)
		}
	}

	d.SetId(key.ID)
	if err := d.Set("label", key.Label); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set label: %w", err))
	}
	if err := d.Set("kind", key.Kind); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set kind: %w", err))
	}
	if err := d.Set("public_key", key.PublicKey); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set public_key: %w", err))
	}
	if err := d.Set("public_key_base64", key.PublicKeyBase64); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set public_key_base64: %w", err))
	}
	if err := d.Set("is_default", key.IsDefault); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set is_default: %w", err))
	}

	return diags
}
