---
page_title: "warpgate_target_role Resource - terraform-provider-warpgate"
subcategory: ""
description: |-
  Creates an association between a target and a role in WarpGate.
---

# warpgate_target_role (Resource)

Creates an association between a target and a role in WarpGate. This relationship makes the target accessible to users who have the same role assigned.

For SSH targets, you can also configure file transfer (SFTP) permissions using the `file_transfer` block.

## Example Usage

### Basic Usage

```hcl
resource "warpgate_target" "web_server" {
  name = "web-server"
  
  ssh_options {
    host     = "10.0.0.1"
    port     = 22
    username = "admin"
    public_key_auth {}
  }
}

resource "warpgate_role" "developers" {
  name = "developers"
}

resource "warpgate_target_role" "developers_web_access" {
  target_id = warpgate_target.web_server.id
  role_id   = warpgate_role.developers.id
}
```

### With File Transfer Restrictions (SSH Targets Only)

```hcl
resource "warpgate_target" "production_server" {
  name = "production-server"
  
  ssh_options {
    host     = "10.0.0.1"
    port     = 22
    username = "deploy"
    public_key_auth {}
  }
}

resource "warpgate_role" "developers" {
  name = "developers"
}

# Allow SSH access but restrict file transfers
resource "warpgate_target_role" "developers_production" {
  target_id = warpgate_target.production_server.id
  role_id   = warpgate_role.developers.id

  file_transfer {
    allow_upload       = true
    allow_download     = false  # Prevent data exfiltration
    allowed_paths      = ["/home/deploy", "/var/log"]
    blocked_extensions = [".exe", ".sh", ".sql"]
    max_file_size      = 10485760  # 10 MB
  }
}
```

### Disable All File Transfers

```hcl
resource "warpgate_target_role" "readonly_access" {
  target_id = warpgate_target.sensitive_server.id
  role_id   = warpgate_role.auditors.id

  file_transfer {
    allow_upload   = false
    allow_download = false
  }
}
```

### SFTP-Only Access (Override per Target)

```hcl
resource "warpgate_target_role" "sftp_only_access" {
  target_id = warpgate_target.file_server.id
  role_id   = warpgate_role.developers.id

  file_transfer {
    file_transfer_only = true  # Blocks shell/exec/forwarding for this target
  }
}
```

## Complete Access Control Example

This example demonstrates a complete access control setup with users, roles, targets, and their relationships:

```hcl
# Create roles
resource "warpgate_role" "developers" {
  name        = "developers"
  description = "Development team members"
}

resource "warpgate_role" "administrators" {
  name        = "administrators"
  description = "System administrators"
}

# Create users
resource "warpgate_user" "eugene" {
  username    = "eugene"
  description = "Developer"
  
  credential_policy {
    ssh = ["Password", "PublicKey"]
  }
}

resource "warpgate_user" "jane" {
  username    = "jane.smith"
  description = "Administrator"
  
  credential_policy {
    ssh = ["Password", "PublicKey"]
  }
}

# Create targets
resource "warpgate_target" "web_server" {
  name        = "web-server"
  description = "Web server"
  
  ssh_options {
    host     = "10.0.0.1"
    port     = 22
    username = "webadmin"
    public_key_auth {}
  }
}

resource "warpgate_target" "db_server" {
  name        = "db-server"
  description = "Database server"
  
  ssh_options {
    host     = "10.0.0.2"
    port     = 22
    username = "dbadmin"
    public_key_auth {}
  }
}

# Assign roles to users
resource "warpgate_user_role" "eugene_developer" {
  user_id = warpgate_user.eugene.id
  role_id = warpgate_role.developers.id
}

resource "warpgate_user_role" "jane_administrator" {
  user_id = warpgate_user.jane.id
  role_id = warpgate_role.administrators.id
}

# Assign roles to targets with file transfer restrictions
resource "warpgate_target_role" "developers_web_access" {
  target_id = warpgate_target.web_server.id
  role_id   = warpgate_role.developers.id

  # Developers can upload but not download from web server
  file_transfer {
    allow_upload   = true
    allow_download = false
  }
}

resource "warpgate_target_role" "admins_web_access" {
  target_id = warpgate_target.web_server.id
  role_id   = warpgate_role.administrators.id
  # No file_transfer block = full access (defaults: allow_upload=true, allow_download=true)
}

resource "warpgate_target_role" "admins_db_access" {
  target_id = warpgate_target.db_server.id
  role_id   = warpgate_role.administrators.id

  # Restrict file transfers on database server
  file_transfer {
    allow_upload       = true
    allow_download     = true
    allowed_paths      = ["/var/log", "/tmp/exports"]
    blocked_extensions = [".sql", ".dump", ".bak"]
    max_file_size      = 52428800  # 50 MB
  }
}
```

This setup results in the following access matrix:
- Eugene (developer) can access the web server (SSH + upload only)
- Jane (administrator) can access both servers with full file transfer on web, restricted on db

## Argument Reference

The following arguments are supported:

* `target_id` - (Required) The ID of the target to assign the role to.
* `role_id` - (Required) The ID of the role to assign.
* `file_transfer` - (Optional) File transfer (SFTP) permission settings. **Only applicable for SSH targets.** See [File Transfer](#file-transfer) below.

### File Transfer

The `file_transfer` block supports the following arguments:

* `allow_upload` - (Optional) Allow file uploads via SFTP. Defaults to `true`.
* `allow_download` - (Optional) Allow file downloads via SFTP. Defaults to `true`.
* `allowed_paths` - (Optional) List of allowed paths for file transfers. If not specified, all paths are allowed.
* `blocked_extensions` - (Optional) List of blocked file extensions (e.g., `[".exe", ".sh"]`). If not specified, no extensions are blocked.
* `max_file_size` - (Optional) Maximum file size in bytes. If not specified, no size limit is enforced.
* `file_transfer_only` - (Optional) When `true`, blocks shell, exec, and port forwarding — only SFTP is allowed. Values: `"inherit"` (from role default), `"true"`, `"false"`. Defaults to `"inherit"`.

~> **Note:** The `file_transfer` block is only applicable for SSH targets. Setting it on non-SSH targets (HTTP, MySQL, PostgreSQL) will result in an error.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The combined ID in the format `target_id:role_id`.

## Import

Target-role associations can be imported using a combined ID with the format `target_id:role_id`:

```
$ terraform import warpgate_target_role.developers_web_access 12345678-1234-1234-1234-123456789012:87654321-4321-4321-4321-210987654321
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `role_id` (String) The ID of the role to assign
- `target_id` (String) The ID of the target to assign the role to

### Optional

- `file_transfer` (Block List, Max: 1) File transfer (SFTP) permissions. Only applicable for SSH targets. (see [below for nested schema](#nestedblock--file_transfer))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--file_transfer"></a>
### Nested Schema for `file_transfer`

Optional:

- `allow_download` (Boolean) Allow file downloads via SFTP. Defaults to `true`.
- `allow_upload` (Boolean) Allow file uploads via SFTP. Defaults to `true`.
- `allowed_paths` (List of String) Allowed paths for file transfers (null = all paths allowed).
- `blocked_extensions` (List of String) Blocked file extensions (null = no extensions blocked).
- `max_file_size` (Number) Maximum file size in bytes (null = no limit).
