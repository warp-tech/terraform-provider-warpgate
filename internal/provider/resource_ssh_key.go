// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// resourceSSHKey creates and returns a schema for the warpgate_ssh_key resource.
func resourceSSHKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		UpdateContext: resourceSSHKeyUpdate,
		DeleteContext: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Label identifying this SSH key",
			},
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The key type for key generation (e.g. Ed25519, Rsa4096). Defaults to Ed25519 when generating.",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Private key in OpenSSH or PKCS#8 PEM format for importing an existing key. If not provided, a new key pair will be generated.",
			},
			"public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public key string",
			},
			"public_key_base64": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The base64-encoded public key body without algorithm prefix",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether this SSH key is marked as default for host authentication",
			},
		},
	}
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	label := d.Get("label").(string)
	isDefault := d.Get("is_default").(bool)

	var key *client.SSHClientKey
	var err error

	if secretKey, ok := d.GetOk("secret_key"); ok && secretKey.(string) != "" {
		req := &client.ImportSSHClientKeyRequest{
			Label:     label,
			SecretKey: secretKey.(string),
			IsDefault: isDefault,
		}
		key, err = c.ImportSSHOwnKey(ctx, req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to import SSH key: %w", err))
		}
	} else {
		kind := "Ed25519"
		if k, ok := d.GetOk("kind"); ok && k.(string) != "" {
			kind = k.(string)
		}
		req := &client.GenerateSSHClientKeyRequest{
			Label: label,
			Kind:  kind,
		}
		key, err = c.GenerateSSHOwnKey(ctx, req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to generate SSH key: %w", err))
		}

		if isDefault {
			key, err = c.UpdateSSHOwnKey(ctx, key.ID, &client.UpdateSSHClientKeyRequest{
				Label:     label,
				IsDefault: true,
			})
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to set SSH key as default: %w", err))
			}
		}
	}

	d.SetId(key.ID)

	return resourceSSHKeyRead(ctx, d, meta)
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	key, err := c.GetSSHOwnKey(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read SSH key: %w", err))
	}

	if key == nil {
		d.SetId("")
		return diags
	}

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

func resourceSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	id := d.Id()
	label := d.Get("label").(string)
	isDefault := d.Get("is_default").(bool)

	req := &client.UpdateSSHClientKeyRequest{
		Label:     label,
		IsDefault: isDefault,
	}

	_, err := c.UpdateSSHOwnKey(ctx, id, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update SSH key: %w", err))
	}

	return resourceSSHKeyRead(ctx, d, meta)
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	err := c.DeleteSSHOwnKey(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete SSH key: %w", err))
	}

	d.SetId("")
	return diags
}
