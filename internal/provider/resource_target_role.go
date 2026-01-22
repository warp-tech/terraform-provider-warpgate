// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// resourceTargetRole creates and returns a schema for the target-role association resource.
func resourceTargetRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTargetRoleCreate,
		ReadContext:   resourceTargetRoleRead,
		UpdateContext: resourceTargetRoleUpdate,
		DeleteContext: resourceTargetRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the target to assign the role to",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the role to assign",
			},
			// File transfer permissions - only applicable for SSH targets
			"file_transfer": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "File transfer (SCP/SFTP) permissions. Only applicable for SSH targets.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_upload": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "inherit",
							Description: "Allow file uploads via SCP/SFTP. Values: 'inherit' (from role), 'true', 'false'",
							ValidateFunc: func(i interface{}, k string) (ws []string, errors []error) {
								v := i.(string)
								if v != "inherit" && v != "true" && v != "false" {
									errors = append(errors, fmt.Errorf("%q must be 'inherit', 'true', or 'false', got: %s", k, v))
								}
								return ws, errors
							},
						},
						"allow_download": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "inherit",
							Description: "Allow file downloads via SCP/SFTP. Values: 'inherit' (from role), 'true', 'false'",
							ValidateFunc: func(i interface{}, k string) (ws []string, errors []error) {
								v := i.(string)
								if v != "inherit" && v != "true" && v != "false" {
									errors = append(errors, fmt.Errorf("%q must be 'inherit', 'true', or 'false', got: %s", k, v))
								}
								return ws, errors
							},
						},
						"allowed_paths": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Allowed paths for file transfers (null = all paths allowed)",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"blocked_extensions": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Blocked file extensions (null = no extensions blocked)",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"max_file_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum file size in bytes (null = no limit)",
						},
					},
				},
			},
		},
		CustomizeDiff: validateTargetRoleConfig,
	}
}

// validateTargetRoleConfig validates that file_transfer is only set for SSH targets.
func validateTargetRoleConfig(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	// Only validate if file_transfer is set
	if v, ok := d.GetOk("file_transfer"); !ok || len(v.([]any)) == 0 {
		return nil
	}

	// We can't validate target type during plan if target_id is unknown
	// The API will return an error if file_transfer is set for non-SSH targets
	// This is acceptable as it provides clear feedback to the user
	return nil
}

// resourceTargetRoleCreate handles the creation of a new target-role association in Warpgate.
func resourceTargetRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	targetID := d.Get("target_id").(string)
	roleID := d.Get("role_id").(string)

	// First, create the role assignment
	err := c.AddTargetRole(ctx, targetID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign role to target: %w", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", targetID, roleID))

	// If file_transfer is specified, update the permissions
	if v, ok := d.GetOk("file_transfer"); ok && len(v.([]any)) > 0 {
		ftOpts := v.([]any)[0].(map[string]any)
		perm := buildFileTransferPermission(ftOpts)

		_, err := c.UpdateTargetRoleFileTransferPermission(ctx, targetID, roleID, perm)
		if err != nil {
			// Role was created but file transfer update failed
			// Don't delete the role, but report the error
			return diag.FromErr(fmt.Errorf("role assigned successfully, but failed to set file transfer permissions: %w", err))
		}
	}

	return resourceTargetRoleRead(ctx, d, meta)
}

// resourceTargetRoleRead retrieves the target-role association data from Warpgate and
// updates the Terraform state accordingly.
func resourceTargetRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return diag.Errorf("invalid ID format: %s (expected target_id:role_id)", id)
	}

	targetID := parts[0]
	roleID := parts[1]

	if err := d.Set("target_id", targetID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set target_id: %w", err))
	}

	if err := d.Set("role_id", roleID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set role_id: %w", err))
	}

	// Check if the role is still assigned to the target
	roles, err := c.GetTargetRoles(ctx, targetID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get target roles: %w", err))
	}

	found := false
	for _, role := range roles {
		if role.ID == roleID {
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
		return diags
	}

	// Read file transfer permissions
	// This will return data for SSH targets, or nil/error for non-SSH targets
	ftPerm, err := c.GetTargetRoleFileTransferPermission(ctx, targetID, roleID)
	if err != nil {
		// If we get an error, it might be because the target is not SSH
		// Check if file_transfer was configured in state
		if v, ok := d.GetOk("file_transfer"); ok && len(v.([]any)) > 0 {
			// file_transfer was configured but API returned error - this is a problem
			return diag.FromErr(fmt.Errorf("failed to get file transfer permissions: %w", err))
		}
		// file_transfer was not configured, so this is expected for non-SSH targets
		// Just clear the file_transfer block
		if err := d.Set("file_transfer", []any{}); err != nil {
			return diag.FromErr(fmt.Errorf("failed to clear file_transfer: %w", err))
		}
		return diags
	}

	if ftPerm != nil {
		// Convert nullable bools to string representation
		allowUpload := "inherit"
		if ftPerm.AllowFileUpload != nil {
			if *ftPerm.AllowFileUpload {
				allowUpload = "true"
			} else {
				allowUpload = "false"
			}
		}

		allowDownload := "inherit"
		if ftPerm.AllowFileDownload != nil {
			if *ftPerm.AllowFileDownload {
				allowDownload = "true"
			} else {
				allowDownload = "false"
			}
		}

		ftBlock := map[string]any{
			"allow_upload":   allowUpload,
			"allow_download": allowDownload,
		}

		if ftPerm.AllowedPaths != nil {
			ftBlock["allowed_paths"] = ftPerm.AllowedPaths
		}

		if ftPerm.BlockedExtensions != nil {
			ftBlock["blocked_extensions"] = ftPerm.BlockedExtensions
		}

		if ftPerm.MaxFileSize != nil {
			ftBlock["max_file_size"] = *ftPerm.MaxFileSize
		}

		// Only set file_transfer if it was configured or has non-default values
		// This prevents showing file_transfer block for SSH targets where user didn't configure it
		if v, ok := d.GetOk("file_transfer"); ok && len(v.([]any)) > 0 {
			if err := d.Set("file_transfer", []any{ftBlock}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set file_transfer: %w", err))
			}
		} else if allowUpload != "inherit" || allowDownload != "inherit" ||
			ftPerm.AllowedPaths != nil || ftPerm.BlockedExtensions != nil || ftPerm.MaxFileSize != nil {
			// Non-default values exist, show them
			if err := d.Set("file_transfer", []any{ftBlock}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to set file_transfer: %w", err))
			}
		} else {
			// Default values, don't show the block
			if err := d.Set("file_transfer", []any{}); err != nil {
				return diag.FromErr(fmt.Errorf("failed to clear file_transfer: %w", err))
			}
		}
	}

	return diags
}

// resourceTargetRoleUpdate handles updates to the target-role association.
// Only file_transfer permissions can be updated; target_id and role_id are ForceNew.
func resourceTargetRoleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	targetID := d.Get("target_id").(string)
	roleID := d.Get("role_id").(string)

	if d.HasChange("file_transfer") {
		if v, ok := d.GetOk("file_transfer"); ok && len(v.([]any)) > 0 {
			ftOpts := v.([]any)[0].(map[string]any)
			perm := buildFileTransferPermission(ftOpts)

			_, err := c.UpdateTargetRoleFileTransferPermission(ctx, targetID, roleID, perm)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to update file transfer permissions: %w", err))
			}
		} else {
			// file_transfer block was removed, reset to inherit from role defaults
			perm := &client.FileTransferPermission{
				AllowFileUpload:   nil, // inherit from role
				AllowFileDownload: nil, // inherit from role
				AllowedPaths:      nil,
				BlockedExtensions: nil,
				MaxFileSize:       nil,
			}

			_, err := c.UpdateTargetRoleFileTransferPermission(ctx, targetID, roleID, perm)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to reset file transfer permissions: %w", err))
			}
		}
	}

	return resourceTargetRoleRead(ctx, d, meta)
}

// resourceTargetRoleDelete removes a target-role association from Warpgate.
func resourceTargetRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	targetID := d.Get("target_id").(string)
	roleID := d.Get("role_id").(string)

	err := c.DeleteTargetRole(ctx, targetID, roleID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to remove role from target: %w", err))
	}

	d.SetId("")

	return diags
}

// buildFileTransferPermission constructs a FileTransferPermission from the Terraform schema.
// The new inheritance model uses nullable bools: nil = inherit from role, true/false = explicit override.
func buildFileTransferPermission(opts map[string]any) *client.FileTransferPermission {
	perm := &client.FileTransferPermission{}

	// Parse allow_upload: "inherit" -> nil, "true" -> true, "false" -> false
	if v, ok := opts["allow_upload"].(string); ok {
		switch v {
		case "true":
			t := true
			perm.AllowFileUpload = &t
		case "false":
			f := false
			perm.AllowFileUpload = &f
		default: // "inherit" or empty
			perm.AllowFileUpload = nil
		}
	}

	// Parse allow_download: "inherit" -> nil, "true" -> true, "false" -> false
	if v, ok := opts["allow_download"].(string); ok {
		switch v {
		case "true":
			t := true
			perm.AllowFileDownload = &t
		case "false":
			f := false
			perm.AllowFileDownload = &f
		default: // "inherit" or empty
			perm.AllowFileDownload = nil
		}
	}

	if v, ok := opts["allowed_paths"]; ok {
		paths := v.([]any)
		if len(paths) > 0 {
			perm.AllowedPaths = make([]string, len(paths))
			for i, p := range paths {
				perm.AllowedPaths[i] = p.(string)
			}
		}
	}

	if v, ok := opts["blocked_extensions"]; ok {
		exts := v.([]any)
		if len(exts) > 0 {
			perm.BlockedExtensions = make([]string, len(exts))
			for i, e := range exts {
				perm.BlockedExtensions[i] = e.(string)
			}
		}
	}

	if v, ok := opts["max_file_size"]; ok && v.(int) > 0 {
		size := int64(v.(int))
		perm.MaxFileSize = &size
	}

	return perm
}
