package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

var (
	validateSecretBackendName = validation.All(
		validation.StringLenBetween(1, 64),
		validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9._-]+$`), "may only contain letters, digits, '.', '_' and '-'"),
	)
	validateSecretBackendPathSegment = validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+$`), "may only contain letters, digits, '_' and '-'"),
	)
)

var secretBackendAuthBlocks = []string{"token_auth", "approle_auth", "kubernetes_auth"}

func resourceSecretBackend() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Vault or OpenBao secret backend that Warpgate resolves secret references and signs SSH certificates against (Warpgate 0.30+).",
		CreateContext: resourceSecretBackendCreate,
		ReadContext:   resourceSecretBackendRead,
		UpdateContext: resourceSecretBackendUpdate,
		DeleteContext: resourceSecretBackendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the backend, used as the backend segment of secret references and by SSH certificate authentication",
				ValidateFunc: validateSecretBackendName,
			},
			"backend_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Kind of server: `vault` or `openbao`",
				ValidateFunc: validation.StringInSlice([]string{"vault", "openbao"}, false),
			},
			"address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Address of the server, e.g. `https://vault.example.com:8200`",
				ValidateFunc: validation.All(validation.StringLenBetween(1, 2048), validation.IsURLWithHTTPorHTTPS),
			},
			"namespace": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Vault Enterprise or OpenBao namespace",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"tls_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skip verification of the server's TLS certificate",
			},
			"allowed_paths": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Path prefixes the backend may be used for, matched on whole segments: `mount/path` for KV secrets, `mount/sign/role` for SSH certificate signing. Empty allows any path.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 1024),
				},
				MaxItems: 256,
			},
			"token_auth": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: secretBackendAuthBlocks,
				Description:  "Authenticate with a static token. A renewable token is renewed by Warpgate.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The token. Warpgate never returns it, so changes made outside Terraform are not detected.",
						},
					},
				},
			},
			"approle_auth": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: secretBackendAuthBlocks,
				Description:  "Authenticate with AppRole",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_id": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The AppRole role ID",
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"secret_id": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "The AppRole secret ID. Warpgate never returns it, so changes made outside Terraform are not detected.",
						},
						"mount": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Mount path of the AppRole auth method. Warpgate uses `approle` when unset.",
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"kubernetes_auth": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: secretBackendAuthBlocks,
				Description:  "Authenticate with the service account token of the pod Warpgate runs in",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The Kubernetes auth role",
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"mount": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Mount path of the Kubernetes auth method. Warpgate uses `kubernetes` when unset.",
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
		},
	}
}

func secretBackendAuthFromResourceData(d *schema.ResourceData) (client.SecretBackendAuth, error) {
	blocks := map[string]any{}
	for _, key := range secretBackendAuthBlocks {
		blocks[key] = d.Get(key)
	}

	if block, ok := firstBlock(blocks, "token_auth"); ok {
		token, _ := block["token"].(string)
		return client.SecretBackendAuth{
			Method: client.SecretBackendAuthToken,
			Token:  token,
		}, nil
	}

	if block, ok := firstBlock(blocks, "approle_auth"); ok {
		roleID, _ := block["role_id"].(string)
		secretID, _ := block["secret_id"].(string)
		mount, _ := block["mount"].(string)
		return client.SecretBackendAuth{
			Method:   client.SecretBackendAuthAppRole,
			RoleID:   roleID,
			SecretID: secretID,
			Mount:    mount,
		}, nil
	}

	if block, ok := firstBlock(blocks, "kubernetes_auth"); ok {
		role, _ := block["role"].(string)
		mount, _ := block["mount"].(string)
		return client.SecretBackendAuth{
			Method: client.SecretBackendAuthKubernetes,
			Role:   role,
			Mount:  mount,
		}, nil
	}

	return client.SecretBackendAuth{}, fmt.Errorf("one of token_auth, approle_auth or kubernetes_auth is required")
}

func secretBackendRequestFromResourceData(d *schema.ResourceData) (*client.SecretBackendRequest, error) {
	auth, err := secretBackendAuthFromResourceData(d)
	if err != nil {
		return nil, err
	}

	allowedPaths := []string{}
	for _, path := range d.Get("allowed_paths").([]any) {
		if s, ok := path.(string); ok {
			allowedPaths = append(allowedPaths, s)
		}
	}

	return &client.SecretBackendRequest{
		Name:          d.Get("name").(string),
		BackendType:   d.Get("backend_type").(string),
		Address:       d.Get("address").(string),
		Namespace:     d.Get("namespace").(string),
		Auth:          auth,
		TLSSkipVerify: d.Get("tls_skip_verify").(bool),
		AllowedPaths:  allowedPaths,
	}, nil
}

func setSecretBackendData(d *schema.ResourceData, backend *client.SecretBackend) error {
	values := map[string]any{
		"name":            backend.Name,
		"backend_type":    backend.BackendType,
		"address":         backend.Address,
		"namespace":       backend.Namespace,
		"tls_skip_verify": backend.TLSSkipVerify,
		"allowed_paths":   backend.AllowedPaths,
	}

	auth := map[string][]any{}
	for _, block := range secretBackendAuthBlocks {
		auth[block] = []any{}
	}

	switch backend.Auth.Method {
	case client.SecretBackendAuthToken:
		token, _ := d.Get("token_auth.0.token").(string)
		auth["token_auth"] = []any{map[string]any{
			"token": token,
		}}
	case client.SecretBackendAuthAppRole:
		secretID, _ := d.Get("approle_auth.0.secret_id").(string)
		auth["approle_auth"] = []any{map[string]any{
			"role_id":   backend.Auth.RoleID,
			"secret_id": secretID,
			"mount":     backend.Auth.Mount,
		}}
	case client.SecretBackendAuthKubernetes:
		auth["kubernetes_auth"] = []any{map[string]any{
			"role":  backend.Auth.Role,
			"mount": backend.Auth.Mount,
		}}
	default:
		return fmt.Errorf("unknown secret backend auth method: %s", backend.Auth.Method)
	}

	for key, value := range values {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}
	for key, value := range auth {
		if err := d.Set(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

func resourceSecretBackendCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*providerMeta).client

	req, err := secretBackendRequestFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	backend, err := c.CreateSecretBackend(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create secret backend: %w", err))
	}

	d.SetId(backend.ID)

	return resourceSecretBackendRead(ctx, d, meta)
}

func resourceSecretBackendRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*providerMeta).client

	backend, err := c.GetSecretBackend(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read secret backend: %w", err))
	}

	if backend == nil {
		d.SetId("")
		return nil
	}

	if err := setSecretBackendData(d, backend); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSecretBackendUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*providerMeta).client

	req, err := secretBackendRequestFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := c.UpdateSecretBackend(ctx, d.Id(), req); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update secret backend: %w", err))
	}

	return resourceSecretBackendRead(ctx, d, meta)
}

func resourceSecretBackendDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*providerMeta).client

	if err := c.DeleteSecretBackend(ctx, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete secret backend: %w", err))
	}

	d.SetId("")

	return nil
}
