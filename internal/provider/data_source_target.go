// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceTarget creates and returns a schema for the target data source.
func dataSourceTarget() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTargetRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the target",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the target",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "SSH target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SSH server hostname or IP address",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The SSH server port",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The SSH username",
						},
						"allow_insecure_algos": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Allow insecure SSH algorithms",
						},
						"password_auth": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Password authentication for SSH",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The password for SSH authentication",
									},
								},
							},
						},
						"public_key_auth": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Public key authentication for SSH",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
					},
				},
			},
			// HTTP Target Configuration
			"http_options": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "HTTP target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The HTTP server URL",
						},
						"tls": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Verify TLS certificates",
									},
								},
							},
						},
						"headers": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "HTTP headers to include in requests",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"external_host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "External host for HTTP requests",
						},
					},
				},
			},
			// MySQL Target Configuration
			"mysql_options": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "MySQL target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The MySQL server hostname or IP address",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The MySQL server port",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The MySQL username",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The MySQL password",
						},
						"tls": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Computed:    true,
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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "PostgreSQL target options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The PostgreSQL server hostname or IP address",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The PostgreSQL server port",
						},
						"username": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The PostgreSQL username",
						},
						"password": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "The PostgreSQL password",
						},
						"tls": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "TLS configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "TLS mode (Disabled, Preferred, Required)",
									},
									"verify": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Verify TLS certificates",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// dataSourceTargetRead retrieves target data from Warpgate by ID and populates
// the Terraform state.
func dataSourceTargetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	id := d.Get("id").(string)

	target, err := c.GetTarget(ctx, id)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read target: %w", err))
	}

	if target == nil {
		return diag.Errorf("target with ID %s not found", id)
	}

	d.SetId(target.ID)
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
