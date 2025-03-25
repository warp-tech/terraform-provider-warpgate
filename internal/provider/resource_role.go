// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		},
	}
}

// resourceRoleCreate handles the creation of a new role in Warpgate based on
// the provided resource data.
func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

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

	return diags
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

	return diags
}

// resourceRoleUpdate handles the update of an existing role in Warpgate based on
// the provided resource data changes.
func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	id := d.Id()
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
