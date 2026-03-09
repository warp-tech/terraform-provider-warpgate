package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

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
		},
	}
}

// resourceParametersCreate handles the creation of Warpgate parameters (singleton resource)
func resourceParametersCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	req := &client.ParametersUpdateRequest{
		AllowOwnCredentialManagement: d.Get("allow_own_credential_management").(bool),
	}

	if rateLimitBytesPerSecond, ok := d.GetOk("rate_limit_bytes_per_second"); ok {
		req.RateLimitBytesPerSecond = rateLimitBytesPerSecond.(int)
	}

	if sshClientAuthPublickey, ok := d.GetOk("ssh_client_auth_publickey"); ok {
		req.SSHClientAuthPublickey = sshClientAuthPublickey.(bool)
	}

	if sshClientAuthPassword, ok := d.GetOk("ssh_client_auth_password"); ok {
		req.SSHClientAuthPassword = sshClientAuthPassword.(bool)
	}

	if sshClientAuthKeyboardInteractive, ok := d.GetOk("ssh_client_auth_keyboard_interactive"); ok {
		req.SSHClientAuthKeyboardInteractive = sshClientAuthKeyboardInteractive.(bool)
	}

	if minimizePasswordLogin, ok := d.GetOk("minimize_password_login"); ok {
		req.MinimizePasswordLogin = minimizePasswordLogin.(bool)
	}

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

	return diags
}

// resourceParametersUpdate handles the update of Warpgate parameters
func resourceParametersUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	req := &client.ParametersUpdateRequest{
		AllowOwnCredentialManagement: d.Get("allow_own_credential_management").(bool),
	}

	if rateLimitBytesPerSecond, ok := d.GetOk("rate_limit_bytes_per_second"); ok {
		req.RateLimitBytesPerSecond = rateLimitBytesPerSecond.(int)
	}

	if sshClientAuthPublickey, ok := d.GetOk("ssh_client_auth_publickey"); ok {
		req.SSHClientAuthPublickey = sshClientAuthPublickey.(bool)
	}

	if sshClientAuthPassword, ok := d.GetOk("ssh_client_auth_password"); ok {
		req.SSHClientAuthPassword = sshClientAuthPassword.(bool)
	}

	if sshClientAuthKeyboardInteractive, ok := d.GetOk("ssh_client_auth_keyboard_interactive"); ok {
		req.SSHClientAuthKeyboardInteractive = sshClientAuthKeyboardInteractive.(bool)
	}

	if minimizePasswordLogin, ok := d.GetOk("minimize_password_login"); ok {
		req.MinimizePasswordLogin = minimizePasswordLogin.(bool)
	}

	_, err := c.UpdateParameters(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update parameters: %w", err))
	}

	return resourceParametersRead(ctx, d, meta)
}

// resourceParametersDelete handles the deletion of Warpgate parameters
func resourceParametersDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// For a global parameters resource, we don't actually delete it from the API
	// but we remove it from Terraform state
	d.SetId("")

	return diags
}
