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
				Description: "The description of the target group",
			},
			"color": {
				Type:        schema.TypeString,
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
