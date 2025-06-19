// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

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
			// SSH Target Configuration
			"ssh_options": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"http_options", "mysql_options", "postgres_options"},
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
						"password_auth": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"ssh_options.0.public_key_auth"},
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
							ConflictsWith: []string{"ssh_options.0.password_auth"},
							Description:   "Public key authentication for SSH",
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
				ConflictsWith: []string{"ssh_options", "mysql_options", "postgres_options"},
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
				ConflictsWith: []string{"ssh_options", "http_options", "postgres_options"},
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
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The MySQL password",
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
				ConflictsWith: []string{"ssh_options", "http_options", "mysql_options"},
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
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The PostgreSQL password",
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
		},
		CustomizeDiff: validateTargetConfig,
	}
}

// validateTargetConfig validates the target configuration in a Terraform resource diff,
// ensuring that exactly one type of target option is specified.
func validateTargetConfig(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	optionBlocks := []string{"ssh_options", "http_options", "mysql_options", "postgres_options"}

	count := 0
	for _, block := range optionBlocks {
		if v, ok := d.GetOk(block); ok && len(v.([]any)) > 0 {
			count++
		}
	}

	if count == 0 {
		return fmt.Errorf("one of ssh_options, http_options, mysql_options, or postgres_options must be specified")
	}

	if count > 1 {
		return fmt.Errorf("only one of ssh_options, http_options, mysql_options, postgres_option can be specified")
	}

	return nil
}

// resourceTargetCreate handles the creation of a new target in Warpgate based on
// the provided resource data.
func resourceTargetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	// Determine which type of target options is being used and build the appropriate request
	targetOptions, err := buildTargetOptions(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to build target options: %w", err))
	}

	req := &client.TargetDataRequest{
		Name:        name,
		Description: description,
		Options:     targetOptions,
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

	if err := d.Set("allow_roles", target.AllowRoles); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set allow_roles: %w", err))
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
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	// Determine which type of target options is being used and build the appropriate request
	targetOptions, err := buildTargetOptions(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to build target options: %w", err))
	}

	req := &client.TargetDataRequest{
		Name:        name,
		Description: description,
		Options:     targetOptions,
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

	return nil, fmt.Errorf("no target options specified")
}

// buildSshTargetOptions creates SSH target options from the resource data map.
func buildSSHTargetOptions(opts map[string]any) (*client.TargetSSHOptions, error) {
	host := opts["host"].(string)
	port := opts["port"].(int)
	username := opts["username"].(string)
	allowInsecureAlgos := opts["allow_insecure_algos"].(bool)

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
		auth = &client.SSHTargetPublicKeyAuth{
			Kind: "PublicKey",
		}
	} else {
		return nil, fmt.Errorf("SSH target requires either password_auth or public_key_auth")
	}

	return &client.TargetSSHOptions{
		Kind:               "Ssh",
		Host:               host,
		Port:               port,
		Username:           username,
		AllowInsecureAlgos: allowInsecureAlgos,
		Auth:               auth,
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

	// Extract headers
	var headers map[string]string
	if v, ok := opts["headers"]; ok {
		headersMap := v.(map[string]any)
		headers = make(map[string]string)
		for k, v := range headersMap {
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

	var password string
	if v, ok := opts["password"]; ok {
		password = v.(string)
	}

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
		Password: password,
		TLS:      tls,
	}, nil
}

// buildPostgresTargetOptions creates PostgreSQL target options from the resource data map.
func buildPostgresTargetOptions(opts map[string]any) (*client.TargetPostgresOptions, error) {
	host := opts["host"].(string)
	port := opts["port"].(int)
	username := opts["username"].(string)

	var password string
	if v, ok := opts["password"]; ok {
		password = v.(string)
	}

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
		Kind:     "Postgres",
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		TLS:      tls,
	}, nil
}

// setTargetOptions populates the appropriate Terraform schema block based on the target type
// from the Warpgate API.
func setTargetOptions(d *schema.ResourceData, options any) error {
	// Reset all options blocks
	if err := d.Set("ssh_options", []any{}); err != nil {
		return fmt.Errorf("failed to reset ssh_options: %w", err)
	}

	if err := d.Set("http_options", []any{}); err != nil {
		return fmt.Errorf("failed to reset http_options: %w", err)
	}

	if err := d.Set("mysql_options", []any{}); err != nil {
		return fmt.Errorf("failed to reset mysql_options: %w", err)
	}

	if err := d.Set("postgres_options", []any{}); err != nil {
		return fmt.Errorf("failed to reset postgres_options: %w", err)
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
			sshOpts["public_key_auth"] = []any{
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

		if password, ok := optionsMap["password"].(string); ok && password != "" {
			mysqlOpts["password"] = password
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

		if password, ok := optionsMap["password"].(string); ok && password != "" {
			pgOpts["password"] = password
		}

		return d.Set("postgres_options", []any{pgOpts})

	default:
		return fmt.Errorf("unknown target kind: %s", kind)
	}
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
