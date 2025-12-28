package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func resourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTargetGroupCreate,
		ReadContext:   resourceTargetGroupRead,
		UpdateContext: resourceTargetGroupUpdate,
		DeleteContext: resourceTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the target group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the target group",
			},
			"color": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The color that the target group should have. Valid values: Primary, Secondary, Success, Danger, Warning, Info, Light, Dark",
			},
		},
	}
}

// resourceTargetGroupCreate handles the creation of a new target group in Warpgate.
func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	color := d.Get("color").(string)

	targetGroup, err := c.CreateTargetGroup(ctx, &client.TargetGroupCreateRequest{
		Name:        name,
		Description: description,
		Color:       color,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to assign role to target: %w", err))
	}

	d.SetId(targetGroup.ID)

	return diags
}

// resourceTargetGroupRead retrieves the target group data from Warpgate and updates the
// Terraform state accordingly.
func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	targetGroup, err := c.GetTargetGroup(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read target group: %w", err))
	}

	// If the target group was not found, return nil to indicate that the resource no longer exists
	if targetGroup == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", targetGroup.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %w", err))
	}

	if err := d.Set("description", targetGroup.Description); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set description: %w", err))
	}

	if err := d.Set("color", targetGroup.Color); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set color: %w", err))
	}

	return diags
}

// resourceTargetGroupUpdate handles updating an existing target group in Warpgate.
func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	color := d.Get("color").(string)

	_, err := c.UpdateTargetGroup(ctx, id, &client.TargetGroupCreateRequest{
		Name:        name,
		Description: description,
		Color:       color,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update target group: %w", err))
	}

	return diags
}

// resourceTargetGroupDelete removes a target group from Warpgate based on the resource data.
func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	err := c.DeleteTargetGroup(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete target group: %w", err))
	}

	d.SetId("")

	return diags
}
