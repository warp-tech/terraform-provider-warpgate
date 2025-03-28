---
page_title: "warpgate_role Resource - terraform-provider-warpgate"
subcategory: ""
description: |-
  Manages a role in WarpGate. Roles define permissions that can be assigned to users and targets.
---

# warpgate_role (Resource)

Manages a role in WarpGate. Roles are the core components of WarpGate's access control system. A role can be assigned to both users and targets, establishing a many-to-many relationship that determines which users can access which targets.

## Example Usage

```hcl
resource "warpgate_role" "developers" {
  name        = "developers"
  description = "Role for development team members"
}

resource "warpgate_role" "admins" {
  name        = "administrators"
  description = "Role for system administrators with full access"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the role. Must be unique within the WarpGate instance.
* `description` - (Optional) A human-readable description of the role and its purpose.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the role.

## Import

Roles can be imported using their ID:

```
$ terraform import warpgate_role.developers 12345678-1234-1234-1234-123456789012
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the role

### Optional

- `description` (String) The description of the role

### Read-Only

- `id` (String) The ID of this resource.
