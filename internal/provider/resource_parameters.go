package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

func resourceParameters() *schema.Resource {
	intAtLeastZero := validation.ToDiagFunc(validation.IntAtLeast(0))

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
			"rate_limit_bytes_per_second": optionalIntParameter(
				"Global bandwidth limit",
				intAtLeastZero,
			),
			"ssh_client_auth_publickey": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable SSH public key authentication",
			},
			"ssh_client_auth_password": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable SSH password authentication",
			},
			"ssh_client_auth_keyboard_interactive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable SSH keyboard interactive authentication",
			},
			"password_login_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "How the password login form is presented on the gateway login page.",
				ValidateFunc: validation.StringInSlice([]string{"Enabled", "Minimized", "Disabled"}, false),
			},
			"ticket_self_service_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable ticket self-service.",
			},
			"ticket_auto_approve_existing_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Automatically approve ticket requests when the requester already has access.",
			},
			"ticket_max_duration_seconds": optionalIntParameter(
				"Maximum ticket duration in seconds.",
				intAtLeastZero,
			),
			"ticket_max_uses": optionalIntParameter(
				"Maximum number of uses for tickets.",
				validation.ToDiagFunc(validation.IntBetween(0, 32767)),
			),
			"ticket_require_description": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Require a description for ticket requests.",
			},
			"ticket_request_show_all_targets": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Show all targets when requesting tickets.",
			},
			"target_click_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Action to take when clicking a target.",
				ValidateFunc: validation.StringInSlice([]string{"Connect", "ShowInstructions"}, false),
			},
			"show_session_menu": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "When enabled, Warpgate injects a session menu into HTTP sessions, allowing users to log out or return to the home page.",
			},
			"password_policy": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Password policy rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_length": optionalIntParameter(
							"Minimum number of characters, or 0 for no requirement.",
							intAtLeastZero,
						),
						"require_uppercase": optionalComputedBoolParameter("Require at least one uppercase character."),
						"require_lowercase": optionalComputedBoolParameter("Require at least one lowercase character."),
						"require_digits":    optionalComputedBoolParameter("Require at least one digit."),
						"require_special":   optionalComputedBoolParameter("Require at least one special character."),
					},
				},
			},
			"max_api_token_duration_seconds": optionalIntParameter(
				"Maximum API token duration in seconds.",
				intAtLeastZero,
			),
			"record_scp": optionalComputedBoolParameter("Record SCP sessions."),
			"login_protection_enabled": optionalComputedBoolParameter(
				"Enable login protection.",
			),
			"login_protection_retention_seconds": optionalIntParameter(
				"How long login protection records are retained, in seconds.",
				intAtLeastZero,
			),
			"lp_ip_max_attempts": optionalIntParameter(
				"Maximum failed login attempts per IP address.",
				intAtLeastZero,
			),
			"lp_ip_time_window_seconds": optionalIntParameter(
				"Time window for failed login attempts per IP address, in seconds.",
				intAtLeastZero,
			),
			"lp_ip_base_block_duration_seconds": optionalIntParameter(
				"Base IP block duration in seconds.",
				intAtLeastZero,
			),
			"lp_ip_block_duration_multiplier": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				Description:  "Multiplier applied to repeated IP block durations.",
				ValidateFunc: validation.FloatAtLeast(0),
			},
			"lp_ip_max_block_duration_seconds": optionalIntParameter(
				"Maximum IP block duration in seconds.",
				intAtLeastZero,
			),
			"lp_ip_cooldown_reset_seconds": optionalIntParameter(
				"Cooldown period before the IP block escalation resets, in seconds.",
				intAtLeastZero,
			),
			"lp_user_max_attempts": optionalIntParameter(
				"Maximum failed login attempts per user.",
				intAtLeastZero,
			),
			"lp_user_time_window_seconds": optionalIntParameter(
				"Time window for failed login attempts per user, in seconds.",
				intAtLeastZero,
			),
			"lp_user_auto_unlock": optionalComputedBoolParameter(
				"Automatically unlock users after the lockout duration.",
			),
			"lp_user_lockout_duration_seconds": optionalIntParameter(
				"User lockout duration in seconds.",
				intAtLeastZero,
			),
			"lp_user_exempt_admins": optionalComputedBoolParameter(
				"Exempt administrators from user login protection lockouts.",
			),
			"banner": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Banner shown to clients before authentication.",
			},
			"web_clients_enabled": optionalComputedBoolParameter(
				"Enable the web-based session clients.",
			),
			"ssh_host_key_verification": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "What to do when a target's SSH host key isn't in the known hosts list.",
				ValidateFunc: validation.StringInSlice([]string{"Prompt", "AutoAccept", "AutoReject", "Ignore"}, false),
			},
			// 0 is rejected: the API distinguishes unset (reauthentication never
			// required) from 0 (reauthentication required on every action), and
			// the unset value reads back as 0, inviting the wrong meaning.
			"web_auth_max_age_seconds": optionalIntParameter(
				"How long a web login stays valid before reauthentication is required, in seconds. Unset means reauthentication is never required.",
				validation.ToDiagFunc(validation.IntAtLeast(1)),
			),
			"web_approval_grace_period_seconds": optionalIntParameter(
				"How long a remembered web approval stays valid, in seconds.",
				intAtLeastZero,
			),
			"recordings_enable": optionalComputedBoolParameter("Record sessions."),
			"recordings_storage": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Where session recordings are stored.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"recordings_storage.0.disk",
								"recordings_storage.0.s3",
							},
							Description: "Local filesystem storage",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Directory recordings are written to",
									},
								},
							},
						},
						"s3": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"recordings_storage.0.disk",
								"recordings_storage.0.s3",
							},
							Description: "S3 or S3-compatible object storage",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Bucket name",
									},
									"region": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Bucket region",
									},
									"endpoint": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Custom endpoint for S3-compatible services. Empty means AWS.",
									},
									"path_style": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Path-style addressing, required by most S3-compatible services",
									},
									"prefix": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Key prefix prepended to every object path",
									},
									"auto_credentials": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										ExactlyOneOf: []string{
											"recordings_storage.0.s3.0.auto_credentials",
											"recordings_storage.0.s3.0.static_credentials",
										},
										Description: "Authenticate with the ambient AWS credential chain",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"static_credentials": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										ExactlyOneOf: []string{
											"recordings_storage.0.s3.0.auto_credentials",
											"recordings_storage.0.s3.0.static_credentials",
										},
										Description: "Authenticate with an explicit key pair",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_key_id": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Access key ID",
												},
												"secret_access_key": {
													Type:        schema.TypeString,
													Optional:    true,
													Sensitive:   true,
													Description: "Secret access key. The API never returns it, so it is carried over from configuration on read; omit to keep the secret already stored in Warpgate.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"analytics_consent": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Whether the instance reports anonymous usage analytics.",
				ValidateFunc: validation.StringInSlice([]string{"Undecided", "Off", "On"}, false),
			},
			"analytics_normal": optionalComputedBoolParameter(
				"Enable the normal analytics payload level.",
			),
		},
	}
}

func optionalComputedBoolParameter(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}

func optionalIntParameter(description string, validateFunc schema.SchemaValidateDiagFunc) *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		Description:      description,
		ValidateDiagFunc: validateFunc,
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

	d.SetId("parameters")

	for _, field := range []struct {
		name  string
		value any
	}{
		{"allow_own_credential_management", params.AllowOwnCredentialManagement},
		{"rate_limit_bytes_per_second", params.RateLimitBytesPerSecond},
		{"ssh_client_auth_publickey", params.SSHClientAuthPublickey},
		{"ssh_client_auth_password", params.SSHClientAuthPassword},
		{"ssh_client_auth_keyboard_interactive", params.SSHClientAuthKeyboardInteractive},
		{"password_login_mode", params.PasswordLoginMode},
		{"ticket_self_service_enabled", params.TicketSelfServiceEnabled},
		{"ticket_auto_approve_existing_access", params.TicketAutoApproveExistingAccess},
		{"ticket_max_duration_seconds", int(params.TicketMaxDurationSeconds)},
		{"ticket_max_uses", params.TicketMaxUses},
		{"ticket_require_description", params.TicketRequireDescription},
		{"ticket_request_show_all_targets", params.TicketRequestShowAllTargets},
		{"target_click_action", params.TargetClickAction},
		{"show_session_menu", params.ShowSessionMenu},
		{"password_policy", flattenPasswordPolicy(params.PasswordPolicy)},
		{"max_api_token_duration_seconds", int(params.MaxAPITokenDurationSeconds)},
		{"record_scp", params.RecordSCP},
		{"login_protection_enabled", params.LoginProtectionEnabled},
		{"login_protection_retention_seconds", params.LoginProtectionRetentionSeconds},
		{"lp_ip_max_attempts", params.LPIPMaxAttempts},
		{"lp_ip_time_window_seconds", params.LPIPTimeWindowSeconds},
		{"lp_ip_base_block_duration_seconds", params.LPIPBaseBlockDurationSeconds},
		{"lp_ip_block_duration_multiplier", params.LPIPBlockDurationMultiplier},
		{"lp_ip_max_block_duration_seconds", params.LPIPMaxBlockDurationSeconds},
		{"lp_ip_cooldown_reset_seconds", params.LPIPCooldownResetSeconds},
		{"lp_user_max_attempts", params.LPUserMaxAttempts},
		{"lp_user_time_window_seconds", params.LPUserTimeWindowSeconds},
		{"lp_user_auto_unlock", params.LPUserAutoUnlock},
		{"lp_user_lockout_duration_seconds", params.LPUserLockoutDurationSeconds},
		{"lp_user_exempt_admins", params.LPUserExemptAdmins},
		{"banner", params.Banner},
		{"web_clients_enabled", params.WebClientsEnabled},
		{"analytics_consent", params.AnalyticsConsent},
		{"analytics_normal", params.AnalyticsNormal},
		{"ssh_host_key_verification", params.SSHHostKeyVerification},
		{"web_auth_max_age_seconds", int(params.WebAuthMaxAgeSeconds)},
		{"web_approval_grace_period_seconds", int(params.WebApprovalGracePeriodSeconds)},
		{"recordings_enable", params.RecordingsEnable},
		{"recordings_storage", flattenRecordingsStorage(d, params.RecordingsStorage)},
	} {
		if err := d.Set(field.name, field.value); err != nil {
			return diag.FromErr(fmt.Errorf("failed to set %s: %w", field.name, err))
		}
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
		RateLimitBytesPerSecond:          optionalIntPointer(d, "rate_limit_bytes_per_second"),
		SSHClientAuthPublickey:           optionalBoolPointer(d, "ssh_client_auth_publickey"),
		SSHClientAuthPassword:            optionalBoolPointer(d, "ssh_client_auth_password"),
		SSHClientAuthKeyboardInteractive: optionalBoolPointer(d, "ssh_client_auth_keyboard_interactive"),
		TicketSelfServiceEnabled:         optionalBoolPointer(d, "ticket_self_service_enabled"),
		TicketAutoApproveExistingAccess:  optionalBoolPointer(d, "ticket_auto_approve_existing_access"),
		TicketMaxDurationSeconds:         optionalInt64Pointer(d, "ticket_max_duration_seconds"),
		TicketMaxUses:                    optionalIntPointer(d, "ticket_max_uses"),
		TicketRequireDescription:         optionalBoolPointer(d, "ticket_require_description"),
		TicketRequestShowAllTargets:      optionalBoolPointer(d, "ticket_request_show_all_targets"),
		TargetClickAction:                optionalStringPointer(d, "target_click_action"),
		ShowSessionMenu:                  optionalBoolPointer(d, "show_session_menu"),
		PasswordPolicy:                   expandPasswordPolicy(d),
		MaxAPITokenDurationSeconds:       optionalInt64Pointer(d, "max_api_token_duration_seconds"),
		RecordSCP:                        optionalBoolPointer(d, "record_scp"),
		LoginProtectionEnabled:           optionalBoolPointer(d, "login_protection_enabled"),
		LoginProtectionRetentionSeconds:  optionalIntPointer(d, "login_protection_retention_seconds"),
		LPIPMaxAttempts:                  optionalIntPointer(d, "lp_ip_max_attempts"),
		LPIPTimeWindowSeconds:            optionalIntPointer(d, "lp_ip_time_window_seconds"),
		LPIPBaseBlockDurationSeconds:     optionalIntPointer(d, "lp_ip_base_block_duration_seconds"),
		LPIPBlockDurationMultiplier:      optionalFloat64Pointer(d, "lp_ip_block_duration_multiplier"),
		LPIPMaxBlockDurationSeconds:      optionalIntPointer(d, "lp_ip_max_block_duration_seconds"),
		LPIPCooldownResetSeconds:         optionalIntPointer(d, "lp_ip_cooldown_reset_seconds"),
		LPUserMaxAttempts:                optionalIntPointer(d, "lp_user_max_attempts"),
		LPUserTimeWindowSeconds:          optionalIntPointer(d, "lp_user_time_window_seconds"),
		LPUserAutoUnlock:                 optionalBoolPointer(d, "lp_user_auto_unlock"),
		LPUserLockoutDurationSeconds:     optionalIntPointer(d, "lp_user_lockout_duration_seconds"),
		LPUserExemptAdmins:               optionalBoolPointer(d, "lp_user_exempt_admins"),
		Banner:                           optionalStringPointer(d, "banner"),
		WebClientsEnabled:                optionalBoolPointer(d, "web_clients_enabled"),
		AnalyticsConsent:                 optionalStringPointer(d, "analytics_consent"),
		AnalyticsNormal:                  optionalBoolPointer(d, "analytics_normal"),
		SSHHostKeyVerification:           optionalStringPointer(d, "ssh_host_key_verification"),
		WebAuthMaxAgeSeconds:             optionalInt64Pointer(d, "web_auth_max_age_seconds"),
		WebApprovalGracePeriodSeconds:    optionalInt64Pointer(d, "web_approval_grace_period_seconds"),
		RecordingsEnable:                 optionalBoolPointer(d, "recordings_enable"),
		RecordingsStorage:                expandRecordingsStorage(d),
	}

	if passwordLoginMode := optionalStringPointer(d, "password_login_mode"); passwordLoginMode != nil {
		req.PasswordLoginMode = passwordLoginMode
	}

	return req
}

func flattenRecordingsStorage(d *schema.ResourceData, cfg client.RecordingsStorageConfig) []any {
	switch cfg.Kind {
	case "Disk":
		return []any{
			map[string]any{
				"disk": []any{
					map[string]any{"path": stringValue(cfg.Path)},
				},
			},
		}
	case "S3":
		s3 := map[string]any{
			"bucket":     stringValue(cfg.Bucket),
			"region":     stringValue(cfg.Region),
			"endpoint":   stringValue(cfg.Endpoint),
			"path_style": boolValue(cfg.PathStyle),
			"prefix":     stringValue(cfg.Prefix),
		}

		if cfg.Credentials != nil {
			switch cfg.Credentials.Mode {
			case "Static":
				s3["static_credentials"] = []any{
					map[string]any{
						"access_key_id": stringValue(cfg.Credentials.AccessKeyID),
						// The API redacts the secret, so echoing the response
						// would wipe it from state on every read.
						"secret_access_key": d.Get("recordings_storage.0.s3.0.static_credentials.0.secret_access_key"),
					},
				}
			case "Auto":
				s3["auto_credentials"] = []any{map[string]any{}}
			}
		}

		return []any{map[string]any{"s3": []any{s3}}}
	default:
		return nil
	}
}

func expandRecordingsStorage(d *schema.ResourceData) *client.RecordingsStorageConfig {
	if !configuredValueExists(d, "recordings_storage") {
		return nil
	}

	raw := d.Get("recordings_storage").([]any)
	if len(raw) == 0 || raw[0] == nil {
		return nil
	}

	return buildRecordingsStorage(raw[0].(map[string]any))
}

func buildRecordingsStorage(storage map[string]any) *client.RecordingsStorageConfig {
	if disk, ok := firstBlock(storage, "disk"); ok {
		path := blockString(disk, "path")

		return &client.RecordingsStorageConfig{Kind: "Disk", Path: &path}
	}

	s3, ok := firstBlock(storage, "s3")
	if !ok {
		return nil
	}

	bucket := blockString(s3, "bucket")
	region := blockString(s3, "region")
	pathStyle := blockBool(s3, "path_style")
	prefix := blockString(s3, "prefix")

	cfg := &client.RecordingsStorageConfig{
		Kind:      "S3",
		Bucket:    &bucket,
		Region:    &region,
		PathStyle: &pathStyle,
		Prefix:    &prefix,
	}

	if endpoint := blockString(s3, "endpoint"); endpoint != "" {
		cfg.Endpoint = &endpoint
	}

	if static, ok := firstBlock(s3, "static_credentials"); ok {
		accessKeyID := blockString(static, "access_key_id")

		cfg.Credentials = &client.S3Credentials{Mode: "Static", AccessKeyID: &accessKeyID}
		// Omitting the secret tells Warpgate to keep the one it already has.
		if secret := blockString(static, "secret_access_key"); secret != "" {
			cfg.Credentials.SecretAccessKey = &secret
		}
	} else if autos, ok := s3["auto_credentials"].([]any); ok && len(autos) > 0 {
		cfg.Credentials = &client.S3Credentials{Mode: "Auto"}
	}

	return cfg
}

// firstBlock returns the single element of a MaxItems:1 nested block.
func firstBlock(m map[string]any, key string) (map[string]any, bool) {
	items, ok := m[key].([]any)
	if !ok || len(items) == 0 || items[0] == nil {
		return nil, false
	}

	block, ok := items[0].(map[string]any)

	return block, ok
}

func blockString(m map[string]any, key string) string {
	v, _ := m[key].(string)

	return v
}

func blockBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)

	return v
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}

func boolValue(v *bool) bool {
	return v != nil && *v
}

func flattenPasswordPolicy(policy client.PasswordPolicy) []any {
	return []any{
		map[string]any{
			"min_length":        policy.MinLength,
			"require_uppercase": policy.RequireUppercase,
			"require_lowercase": policy.RequireLowercase,
			"require_digits":    policy.RequireDigits,
			"require_special":   policy.RequireSpecial,
		},
	}
}

func expandPasswordPolicy(d *schema.ResourceData) *client.PasswordPolicy {
	if !configuredValueExists(d, "password_policy") {
		return nil
	}

	rawPolicies := d.Get("password_policy").([]any)
	if len(rawPolicies) == 0 || rawPolicies[0] == nil {
		return nil
	}

	rawPolicy := rawPolicies[0].(map[string]any)

	return &client.PasswordPolicy{
		MinLength:        rawPolicy["min_length"].(int),
		RequireUppercase: rawPolicy["require_uppercase"].(bool),
		RequireLowercase: rawPolicy["require_lowercase"].(bool),
		RequireDigits:    rawPolicy["require_digits"].(bool),
		RequireSpecial:   rawPolicy["require_special"].(bool),
	}
}

// resourceParametersDelete handles the deletion of Warpgate parameters
func resourceParametersDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// For a global parameters resource, we don't actually delete it from the API.
	d.SetId("")

	return diags
}
