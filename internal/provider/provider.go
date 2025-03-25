// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Thunderbottom/terraform-provider-warpgate/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("WARPGATE_TOKEN", nil),
					Description: "API token for authenticating with Warpgate API",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"warpgate_role":        resourceRole(),
				"warpgate_user":        resourceUser(),
				"warpgate_target":      resourceTarget(),
				"warpgate_user_role":   resourceUserRole(),
				"warpgate_target_role": resourceTargetRole(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"warpgate_role":   dataSourceRole(),
				"warpgate_user":   dataSourceUser(),
				"warpgate_target": dataSourceTarget(),
			},
		}

		p.ConfigureContextFunc = configure()
		p.TerraformVersion = "0.13+"

		return p
	}
}

type providerMeta struct {
	client  *client.Client
	version string
}

// configure creates a configuration function for the Warpgate provider.
// It establishes a client connection to the Warpgate API using the provided
// host and token, and returns a metadata object containing the client and version.
func configure() func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		var diags diag.Diagnostics

		host := d.Get("host").(string)
		token := d.Get("token").(string)

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
			Host:  host,
			Token: token,
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
