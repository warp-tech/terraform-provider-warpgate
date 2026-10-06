// Package provider implements the Terraform provider for Warpgate
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/warp-tech/terraform-provider-warpgate/internal/client"
)

// adminRolePermissionAttribute ties a boolean attribute of the admin role
// schema to the matching field of the API permission set, so the schema and
// the state round-trips are driven by one table.
type adminRolePermissionAttribute struct {
	key         string
	description string
	field       func(*client.AdminRolePermissions) *bool
}

// adminRolePermissionAttributes lists every admin permission Warpgate knows, in
// the order the admin UI shows them.
var adminRolePermissionAttributes = []adminRolePermissionAttribute{
	{"targets_create", "Allow creating targets", func(p *client.AdminRolePermissions) *bool { return &p.TargetsCreate }},
	{"targets_edit", "Allow editing targets", func(p *client.AdminRolePermissions) *bool { return &p.TargetsEdit }},
	{"targets_delete", "Allow deleting targets", func(p *client.AdminRolePermissions) *bool { return &p.TargetsDelete }},
	{"users_create", "Allow creating users", func(p *client.AdminRolePermissions) *bool { return &p.UsersCreate }},
	{"users_edit", "Allow editing users", func(p *client.AdminRolePermissions) *bool { return &p.UsersEdit }},
	{"users_delete", "Allow deleting users", func(p *client.AdminRolePermissions) *bool { return &p.UsersDelete }},
	{"access_roles_create", "Allow creating access roles", func(p *client.AdminRolePermissions) *bool { return &p.AccessRolesCreate }},
	{"access_roles_edit", "Allow editing access roles", func(p *client.AdminRolePermissions) *bool { return &p.AccessRolesEdit }},
	{"access_roles_delete", "Allow deleting access roles", func(p *client.AdminRolePermissions) *bool { return &p.AccessRolesDelete }},
	{"access_roles_assign", "Allow assigning access roles to users and targets", func(p *client.AdminRolePermissions) *bool { return &p.AccessRolesAssign }},
	{"sessions_view", "Allow viewing sessions", func(p *client.AdminRolePermissions) *bool { return &p.SessionsView }},
	{"sessions_terminate", "Allow terminating sessions", func(p *client.AdminRolePermissions) *bool { return &p.SessionsTerminate }},
	{"approve_sessions", "Allow approving sessions that wait for an administrator", func(p *client.AdminRolePermissions) *bool { return &p.ApproveSessions }},
	{"recordings_view", "Allow viewing session recordings", func(p *client.AdminRolePermissions) *bool { return &p.RecordingsView }},
	{"tickets_create", "Allow creating tickets", func(p *client.AdminRolePermissions) *bool { return &p.TicketsCreate }},
	{"tickets_delete", "Allow deleting tickets", func(p *client.AdminRolePermissions) *bool { return &p.TicketsDelete }},
	{"config_edit", "Allow editing the global parameters", func(p *client.AdminRolePermissions) *bool { return &p.ConfigEdit }},
	{"admin_roles_manage", "Allow managing admin roles and their assignments", func(p *client.AdminRolePermissions) *bool { return &p.AdminRolesManage }},
	{"ticket_requests_manage", "Allow approving and denying ticket requests", func(p *client.AdminRolePermissions) *bool { return &p.TicketRequestsManage }},
}

// adminRolePermissionsSchema returns one optional boolean attribute per admin
// permission. Computed is set for data sources, where the attributes are
// read-only.
func adminRolePermissionsSchema(computed bool) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema, len(adminRolePermissionAttributes))
	for _, attribute := range adminRolePermissionAttributes {
		s := &schema.Schema{
			Type:        schema.TypeBool,
			Description: attribute.description,
		}
		if computed {
			s.Computed = true
		} else {
			s.Optional = true
			s.Default = false
		}
		result[attribute.key] = s
	}
	return result
}

// adminRolePermissionsFromResourceData collects the permission attributes into
// an API permission set.
func adminRolePermissionsFromResourceData(d *schema.ResourceData) client.AdminRolePermissions {
	var permissions client.AdminRolePermissions
	for _, attribute := range adminRolePermissionAttributes {
		*attribute.field(&permissions) = d.Get(attribute.key).(bool)
	}
	return permissions
}

// setAdminRoleData writes an admin role into the Terraform state.
func setAdminRoleData(d *schema.ResourceData, role *client.AdminRole) error {
	if err := d.Set("name", role.Name); err != nil {
		return fmt.Errorf("failed to set name: %w", err)
	}

	if err := d.Set("description", role.Description); err != nil {
		return fmt.Errorf("failed to set description: %w", err)
	}

	for _, attribute := range adminRolePermissionAttributes {
		if err := d.Set(attribute.key, *attribute.field(&role.AdminRolePermissions)); err != nil {
			return fmt.Errorf("failed to set %s: %w", attribute.key, err)
		}
	}

	return nil
}

// resourceAdminRole creates and returns a schema for the admin role resource.
func resourceAdminRole() *schema.Resource {
	s := map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "The name of the admin role",
			ValidateFunc: validation.StringLenBetween(1, 255),
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description of the admin role",
		},
	}
	for key, permission := range adminRolePermissionsSchema(false) {
		s[key] = permission
	}

	return &schema.Resource{
		CreateContext: resourceAdminRoleCreate,
		ReadContext:   resourceAdminRoleRead,
		UpdateContext: resourceAdminRoleUpdate,
		DeleteContext: resourceAdminRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: s,
	}
}

func adminRoleRequestFromResourceData(d *schema.ResourceData) *client.AdminRoleRequest {
	return &client.AdminRoleRequest{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		AdminRolePermissions: adminRolePermissionsFromResourceData(d),
	}
}

// resourceAdminRoleCreate handles the creation of a new admin role in Warpgate
// based on the provided resource data.
func resourceAdminRoleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	role, err := c.CreateAdminRole(ctx, adminRoleRequestFromResourceData(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create admin role: %w", err))
	}

	d.SetId(role.ID)

	return diags
}

// resourceAdminRoleRead retrieves the admin role data from Warpgate and updates
// the Terraform state accordingly.
func resourceAdminRoleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	role, err := c.GetAdminRole(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to read admin role: %w", err))
	}

	// If the admin role was not found, return nil to indicate that the resource no longer exists
	if role == nil {
		d.SetId("")
		return diags
	}

	if err := setAdminRoleData(d, role); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// resourceAdminRoleUpdate handles the update of an existing admin role in
// Warpgate based on the provided resource data changes.
func resourceAdminRoleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	_, err := c.UpdateAdminRole(ctx, d.Id(), adminRoleRequestFromResourceData(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update admin role: %w", err))
	}

	return resourceAdminRoleRead(ctx, d, meta)
}

// resourceAdminRoleDelete removes an admin role from Warpgate based on the
// resource data.
func resourceAdminRoleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	providerMeta := meta.(*providerMeta)
	c := providerMeta.client

	var diags diag.Diagnostics

	err := c.DeleteAdminRole(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete admin role: %w", err))
	}

	d.SetId("")

	return diags
}
