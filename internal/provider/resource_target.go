// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// targetOptionBlocks are the per-protocol configuration blocks; a target carries
// exactly one of them.
var targetOptionBlocks = []string{"ssh_options", "http_options", "mysql_options", "postgres_options", "kubernetes_options", "rdp_options", "vnc_options"}

// otherTargetOptionBlocks lists every protocol block except the given one.
func otherTargetOptionBlocks(except string) []string {
	var others []string
	for _, block := range targetOptionBlocks {
		if block != except {
			others = append(others, block)
		}
	}
	return others
}

// resourceTarget creates and returns a schema for the target resource.
func resourceTarget() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTargetCreate,
		ReadContext:   resourceTargetRead,
		UpdateContext: resourceTargetUpdate,
		DeleteContext: resourceTargetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the target",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the target",
			},
			"allow_roles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of roles allowed to access this target",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Which target group this target is assigned to",
			},
			rateLimitBytesPerSecondKey: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Bandwidth limit in bytes per second",
				ValidateFunc: validation.IntAtLeast(0),
			},
			"ticket_max_duration_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum ticket duration in seconds for this target",
			},
			"ticket_requests_disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether ticket requests are disabled for this target",
			},
			"ticket_require_approval": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether ticket requests require manual approval",
			},
			"require_approval": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Hold new sessions to this target until an administrator approves them",
			},
			"ticket_max_uses": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of uses allowed per ticket",
			},
			// SSH Target Configuration
			"ssh_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("ssh_options"),
				Description:   "SSH target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The SSH server hostname or IP address",
						},
						"port": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "The SSH server port",
							ValidateFunc: validation.IsPortNumber,
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The SSH username",
						},
						"allow_insecure_algos": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Allow insecure SSH algorithms",
						},
						"jump_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of another target to use as an SSH jump host",
						},
						"password_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"ssh_options.0.public_key_auth", "ssh_options.0.iam_role_auth"},
							Description:   "Password authentication for SSH",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "The password for SSH authentication",
									},
								},
							},
						},
						"public_key_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"ssh_options.0.password_auth", "ssh_options.0.iam_role_auth"},
							Description:   "Public key authentication for SSH",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Specific stored client key ID to authenticate with. If omitted, default keys are used.",
									},
								},
							},
						},
						"iam_role_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"ssh_options.0.password_auth", "ssh_options.0.public_key_auth"},
							Description:   "IAM Role authentication for SSH",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
					},
				},
			},
			// HTTP Target Configuration
			"http_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("http_options"),
				Description:   "HTTP target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The HTTP server URL",
							ValidateFunc: validation.IsURLWithHTTPorHTTPS,
						},
						"tls": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"Disabled", "Preferred", "Required"}, false),
										Description:  "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Verify TLS certificates",
									},
								},
							},
						},
						"headers": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "HTTP headers to include in requests",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"external_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "External host for HTTP requests",
						},
					},
				},
			},
			// MySQL Target Configuration
			"mysql_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("mysql_options"),
				Description:   "MySQL target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The MySQL server hostname or IP address",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The MySQL server port",
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The MySQL username",
						},
						"password": {
							Type:          schema.TypeString,
							Optional:      true,
							Sensitive:     true,
							ConflictsWith: []string{"mysql_options.0.iam_role_auth"},
							Description:   "The MySQL password",
						},
						"iam_role_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"mysql_options.0.password"},
							Description:   "AWS IAM authentication instead of a password",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
						"tls": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"Disabled", "Preferred", "Required"}, false),
										Description:  "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Verify TLS certificates",
									},
								},
							},
						},
					},
				},
			},
			// PostgreSQL Target Configuration
			"postgres_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("postgres_options"),
				Description:   "PostgreSQL target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The PostgreSQL server hostname or IP address",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The PostgreSQL server port",
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The PostgreSQL username",
						},
						"default_database_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The default PostgreSQL database name to connect to",
						},
						"protocol_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "3.2",
							ValidateFunc: validation.StringInSlice([]string{"3.0", "3.2"}, false),
							Description:  "The PostgreSQL protocol version to request. Valid values: 3.0, 3.2",
						},
						"idle_timeout": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Idle connection timeout as a duration string, e.g. 10m",
						},
						"password": {
							Type:          schema.TypeString,
							Optional:      true,
							Sensitive:     true,
							ConflictsWith: []string{"postgres_options.0.iam_role_auth"},
							Description:   "The PostgreSQL password",
						},
						"iam_role_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"postgres_options.0.password"},
							Description:   "AWS IAM authentication instead of a password",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
						"tls": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"Disabled", "Preferred", "Required"}, false),
										Description:  "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Verify TLS certificates",
									},
								},
							},
						},
					},
				},
			},
			// Kubernetes Target Configuration
			"kubernetes_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("kubernetes_options"),
				Description:   "Kubernetes target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_url": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The Kubernetes cluster URL",
							ValidateFunc: validation.IsURLWithHTTPorHTTPS,
						},
						"tls": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"Disabled", "Preferred", "Required"}, false),
										Description:  "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Required:    true,
										Description: "Verify TLS certificates",
									},
								},
							},
						},
						"token_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"kubernetes_options.0.certificate_auth", "kubernetes_options.0.iam_role_auth"},
							Description:   "Token authentication for Kubernetes",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "The bearer token for Kubernetes authentication",
									},
								},
							},
						},
						"certificate_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"kubernetes_options.0.token_auth", "kubernetes_options.0.iam_role_auth"},
							Description:   "Certificate authentication for Kubernetes",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The client certificate PEM",
									},
									"private_key": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "The client private key PEM",
									},
								},
							},
						},
						"iam_role_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"kubernetes_options.0.token_auth", "kubernetes_options.0.certificate_auth"},
							Description:   "AWS IAM authentication (EKS) instead of a token or certificate",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
					},
				},
			},
			// RDP Target Configuration
			"rdp_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("rdp_options"),
				Description:   "RDP target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The RDP server hostname or IP address",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3389,
							Description:  "The RDP server port",
							ValidateFunc: validation.IsPortNumber,
						},
						"username": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The RDP username",
						},
						"domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The RDP authentication domain (Windows domain)",
						},
						"password": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The password for RDP authentication",
						},
						"verify_tls": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Verify the RDP server's TLS certificate",
						},
						"tls_security": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Tls12",
							ValidateFunc: validation.StringInSlice([]string{"Tls12", "Tls12WithLegacyCiphers", "Tls10Unsafe"}, false),
							Description:  "TLS security profile for the RDP connection: Tls12, Tls12WithLegacyCiphers, Tls10Unsafe",
						},
						"compression": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "remotefx",
							ValidateFunc: validation.StringInSlice([]string{"remotefx", "lossless"}, false),
							Description:  "Codec advertised to the RDP server: remotefx, or lossless when Warpgate and the target share a network",
						},
						"interactive_logon": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Show the target's own sign-in screen instead of logging on automatically",
						},
					},
				},
			},
			// VNC Target Configuration
			"vnc_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: otherTargetOptionBlocks("vnc_options"),
				Description:   "VNC target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The VNC server hostname or IP address",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5900,
							Description:  "The VNC server port",
							ValidateFunc: validation.IsPortNumber,
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The VNC password; omit for a server without authentication",
						},
					},
				},
			},
		},
		CustomizeDiff: validateTargetConfig,
	}
}

// validateTargetConfig validates the target configuration in a Terraform resource diff,
// ensuring that exactly one type of target option is specified.
func validateTargetConfig(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	count := 0
	for _, block := range targetOptionBlocks {
		if v, ok := d.GetOk(block); ok && len(v.([]any)) > 0 {
			count++
		}
	}

	if count == 0 {
		return fmt.Errorf("one of %s must be specified", strings.Join(targetOptionBlocks, ", "))
	}

	if count > 1 {
		return fmt.Errorf("only one of %s can be specified", strings.Join(targetOptionBlocks, ", "))
	}

	return nil
}

// resourceTargetCreate handles the creation of a new target in Warpgate based on
// the provided resource data.
func resourceTargetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	req, err := buildTargetDataRequest(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to build target request: %w", err))
	}

	target, err := c.CreateTarget(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create target: %w", err))
	}

	d.SetId(target.ID)

	return resourceTargetRead(ctx, d, meta)
}

// resourceTargetRead retrieves the target data from Warpgate and updates the
// Terraform state accordingly.
func resourceTargetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	target, err := c.GetTarget(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read target: %w", err))
	}

	// If the target was not found, return nil to indicate that the resource no longer exists
	if target == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", target.Name); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set name: %w", err))
	}

	if err := d.Set("description", target.Description); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set description: %w", err))
	}

	if err := d.Set("group_id", target.GroupId); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set group_id: %w", err))
	}

	if err := setOptionalInt(d, rateLimitBytesPerSecondKey, target.RateLimitBytesPerSecond); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set rate_limit_bytes_per_second: %w", err))
	}

	if err := d.Set("allow_roles", target.AllowRoles); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set allow_roles: %w", err))
	}

	if err := setOptionalInt64(d, "ticket_max_duration_seconds", target.TicketMaxDurationSeconds); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_max_duration_seconds: %w", err))
	}

	if err := d.Set("ticket_requests_disabled", target.TicketRequestsDisabled); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_requests_disabled: %w", err))
	}

	if err := d.Set("ticket_require_approval", target.TicketRequireApproval); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_require_approval: %w", err))
	}

	if err := d.Set("require_approval", target.RequireApproval); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set require_approval: %w", err))
	}

	if err := setOptionalInt(d, "ticket_max_uses", target.TicketMaxUses); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set ticket_max_uses: %w", err))
	}

	// Set the appropriate options block based on target type
	if err := setTargetOptions(d, target.Options); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set target options: %w", err))
	}

	return diags
}

// resourceTargetUpdate handles the update of an existing target in Warpgate based on
// the provided resource data changes.
func resourceTargetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	id := d.Id()

	req, err := buildTargetDataRequest(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to build target request: %w", err))
	}

	_, err = c.UpdateTarget(ctx, id, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update target: %w", err))
	}

	return resourceTargetRead(ctx, d, meta)
}

// resourceTargetDelete removes a target from Warpgate based on the resource data.
func resourceTargetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Id()

	err := c.DeleteTarget(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete target: %w", err))
	}

	d.SetId("")

	return diags
}

func buildTargetDataRequest(d *schema.ResourceData) (*client.TargetDataRequest, error) {
	targetOptions, err := buildTargetOptions(d)
	if err != nil {
		return nil, err
	}

	return &client.TargetDataRequest{
		Name:                     d.Get("name").(string),
		Description:              d.Get("description").(string),
		GroupId:                  d.Get("group_id").(string),
		RateLimitBytesPerSecond:  optionalIntPointer(d, rateLimitBytesPerSecondKey),
		Options:                  targetOptions,
		TicketMaxDurationSeconds: optionalInt64Pointer(d, "ticket_max_duration_seconds"),
		TicketRequestsDisabled:   d.Get("ticket_requests_disabled").(bool),
		TicketRequireApproval:    d.Get("ticket_require_approval").(bool),
		TicketMaxUses:            optionalIntPointer(d, "ticket_max_uses"),
		RequireApproval:          d.Get("require_approval").(bool),
	}, nil
}

// buildTargetOptions constructs the appropriate target options based on which configuration
// block is specified in the resource data.
func buildTargetOptions(d *schema.ResourceData) (client.TargetOptions, error) {
	// Check for SSH options
	if v, ok := d.GetOk("ssh_options"); ok && len(v.([]any)) > 0 {
		sshOpts := v.([]any)[0].(map[string]any)
		return buildSSHTargetOptions(sshOpts)
	}

	// Check for HTTP options
	if v, ok := d.GetOk("http_options"); ok && len(v.([]any)) > 0 {
		httpOpts := v.([]any)[0].(map[string]any)
		return buildHTTPTargetOptions(httpOpts)
	}

	// Check for MySQL options
	if v, ok := d.GetOk("mysql_options"); ok && len(v.([]any)) > 0 {
		mysqlOpts := v.([]any)[0].(map[string]any)
		return buildMysqlTargetOptions(mysqlOpts)
	}

	// Check for PostgreSQL options
	if v, ok := d.GetOk("postgres_options"); ok && len(v.([]any)) > 0 {
		pgOpts := v.([]any)[0].(map[string]any)
		return buildPostgresTargetOptions(pgOpts)
	}

	// Check for Kubernetes options
	if v, ok := d.GetOk("kubernetes_options"); ok && len(v.([]any)) > 0 {
		k8sOpts := v.([]any)[0].(map[string]any)
		return buildKubernetesTargetOptions(k8sOpts)
	}

	// Check for RDP options
	if v, ok := d.GetOk("rdp_options"); ok && len(v.([]any)) > 0 {
		rdpOpts := v.([]any)[0].(map[string]any)
		return buildRDPTargetOptions(rdpOpts)
	}

	// Check for VNC options
	if v, ok := d.GetOk("vnc_options"); ok && len(v.([]any)) > 0 {
		vncOpts := v.([]any)[0].(map[string]any)
		return buildVNCTargetOptions(vncOpts), nil
	}

	return nil, fmt.Errorf("no target options specified")
}

// buildSSHTargetOptions creates SSH target options from the resource data map.
func buildSSHTargetOptions(opts map[string]any) (*client.TargetSSHOptions, error) {
	host := opts["host"].(string)
	port := opts["port"].(int)
	username := opts["username"].(string)
	allowInsecureAlgos := opts["allow_insecure_algos"].(bool)

	var jumpHost string
	if jh, ok := opts["jump_host"]; ok && jh != nil {
		jumpHost = jh.(string)
	}

	// Determine which auth method is being used
	var auth client.SSHTargetAuth

	if v, ok := opts["password_auth"]; ok && len(v.([]any)) > 0 {
		pwAuth := v.([]any)[0].(map[string]any)
		password := pwAuth["password"].(string)
		auth = &client.SSHTargetPasswordAuth{
			Kind:     "Password",
			Password: password,
		}
	} else if v, ok := opts["public_key_auth"]; ok && len(v.([]any)) > 0 {
		var keyID string
		if pkAuth, ok := v.([]any)[0].(map[string]any); ok {
			if k, ok := pkAuth["key_id"]; ok && k != nil {
				keyID = k.(string)
			}
		}
		auth = &client.SSHTargetPublicKeyAuth{
			Kind:  "PublicKey",
			KeyID: keyID,
		}
	} else if v, ok := opts["iam_role_auth"]; ok && len(v.([]any)) > 0 {
		auth = &client.SSHTargetIamRoleAuth{
			Kind: "IamRole",
		}
	} else {
		return nil, fmt.Errorf("SSH target requires password_auth, public_key_auth, or iam_role_auth")
	}

	return &client.TargetSSHOptions{
		Kind:               "Ssh",
		Host:               host,
		Port:               port,
		Username:           username,
		AllowInsecureAlgos: allowInsecureAlgos,
		Auth:               auth,
		JumpHost:           jumpHost,
	}, nil
}

// buildHttpTargetOptions creates HTTP target options from the resource data map.
func buildHTTPTargetOptions(opts map[string]any) (*client.TargetHTTPOptions, error) {
	url := opts["url"].(string)

	// Extract TLS settings
	var tls client.TLS
	if v, ok := opts["tls"]; ok {
		var err error
		tls, err = parseTLSConfig(v.([]any))
		if err != nil {
			return nil, fmt.Errorf("invalid TLS configuration for HTTP target: %w", err)
		}
	}

	headers := map[string]string{}
	if v, ok := opts["headers"].(map[string]any); ok {
		for k, v := range v {
			headers[k] = v.(string)
		}
	}

	// Check for external_host
	var externalHost string
	if v, ok := opts["external_host"]; ok {
		externalHost = v.(string)
	}

	return &client.TargetHTTPOptions{
		Kind:         "Http",
		URL:          url,
		TLS:          tls,
		Headers:      headers,
		ExternalHost: externalHost,
	}, nil
}

// buildMysqlTargetOptions creates MySQL target options from the resource datamap.
func buildMysqlTargetOptions(opts map[string]any) (*client.TargetMySQLOptions, error) {
	host := opts["host"].(string)
	port := opts["port"].(int)
	username := opts["username"].(string)

	// Extract TLS settings
	var tls client.TLS
	if v, ok := opts["tls"]; ok {
		var err error
		tls, err = parseTLSConfig(v.([]any))
		if err != nil {
			return nil, fmt.Errorf("invalid TLS configuration for MySQL target: %w", err)
		}
	}

	return &client.TargetMySQLOptions{
		Kind:     "MySql",
		Host:     host,
		Port:     port,
		Username: username,
		Auth:     buildDatabaseTargetAuth(opts),
		TLS:      tls,
	}, nil
}

// buildDatabaseTargetAuth picks the MySQL/PostgreSQL auth method: an
// iam_role_auth block wins, otherwise the password (possibly empty) is sent.
func buildDatabaseTargetAuth(opts map[string]any) client.DatabaseTargetAuth {
	if v, ok := opts["iam_role_auth"].([]any); ok && len(v) > 0 {
		return &client.DatabaseTargetIamRoleAuth{Kind: "IamRole"}
	}

	password, _ := opts["password"].(string)

	return &client.DatabaseTargetPasswordAuth{Kind: "Password", Password: password}
}

// buildPostgresTargetOptions creates PostgreSQL target options from the resource data map.
func buildPostgresTargetOptions(opts map[string]any) (*client.TargetPostgresOptions, error) {
	host := opts["host"].(string)
	port := opts["port"].(int)
	username := opts["username"].(string)

	var defaultDatabaseName string
	if v, ok := opts["default_database_name"]; ok {
		defaultDatabaseName = v.(string)
	}

	var protocolVersion string
	if v, ok := opts["protocol_version"]; ok {
		protocolVersion = v.(string)
	}

	idleTimeout, _ := opts["idle_timeout"].(string)

	// Extract TLS settings
	var tls client.TLS
	if v, ok := opts["tls"]; ok {
		var err error
		tls, err = parseTLSConfig(v.([]any))
		if err != nil {
			return nil, fmt.Errorf("invalid TLS configuration for PostgreSQL target: %w", err)
		}
	}

	return &client.TargetPostgresOptions{
		Kind:                "Postgres",
		Host:                host,
		Port:                port,
		Username:            username,
		DefaultDatabaseName: defaultDatabaseName,
		IdleTimeout:         idleTimeout,
		ProtocolVersion:     protocolVersion,
		Auth:                buildDatabaseTargetAuth(opts),
		TLS:                 tls,
	}, nil
}

// buildKubernetesTargetOptions creates Kubernetes target options from the resource data map.
func buildKubernetesTargetOptions(opts map[string]any) (*client.TargetKubernetesOptions, error) {
	clusterURL := opts["cluster_url"].(string)

	// Extract TLS settings
	var tls client.TLS
	if v, ok := opts["tls"]; ok {
		var err error
		tls, err = parseTLSConfig(v.([]any))
		if err != nil {
			return nil, fmt.Errorf("invalid TLS configuration for Kubernetes target: %w", err)
		}
	}

	// Determine which auth method is being used
	var auth client.KubernetesTargetAuth

	if v, ok := opts["token_auth"]; ok && len(v.([]any)) > 0 {
		tokenAuth := v.([]any)[0].(map[string]any)
		token := tokenAuth["token"].(string)
		auth = &client.KubernetesTargetTokenAuth{
			Kind:  "Token",
			Token: token,
		}
	} else if v, ok := opts["certificate_auth"]; ok && len(v.([]any)) > 0 {
		certAuth := v.([]any)[0].(map[string]any)
		certificate := certAuth["certificate"].(string)
		privateKey := certAuth["private_key"].(string)
		auth = &client.KubernetesTargetCertificateAuth{
			Kind:        "Certificate",
			Certificate: certificate,
			PrivateKey:  privateKey,
		}
	} else if v, ok := opts["iam_role_auth"]; ok && len(v.([]any)) > 0 {
		auth = &client.KubernetesTargetIamRoleAuth{Kind: "IamRole"}
	} else {
		return nil, fmt.Errorf("kubernetes target requires token_auth, certificate_auth, or iam_role_auth")
	}

	return &client.TargetKubernetesOptions{
		Kind:       "Kubernetes",
		ClusterURL: clusterURL,
		TLS:        tls,
		Auth:       auth,
	}, nil
}

// buildRDPTargetOptions creates RDP target options from the resource data map.
func buildRDPTargetOptions(opts map[string]any) (*client.TargetRDPOptions, error) {
	host := opts["host"].(string)
	port := opts["port"].(int)
	username := opts["username"].(string)
	password := opts["password"].(string)

	var domain string
	if v, ok := opts["domain"]; ok {
		domain = v.(string)
	}

	verifyTLS := opts["verify_tls"].(bool)

	var tlsSecurity string
	if v, ok := opts["tls_security"]; ok {
		tlsSecurity = v.(string)
	}

	compression, _ := opts["compression"].(string)
	interactiveLogon, _ := opts["interactive_logon"].(bool)

	auth := &client.RDPTargetPasswordAuth{
		Kind:     "Password",
		Password: password,
	}

	return &client.TargetRDPOptions{
		Kind:             "Rdp",
		Host:             host,
		Port:             port,
		Username:         username,
		Domain:           domain,
		Auth:             auth,
		VerifyTLS:        verifyTLS,
		Compression:      compression,
		InteractiveLogon: interactiveLogon,
		TLSSecurity:      tlsSecurity,
	}, nil
}

// buildVNCTargetOptions creates VNC target options from the resource data map.
// An empty password selects the server's no-authentication mode.
func buildVNCTargetOptions(opts map[string]any) *client.TargetVncOptions {
	var auth client.VncTargetAuth = &client.VncTargetNoneAuth{Kind: "None"}
	if password, _ := opts["password"].(string); password != "" {
		auth = &client.VncTargetPasswordAuth{Kind: "Password", Password: password}
	}

	return &client.TargetVncOptions{
		Kind: "Vnc",
		Host: opts["host"].(string),
		Port: opts["port"].(int),
		Auth: auth,
	}
}

// setTargetOptions populates the appropriate Terraform schema block based on the target type
// from the Warpgate API.
func setTargetOptions(d *schema.ResourceData, options any) error {
	for _, block := range targetOptionBlocks {
		if err := d.Set(block, []any{}); err != nil {
			return fmt.Errorf("failed to reset %s: %w", block, err)
		}
	}

	// Type assertion based on the "kind" field in the options map
	optionsMap, err := targetOptionsToMap(options)
	if err != nil {
		return fmt.Errorf("failed to convert target options to map: %w", err)
	}

	kind, ok := optionsMap["kind"].(string)
	if !ok {
		return fmt.Errorf("missing 'kind' field in target options")
	}

	switch kind {
	case "Ssh":
		sshOpts := map[string]any{
			"host":                 optionsMap["host"],
			"port":                 optionsMap["port"],
			"username":             optionsMap["username"],
			"allow_insecure_algos": optionsMap["allow_insecure_algos"],
		}

		if jumpHost, ok := optionsMap["jump_host"].(string); ok && jumpHost != "" {
			sshOpts["jump_host"] = jumpHost
		}

		// Handle auth block
		auth, ok := optionsMap["auth"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid auth field in SSH options")
		}

		authKind, ok := auth["kind"].(string)
		if !ok {
			return fmt.Errorf("missing 'kind' field in auth options")
		}

		switch authKind {
		case "Password":
			sshOpts["password_auth"] = []any{
				map[string]any{
					"password": auth["password"],
				},
			}
		case "PublicKey":
			pkAuthMap := map[string]any{}
			if keyID, ok := auth["key_id"].(string); ok && keyID != "" {
				pkAuthMap["key_id"] = keyID
			}
			sshOpts["public_key_auth"] = []any{pkAuthMap}
		case "IamRole":
			sshOpts["iam_role_auth"] = []any{
				map[string]any{},
			}
		default:
			return fmt.Errorf("unknown SSH auth kind: %s", authKind)
		}

		return d.Set("ssh_options", []any{sshOpts})

	case "Http":
		tls, ok := optionsMap["tls"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid tls field in HTTP options")
		}

		tlsOpts := map[string]any{
			"mode":   tls["mode"],
			"verify": tls["verify"],
		}

		httpOpts := map[string]any{
			"url": optionsMap["url"],
			"tls": []any{tlsOpts},
		}

		if headers, ok := optionsMap["headers"].(map[string]any); ok && len(headers) > 0 {
			httpOpts["headers"] = headers
		}

		if externalHost, ok := optionsMap["external_host"].(string); ok && externalHost != "" {
			httpOpts["external_host"] = externalHost
		}

		return d.Set("http_options", []any{httpOpts})

	case "MySql":
		tls, ok := optionsMap["tls"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid tls field in MySQL options")
		}

		tlsOpts := map[string]any{
			"mode":   tls["mode"],
			"verify": tls["verify"],
		}

		mysqlOpts := map[string]any{
			"host":     optionsMap["host"],
			"port":     optionsMap["port"],
			"username": optionsMap["username"],
			"tls":      []any{tlsOpts},
		}

		if err := setDatabaseTargetAuth(mysqlOpts, optionsMap); err != nil {
			return err
		}

		return d.Set("mysql_options", []any{mysqlOpts})

	case "Postgres":
		tls, ok := optionsMap["tls"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid tls field in PostgreSQL options")
		}

		tlsOpts := map[string]any{
			"mode":   tls["mode"],
			"verify": tls["verify"],
		}

		pgOpts := map[string]any{
			"host":     optionsMap["host"],
			"port":     optionsMap["port"],
			"username": optionsMap["username"],
			"tls":      []any{tlsOpts},
		}

		if defaultDatabaseName, ok := optionsMap["default_database_name"].(string); ok && defaultDatabaseName != "" {
			pgOpts["default_database_name"] = defaultDatabaseName
		}

		if protocolVersion, ok := optionsMap["protocol_version"].(string); ok && protocolVersion != "" {
			pgOpts["protocol_version"] = protocolVersion
		}

		if idleTimeout, ok := optionsMap["idle_timeout"].(string); ok && idleTimeout != "" {
			pgOpts["idle_timeout"] = idleTimeout
		}

		if err := setDatabaseTargetAuth(pgOpts, optionsMap); err != nil {
			return err
		}

		return d.Set("postgres_options", []any{pgOpts})

	case "Kubernetes":
		tls, ok := optionsMap["tls"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid tls field in Kubernetes options")
		}

		tlsOpts := map[string]any{
			"mode":   tls["mode"],
			"verify": tls["verify"],
		}

		k8sOpts := map[string]any{
			"cluster_url": optionsMap["cluster_url"],
			"tls":         []any{tlsOpts},
		}

		// Handle auth block
		auth, ok := optionsMap["auth"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid auth field in Kubernetes options")
		}

		authKind, ok := auth["kind"].(string)
		if !ok {
			return fmt.Errorf("missing 'kind' field in Kubernetes auth options")
		}

		switch authKind {
		case "Token":
			k8sOpts["token_auth"] = []any{
				map[string]any{
					"token": auth["token"],
				},
			}
		case "Certificate":
			k8sOpts["certificate_auth"] = []any{
				map[string]any{
					"certificate": auth["certificate"],
					"private_key": auth["private_key"],
				},
			}
		case "IamRole":
			k8sOpts["iam_role_auth"] = []any{map[string]any{}}
		default:
			return fmt.Errorf("unknown Kubernetes auth kind: %s", authKind)
		}

		return d.Set("kubernetes_options", []any{k8sOpts})

	case "Rdp":
		rdpOpts := map[string]any{
			"host":              optionsMap["host"],
			"port":              optionsMap["port"],
			"username":          optionsMap["username"],
			"verify_tls":        optionsMap["verify_tls"],
			"tls_security":      optionsMap["tls_security"],
			"compression":       optionsMap["compression"],
			"interactive_logon": optionsMap["interactive_logon"],
		}

		if domain, ok := optionsMap["domain"].(string); ok && domain != "" {
			rdpOpts["domain"] = domain
		}

		auth, ok := optionsMap["auth"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid auth field in RDP options")
		}

		if password, ok := auth["password"].(string); ok && password != "" {
			rdpOpts["password"] = password
		}

		return d.Set("rdp_options", []any{rdpOpts})

	case "Vnc":
		vncOpts := map[string]any{
			"host": optionsMap["host"],
			"port": optionsMap["port"],
		}

		auth, ok := optionsMap["auth"].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid auth field in VNC options")
		}

		if password, ok := auth["password"].(string); ok && password != "" {
			vncOpts["password"] = password
		}

		return d.Set("vnc_options", []any{vncOpts})

	default:
		return fmt.Errorf("unknown target kind: %s", kind)
	}
}

// setDatabaseTargetAuth is the read-side mirror of buildDatabaseTargetAuth. An
// empty password is left unset so a target configured without one stays clean.
func setDatabaseTargetAuth(block map[string]any, optionsMap map[string]any) error {
	auth, ok := optionsMap["auth"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid auth field in database options")
	}

	switch kind, _ := auth["kind"].(string); kind {
	case "Password":
		if password, ok := auth["password"].(string); ok && password != "" {
			block["password"] = password
		}
	case "IamRole":
		block["iam_role_auth"] = []any{map[string]any{}}
	default:
		return fmt.Errorf("unknown database auth kind: %s", kind)
	}

	return nil
}

// targetOptionsToMap converts target options from the Warpgate API to a map
// that can be used to populate the Terraform schema.
func targetOptionsToMap(options any) (map[string]any, error) {
	// Marshal the options to JSON for easy conversion to map
	jsonData, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal target options: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal target options to map: %w", err)
	}

	return result, nil
}

// parseTLSConfig extracts TLS configuration from the Terraform schema representation.
func parseTLSConfig(tlsData []any) (client.TLS, error) {
	if len(tlsData) == 0 {
		return client.TLS{}, fmt.Errorf("tls configuration not provided")
	}

	tlsMap := tlsData[0].(map[string]any)
	mode := client.TLSMode(tlsMap["mode"].(string))
	verify := tlsMap["verify"].(bool)

	return client.TLS{
		Mode:   mode,
		Verify: verify,
	}, nil
}
