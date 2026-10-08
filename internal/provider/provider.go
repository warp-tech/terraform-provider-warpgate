// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// New returns a function that creates a new Terraform provider for Warpgate
// with the specified version information. The returned provider is configured
// with resources and data sources for managing Warpgate entities.
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("WARPGATE_HOST", nil),
					Description: "The Warpgate API host URL (e.g., https://warpgate.example.com)",
				},
				"insecure_skip_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("WARPGATE_INSECURE_SKIP_VERIFY", nil),
					Description: "Whether to skip the TLS certificate verification (self-signed certificates)",
				},
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("WARPGATE_TOKEN", nil),
					Description: "API token for authenticating with Warpgate API",
				},
				"headers": {
					Type:        schema.TypeMap,
					Optional:    true,
					Sensitive:   true,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Description: "Additional HTTP headers sent with every API request, for example the service token headers of an access proxy in front of Warpgate. They cannot override the `X-Warpgate-Token`, `Content-Type` and `Accept` headers the provider sets. `Host`, `Content-Length`, `Transfer-Encoding` and `Trailer` are rejected, because the HTTP client never sends them from a header map.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"warpgate_role":                  resourceRole(),
				"warpgate_admin_role":            resourceAdminRole(),
				"warpgate_user":                  resourceUser(),
				"warpgate_target":                resourceTarget(),
				"warpgate_user_role":             resourceUserRole(),
				"warpgate_user_admin_role":       resourceUserAdminRole(),
				"warpgate_target_role":           resourceTargetRole(),
				"warpgate_target_group":          resourceTargetGroup(),
				"warpgate_password_credential":   resourcePasswordCredential(),
				"warpgate_public_key_credential": resourcePublicKeyCredential(),
				"warpgate_user_sso_credential":   resourceUserSsoCredential(),
				"warpgate_ticket":                resourceTicket(),
				"warpgate_parameters":            resourceParameters(),
				"warpgate_ssh_key":               resourceSSHKey(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"warpgate_role":         dataSourceRole(),
				"warpgate_admin_role":   dataSourceAdminRole(),
				"warpgate_user":         dataSourceUser(),
				"warpgate_target":       dataSourceTarget(),
				"warpgate_ssh_own_keys": dataSourceSSHOwnKeys(),
				"warpgate_ssh_key":      dataSourceSSHKey(),
			},
		}

		p.ConfigureContextFunc = configure()
		p.TerraformVersion = "0.13+"

		return p
	}
}

type providerMeta struct {
	client *client.Client
}

// configure creates a configuration function for the Warpgate provider.
// It establishes a client connection to the Warpgate API using the provided
// host and token, and returns a metadata object containing the client and version.
func configure() func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		var diags diag.Diagnostics

		host := d.Get("host").(string)
		token := d.Get("token").(string)
		insecureSkipVerify := d.Get("insecure_skip_verify").(bool)

		headers, err := providerHeaders(d.Get("headers").(map[string]any))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// Ensure the host has the API path
		apiPath := "/@warpgate/admin/api"
		if !strings.Contains(host, apiPath) {
			if strings.HasSuffix(host, "/") {
				host = host + strings.TrimPrefix(apiPath, "/")
			} else {
				host = host + apiPath
			}
		}

		cfg := &client.Config{
			Host:               host,
			Token:              token,
			InsecureSkipVerify: insecureSkipVerify,
			Headers:            headers,
		}

		c, err := client.NewClient(cfg)
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("error creating client: %w", err))
		}

		meta := &providerMeta{
			client: c,
		}

		return meta, diags
	}
}

// unsendableHeaders are set by Go's HTTP client from the request itself and
// silently dropped from a header map, so configuring them would have no effect.
var unsendableHeaders = map[string]bool{
	"Host":              true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// providerHeaders converts the provider's `headers` map and rejects the
// headers the HTTP client would never send.
func providerHeaders(raw map[string]any) (map[string]string, error) {
	headers := make(map[string]string, len(raw))
	for name, value := range raw {
		if unsendableHeaders[http.CanonicalHeaderKey(name)] {
			return nil, fmt.Errorf("headers: %q cannot be set; the HTTP client derives it from the request", name)
		}
		headers[name] = value.(string)
	}
	return headers, nil
}

// parseCompositeID parses a composite ID in the format "id1:id2" and returns
// the individual components. Used for importing resources that have composite identifiers.
func parseCompositeID(id string, part1Name, part2Name string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected ID in format '%s:%s', got: %s", part1Name, part2Name, id)
	}
	return parts[0], parts[1], nil
}
