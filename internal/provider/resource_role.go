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

// resourceRole creates and returns a schema for the role resource.
func resourceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the role",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the role",
			},
			"file_transfer_defaults": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Default file transfer (SCP/SFTP) permissions for this role. Target-role assignments inherit these unless explicitly overridden.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_upload": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Default permission for file uploads via SCP/SFTP",
						},
						"allow_download": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Default permission for file downloads via SCP/SFTP",
						},
						"allowed_paths": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Default allowed paths for file transfers (null = all paths allowed)",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"blocked_extensions": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Default blocked file extensions (null = no extensions blocked)",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"max_file_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Default maximum file size in bytes (null = no limit)",
						},
					},
				},
			},
		},
	}
}

// resourceRoleCreate handles the creation of a new role in Warpgate based on
// the provided resource data.
func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	req := &client.RoleCreateRequest{
		Name:        name,
		Description: description,
	}

	role, err := c.CreateRole(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create role: %w", err))
	}

	d.SetId(role.ID)

	// If file_transfer_defaults is specified, update the defaults
	if v, ok := d.GetOk("file_transfer_defaults"); ok && len(v.([]any)) > 0 {
		ftOpts := v.([]any)[0].(map[string]any)
		defaults := buildRoleFileTransferDefaults(ftOpts)

		_, err := c.UpdateRoleFileTransferDefaults(ctx, role.ID, defaults)
		if err != nil {
			return diag.FromErr(fmt.Errorf("role created, but failed to set file transfer defaults: %w", err))
		}
	}

	return resourceRoleRead(ctx, d, meta)
}

// resourceRoleRead retrieves the role data from Warpgate and updates the
// Terraform state accordingly.
func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	role, err := c.GetRole(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read role: %w", err))
	}

	// If the role was not found, return nil to indicate that the resource no longer exists
	if role == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", role.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %w", err))
	}

	if err := d.Set("description", role.Description); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set description: %w", err))
	}

	// Read file transfer defaults
	ftDefaults, err := c.GetRoleFileTransferDefaults(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get file transfer defaults: %w", err))
	}

	if ftDefaults != nil {
		ftBlock := map[string]any{
			"allow_upload":   ftDefaults.AllowFileUpload,
			"allow_download": ftDefaults.AllowFileDownload,
		}

		if ftDefaults.AllowedPaths != nil {
			ftBlock["allowed_paths"] = ftDefaults.AllowedPaths
		}

		if ftDefaults.BlockedExtensions != nil {
			ftBlock["blocked_extensions"] = ftDefaults.BlockedExtensions
		}

		if ftDefaults.MaxFileSize != nil {
			ftBlock["max_file_size"] = *ftDefaults.MaxFileSize
		}

		// Only set file_transfer_defaults if it was configured or has non-default values
		if v, ok := d.GetOk("file_transfer_defaults"); ok && len(v.([]any)) > 0 {
			if err := d.Set("file_transfer_defaults", []any{ftBlock}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set file_transfer_defaults: %w", err))
			}
		} else if !ftDefaults.AllowFileUpload || !ftDefaults.AllowFileDownload ||
			ftDefaults.AllowedPaths != nil || ftDefaults.BlockedExtensions != nil || ftDefaults.MaxFileSize != nil {
			// Non-default values exist, show them
			if err := d.Set("file_transfer_defaults", []any{ftBlock}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set file_transfer_defaults: %w", err))
			}
		}
	}

	return diags
}

// resourceRoleUpdate handles the update of an existing role in Warpgate based on
// the provided resource data changes.
func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	id := d.Id()

	// Update name and description if changed
	if d.HasChange("name") || d.HasChange("description") {
		name := d.Get("name").(string)
		description := d.Get("description").(string)

		req := &client.RoleCreateRequest{
			Name:        name,
			Description: description,
		}

		_, err := c.UpdateRole(ctx, id, req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update role: %w", err))
		}
	}

	// Update file transfer defaults if changed
	if d.HasChange("file_transfer_defaults") {
		if v, ok := d.GetOk("file_transfer_defaults"); ok && len(v.([]any)) > 0 {
			ftOpts := v.([]any)[0].(map[string]any)
			defaults := buildRoleFileTransferDefaults(ftOpts)

			_, err := c.UpdateRoleFileTransferDefaults(ctx, id, defaults)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to update file transfer defaults: %w", err))
			}
		} else {
			// file_transfer_defaults block was removed, reset to API defaults
			defaults := &client.RoleFileTransferDefaults{
				AllowFileUpload:   true,
				AllowFileDownload: true,
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			}

			_, err := c.UpdateRoleFileTransferDefaults(ctx, id, defaults)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to reset file transfer defaults: %w", err))
			}
		}
	}

	return resourceRoleRead(ctx, d, meta)
}

// resourceRoleDelete removes a role from Warpgate based on the resource data.
func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	err := c.DeleteRole(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete role: %w", err))
	}

	d.SetId("")

	return diags
}

// buildRoleFileTransferDefaults constructs RoleFileTransferDefaults from the Terraform schema.
func buildRoleFileTransferDefaults(opts map[string]any) *client.RoleFileTransferDefaults {
	defaults := &client.RoleFileTransferDefaults{
		AllowFileUpload:   true, // default
		AllowFileDownload: true, // default
	}

	if v, ok := opts["allow_upload"]; ok {
		defaults.AllowFileUpload = v.(bool)
	}

	if v, ok := opts["allow_download"]; ok {
		defaults.AllowFileDownload = v.(bool)
	}

	if v, ok := opts["allowed_paths"]; ok {
		paths := v.([]any)
		if len(paths) > 0 {
			defaults.AllowedPaths = make([]string, len(paths))
			for i, p := range paths {
				defaults.AllowedPaths[i] = p.(string)
			}
		}
	}

	if v, ok := opts["blocked_extensions"]; ok {
		exts := v.([]any)
		if len(exts) > 0 {
			defaults.BlockedExtensions = make([]string, len(exts))
			for i, e := range exts {
				defaults.BlockedExtensions[i] = e.(string)
			}
		}
	}

	if v, ok := opts["max_file_size"]; ok && v.(int) > 0 {
		size := int64(v.(int))
		defaults.MaxFileSize = &size
	}

	return defaults
}
