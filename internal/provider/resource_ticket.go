package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// resourceRole creates and returns a schema for the role resource.
func resourceTicket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTicketCreate,
		ReadContext:   resourceTicketRead,
		DeleteContext: resourceTicketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user associated with the ticket. Will determine the permissions and access rights.",
			},
			"target_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the target the ticket grants access to.",
			},
			"expiry": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The expiry time of the ticket.",
			},
			"number_of_uses": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The number of uses allowed for the ticket before it becomes invalid.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The description of the ticket.",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The secret value of the ticket used for authentication.",
			},
		},
	}
}

// resourceTicketCreate handles the creation of a new ticket in Warpgate based on
// the provided resource data.
func resourceTicketCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	username := d.Get("username").(string)
	targetName := d.Get("target_name").(string)
	expiry := d.Get("expiry").(string)
	numberOfUses := d.Get("number_of_uses").(int)
	description := d.Get("description").(string)

	req := &client.TicketCreateRequest{
		Username:     username,
		TargetName:   targetName,
		Expiry:       expiry,
		NumberOfUses: numberOfUses,
		Description:  description,
	}

	ticket, err := c.CreateTicket(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create role: %w", err))
	}

	d.SetId(ticket.Ticket.ID)
	if err := d.Set("secret", ticket.Secret); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set secret: %w", err))
	}

	return diags
}

// resourceTicketRead retrieves the ticket data from Warpgate and updates the
// Terraform state accordingly.
func resourceTicketRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// TODO: We do not refresh from warpgate yet as warpgate does not yet support a GET ticket endpoint.
	// So for now we just use the existing state information and don't refresh it.

	var diags diag.Diagnostics

	return diags
}

// resourceTicketDelete removes a ticket from Warpgate based on the resource data.
func resourceTicketDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	err := c.DeleteTicket(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete ticket: %w", err))
	}

	return diags
}
