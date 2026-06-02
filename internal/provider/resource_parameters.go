package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// TEST
func resourceParameters() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceParametersCreate,
		ReadContext:   resourceParametersRead,
		UpdateContext: resourceParametersUpdate,
		DeleteContext: resourceParametersDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"allow_own_credential_management": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Allow users to manage their own credentials",
			},
			"rate_limit_bytes_per_second": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Global bandwidth limit",
			},
			"ssh_client_auth_publickey": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable SSH public key authentication",
			},
			"ssh_client_auth_password": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable SSH password authentication",
			},
			"ssh_client_auth_keyboard_interactive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable SSH keyboard interactive authentication",
			},
			"minimize_password_login": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "When enabled, the username and password fields are hidden behind a link on the login page, with the focus on the SSO buttons.",
			},
			"ticket_self_service_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable ticket self-service.",
			},
			"ticket_auto_approve_existing_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Automatically approve ticket requests when the requester already has access.",
			},
			"ticket_max_duration_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum ticket duration in seconds.",
			},
			"ticket_max_uses": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum number of uses for tickets.",
			},
			"ticket_require_description": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Require a description for ticket requests.",
			},
			"ticket_request_show_all_targets": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Show all targets when requesting tickets.",
			},
			"target_click_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Action to take when clicking a target.",
				ValidateFunc: validation.StringInSlice([]string{"Connect", "ShowInstructions"}, false),
			},
			"show_session_menu": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "When enabled, Warpgate injects a session menu into HTTP sessions, allowing users to log out or return to the home page.",
			},
			"max_api_token_duration_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum API token duration in seconds.",
			},
			"record_scp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Record SCP sessions.",
			},
		},
	}
}

// resourceParametersCreate handles the creation of Warpgate parameters (singleton resource)
func resourceParametersCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	req := expandParametersUpdateRequest(d)

	_, err := c.UpdateParameters(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create parameters: %w", err))
	}

	// Use a dummy ID for this singleton resource
	d.SetId("parameters")

	return resourceParametersRead(ctx, d, meta)
}

// resourceParametersRead retrieves the parameters from Warpgate and updates the Terraform state
func resourceParametersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	params, err := c.GetParameters(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read parameters: %w", err))
	}

	if params == nil {
		d.SetId("")
		return diags
	}

	// Set dummy ID for singleton resource
	d.SetId("parameters")

	if err := d.Set("allow_own_credential_management", params.AllowOwnCredentialManagement); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set allow_own_credential_management: %w", err))
	}

	if err := d.Set("rate_limit_bytes_per_second", params.RateLimitBytesPerSecond); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set rate_limit_bytes_per_second: %w", err))
	}

	if err := d.Set("ssh_client_auth_publickey", params.SSHClientAuthPublickey); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ssh_client_auth_publickey: %w", err))
	}

	if err := d.Set("ssh_client_auth_password", params.SSHClientAuthPassword); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ssh_client_auth_password: %w", err))
	}

	if err := d.Set("ssh_client_auth_keyboard_interactive", params.SSHClientAuthKeyboardInteractive); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ssh_client_auth_keyboard_interactive: %w", err))
	}

	if err := d.Set("minimize_password_login", params.MinimizePasswordLogin); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set minimize_password_login: %w", err))
	}

	if err := d.Set("ticket_self_service_enabled", params.TicketSelfServiceEnabled); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_self_service_enabled: %w", err))
	}

	if err := d.Set("ticket_auto_approve_existing_access", params.TicketAutoApproveExistingAccess); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_auto_approve_existing_access: %w", err))
	}

	if err := d.Set("ticket_max_duration_seconds", int(params.TicketMaxDurationSeconds)); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_max_duration_seconds: %w", err))
	}

	if err := d.Set("ticket_max_uses", params.TicketMaxUses); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_max_uses: %w", err))
	}

	if err := d.Set("ticket_require_description", params.TicketRequireDescription); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_require_description: %w", err))
	}

	if err := d.Set("ticket_request_show_all_targets", params.TicketRequestShowAllTargets); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_request_show_all_targets: %w", err))
	}

	if err := d.Set("target_click_action", params.TargetClickAction); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set target_click_action: %w", err))
	}

	if err := d.Set("show_session_menu", params.ShowSessionMenu); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set show_session_menu: %w", err))
	}

	if err := d.Set("max_api_token_duration_seconds", int(params.MaxAPITokenDurationSeconds)); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set max_api_token_duration_seconds: %w", err))
	}

	if err := d.Set("record_scp", params.RecordSCP); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set record_scp: %w", err))
	}

	return diags
}

// resourceParametersUpdate handles the update of Warpgate parameters
func resourceParametersUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	req := expandParametersUpdateRequest(d)

	_, err := c.UpdateParameters(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update parameters: %w", err))
	}

	return resourceParametersRead(ctx, d, meta)
}

func expandParametersUpdateRequest(d *schema.ResourceData) *client.ParametersUpdateRequest {
	req := &client.ParametersUpdateRequest{
		AllowOwnCredentialManagement:     d.Get("allow_own_credential_management").(bool),
		SSHClientAuthPublickey:           d.Get("ssh_client_auth_publickey").(bool),
		SSHClientAuthPassword:            d.Get("ssh_client_auth_password").(bool),
		SSHClientAuthKeyboardInteractive: d.Get("ssh_client_auth_keyboard_interactive").(bool),
		MinimizePasswordLogin:            d.Get("minimize_password_login").(bool),
		TicketSelfServiceEnabled:         d.Get("ticket_self_service_enabled").(bool),
		TicketAutoApproveExistingAccess:  d.Get("ticket_auto_approve_existing_access").(bool),
		TicketRequireDescription:         d.Get("ticket_require_description").(bool),
		TicketRequestShowAllTargets:      d.Get("ticket_request_show_all_targets").(bool),
		ShowSessionMenu:                  d.Get("show_session_menu").(bool),
		RecordSCP:                        d.Get("record_scp").(bool),
	}

	if rateLimitBytesPerSecond, ok := d.GetOk("rate_limit_bytes_per_second"); ok {
		req.RateLimitBytesPerSecond = rateLimitBytesPerSecond.(int)
	}

	if ticketMaxDurationSeconds, ok := d.GetOk("ticket_max_duration_seconds"); ok {
		req.TicketMaxDurationSeconds = int64(ticketMaxDurationSeconds.(int))
	}

	if ticketMaxUses, ok := d.GetOk("ticket_max_uses"); ok {
		req.TicketMaxUses = ticketMaxUses.(int)
	}

	if targetClickAction, ok := d.GetOk("target_click_action"); ok {
		req.TargetClickAction = targetClickAction.(string)
	}

	if maxAPITokenDurationSeconds, ok := d.GetOk("max_api_token_duration_seconds"); ok {
		req.MaxAPITokenDurationSeconds = int64(maxAPITokenDurationSeconds.(int))
	}

	return req
}

// resourceParametersDelete handles the deletion of Warpgate parameters
func resourceParametersDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// For a global parameters resource, we don't actually delete it from the API
	// but we remove it from Terraform state
	d.SetId("")

	return diags
}
