---
page_title: "warpgate_parameters Resource - terraform-provider-warpgate"
subcategory: ""
description: |-
  Manages global parameters in Warpgate. These settings control behavior across the entire Warpgate instance.
---

# warpgate_parameters (Resource)

Manages global parameters in Warpgate. These settings control behavior across the entire Warpgate instance. This is a singleton resource - only one `warpgate_parameters` resource should exist per Warpgate instance.

## Example Usage

```hcl
resource "warpgate_parameters" "global_settings" {
  allow_own_credential_management     = true
  rate_limit_bytes_per_second         = 1000000
  ssh_client_auth_publickey           = true
  ssh_client_auth_password            = true
  ssh_client_auth_keyboard_interactive = false
  password_login_mode                 = "Enabled"
  ticket_self_service_enabled         = true
  ticket_auto_approve_existing_access = true
  ticket_max_duration_seconds         = 86400
  ticket_max_uses                     = 5
  ticket_require_description          = true
  ticket_request_show_all_targets     = false
  target_click_action                 = "Connect"
  show_session_menu                   = true

  password_policy {
    min_length        = 12
    require_uppercase = true
    require_lowercase = true
    require_digits    = true
    require_special   = true
  }

  max_api_token_duration_seconds      = 2592000
  record_scp                          = true

  login_protection_enabled                 = true
  login_protection_retention_seconds       = 2592000
  lp_ip_max_attempts                       = 10
  lp_ip_time_window_seconds                = 300
  lp_ip_base_block_duration_seconds        = 60
  lp_ip_block_duration_multiplier          = 2
  lp_ip_max_block_duration_seconds         = 3600
  lp_ip_cooldown_reset_seconds             = 86400
  lp_user_max_attempts                     = 10
  lp_user_time_window_seconds              = 300
  lp_user_auto_unlock                      = true
  lp_user_lockout_duration_seconds         = 900
  lp_user_exempt_admins                    = true
  banner                                   = "Authorized access only"
  web_clients_enabled                      = true
  analytics_consent                        = "Off"
  analytics_normal                         = false

  ssh_host_key_verification                = "AutoAccept"
  web_auth_max_age_seconds                 = 28800
  web_approval_grace_period_seconds        = 3600

  recordings_enable                        = true
  recordings_storage {
    disk {
      path = "/var/lib/warpgate/recordings"
    }
  }
}
```

Recordings on S3-compatible object storage instead of a local disk:

```hcl
resource "warpgate_parameters" "main" {
  allow_own_credential_management = true

  recordings_enable = true
  recordings_storage {
    s3 {
      bucket     = "warpgate-recordings"
      region     = "us-east-1"
      endpoint   = "https://minio.example.com"
      path_style = true
      prefix     = "sessions/"

      static_credentials {
        access_key_id     = var.s3_access_key_id
        secret_access_key = var.s3_secret_access_key
      }

      # Or, to use the ambient AWS credential chain:
      # auto_credentials {}
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `allow_own_credential_management` - (Required) Allow users to manage their own credentials.
* `rate_limit_bytes_per_second` - (Optional) Rate limit for data transfer in bytes per second.
* `ssh_client_auth_publickey` - (Optional) Enable SSH public key authentication for clients.
* `ssh_client_auth_password` - (Optional) Enable SSH password authentication for clients.
* `ssh_client_auth_keyboard_interactive` - (Optional) Enable SSH keyboard interactive authentication for clients.
* `password_login_mode` - (Optional) How the password login form is presented on the gateway login page. Allowed values: `Enabled`, `Minimized`, `Disabled`.
* `ticket_self_service_enabled` - (Optional) Enable ticket self-service.
* `ticket_auto_approve_existing_access` - (Optional) Automatically approve ticket requests when the requester already has access.
* `ticket_max_duration_seconds` - (Optional) Maximum ticket duration in seconds.
* `ticket_max_uses` - (Optional) Maximum number of uses for tickets.
* `ticket_require_description` - (Optional) Require a description for ticket requests.
* `ticket_request_show_all_targets` - (Optional) Show all targets when requesting tickets.
* `target_click_action` - (Optional) Action to take when clicking a target. Allowed values: `Connect`, `ShowInstructions`.
* `show_session_menu` - (Optional) When enabled, Warpgate injects a session menu into HTTP sessions, allowing users to log out or return to the home page.
* `password_policy` - (Optional) Password policy rules.
* `max_api_token_duration_seconds` - (Optional) Maximum API token duration in seconds.
* `record_scp` - (Optional) Record SCP sessions.
* `login_protection_enabled` - (Optional) Enable login protection.
* `login_protection_retention_seconds` - (Optional) How long login protection records are retained, in seconds.
* `lp_ip_max_attempts` - (Optional) Maximum failed login attempts per IP address.
* `lp_ip_time_window_seconds` - (Optional) Time window for failed login attempts per IP address, in seconds.
* `lp_ip_base_block_duration_seconds` - (Optional) Base IP block duration in seconds.
* `lp_ip_block_duration_multiplier` - (Optional) Multiplier applied to repeated IP block durations.
* `lp_ip_max_block_duration_seconds` - (Optional) Maximum IP block duration in seconds.
* `lp_ip_cooldown_reset_seconds` - (Optional) Cooldown period before the IP block escalation resets, in seconds.
* `lp_user_max_attempts` - (Optional) Maximum failed login attempts per user.
* `lp_user_time_window_seconds` - (Optional) Time window for failed login attempts per user, in seconds.
* `lp_user_auto_unlock` - (Optional) Automatically unlock users after the lockout duration.
* `lp_user_lockout_duration_seconds` - (Optional) User lockout duration in seconds.
* `lp_user_exempt_admins` - (Optional) Exempt administrators from user login protection lockouts.
* `banner` - (Optional) Banner shown to clients before authentication.
* `web_clients_enabled` - (Optional) Enable the web-based session clients.
* `analytics_consent` - (Optional) Whether the instance reports anonymous usage analytics. Allowed values: `Undecided`, `Off`, `On`.
* `analytics_normal` - (Optional) Enable the normal analytics payload level.
* `ssh_host_key_verification` - (Optional) What to do when a target's SSH host key isn't in the known hosts list. Allowed values: `Prompt`, `AutoAccept`, `AutoReject`, `Ignore`.
* `web_auth_max_age_seconds` - (Optional) How long a web login stays valid before reauthentication is required, in seconds. Must be at least 1; unset means reauthentication is never required.
* `web_approval_grace_period_seconds` - (Optional) How long a remembered web approval stays valid, in seconds.
* `recordings_enable` - (Optional) Record sessions.
* `recordings_storage` - (Optional) Where session recordings are stored. Exactly one of the `disk` or `s3` blocks.
  * `disk` - Local filesystem storage.
    * `path` - (Required) Directory recordings are written to.
  * `s3` - S3 or S3-compatible object storage.
    * `bucket` - (Required) Bucket name.
    * `region` - (Required) Bucket region.
    * `endpoint` - (Optional) Custom endpoint for S3-compatible services. Empty means AWS.
    * `path_style` - (Optional) Path-style addressing, required by most S3-compatible services.
    * `prefix` - (Optional) Key prefix prepended to every object path.
    * `auto_credentials` - Authenticate with the ambient AWS credential chain. Exactly one of `auto_credentials` or `static_credentials` is required.
    * `static_credentials` - Authenticate with an explicit key pair. Exactly one of `auto_credentials` or `static_credentials` is required.
      * `access_key_id` - (Required) Access key ID.
      * `secret_access_key` - (Optional) Secret access key. The API never returns it, so it is carried over from configuration on read; omit to keep the secret already stored in Warpgate.

~> **Note** `recordings_enable` and `recordings_storage` require Warpgate 0.27 or newer, where recording storage moved out of `warpgate.yaml` into the database.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of this resource (always set to "parameters").

## Import

Parameters can be imported with a dummy ID:

```
$ terraform import warpgate_parameters.global_settings parameters
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `allow_own_credential_management` (Boolean) Allow users to manage their own credentials

### Optional

- `analytics_consent` (String) Whether the instance reports anonymous usage analytics.
- `analytics_normal` (Boolean) Enable the normal analytics payload level.
- `banner` (String) Banner shown to clients before authentication.
- `login_protection_enabled` (Boolean) Enable login protection.
- `login_protection_retention_seconds` (Number) How long login protection records are retained, in seconds.
- `lp_ip_base_block_duration_seconds` (Number) Base IP block duration in seconds.
- `lp_ip_block_duration_multiplier` (Number) Multiplier applied to repeated IP block durations.
- `lp_ip_cooldown_reset_seconds` (Number) Cooldown period before the IP block escalation resets, in seconds.
- `lp_ip_max_attempts` (Number) Maximum failed login attempts per IP address.
- `lp_ip_max_block_duration_seconds` (Number) Maximum IP block duration in seconds.
- `lp_ip_time_window_seconds` (Number) Time window for failed login attempts per IP address, in seconds.
- `lp_user_auto_unlock` (Boolean) Automatically unlock users after the lockout duration.
- `lp_user_exempt_admins` (Boolean) Exempt administrators from user login protection lockouts.
- `lp_user_lockout_duration_seconds` (Number) User lockout duration in seconds.
- `lp_user_max_attempts` (Number) Maximum failed login attempts per user.
- `lp_user_time_window_seconds` (Number) Time window for failed login attempts per user, in seconds.
- `max_api_token_duration_seconds` (Number) Maximum API token duration in seconds.
- `password_login_mode` (String) How the password login form is presented on the gateway login page.
- `password_policy` (Block List, Max: 1) Password policy rules. (see [below for nested schema](#nestedblock--password_policy))
- `rate_limit_bytes_per_second` (Number) Global bandwidth limit
- `record_scp` (Boolean) Record SCP sessions.
- `recordings_enable` (Boolean) Record sessions.
- `recordings_storage` (Block List, Max: 1) Where session recordings are stored. (see [below for nested schema](#nestedblock--recordings_storage))
- `show_session_menu` (Boolean) When enabled, Warpgate injects a session menu into HTTP sessions, allowing users to log out or return to the home page.
- `ssh_client_auth_keyboard_interactive` (Boolean) Enable SSH keyboard interactive authentication
- `ssh_client_auth_password` (Boolean) Enable SSH password authentication
- `ssh_client_auth_publickey` (Boolean) Enable SSH public key authentication
- `ssh_host_key_verification` (String) What to do when a target's SSH host key isn't in the known hosts list.
- `target_click_action` (String) Action to take when clicking a target.
- `ticket_auto_approve_existing_access` (Boolean) Automatically approve ticket requests when the requester already has access.
- `ticket_max_duration_seconds` (Number) Maximum ticket duration in seconds.
- `ticket_max_uses` (Number) Maximum number of uses for tickets.
- `ticket_request_show_all_targets` (Boolean) Show all targets when requesting tickets.
- `ticket_require_description` (Boolean) Require a description for ticket requests.
- `ticket_self_service_enabled` (Boolean) Enable ticket self-service.
- `web_approval_grace_period_seconds` (Number) How long a remembered web approval stays valid, in seconds.
- `web_auth_max_age_seconds` (Number) How long a web login stays valid before reauthentication is required, in seconds. Unset means reauthentication is never required.
- `web_clients_enabled` (Boolean) Enable the web-based session clients.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--password_policy"></a>
### Nested Schema for `password_policy`

Optional:

- `min_length` (Number) Minimum number of characters, or 0 for no requirement.
- `require_digits` (Boolean) Require at least one digit.
- `require_lowercase` (Boolean) Require at least one lowercase character.
- `require_special` (Boolean) Require at least one special character.
- `require_uppercase` (Boolean) Require at least one uppercase character.


<a id="nestedblock--recordings_storage"></a>
### Nested Schema for `recordings_storage`

Optional:

- `disk` (Block List, Max: 1) Local filesystem storage (see [below for nested schema](#nestedblock--recordings_storage--disk))
- `s3` (Block List, Max: 1) S3 or S3-compatible object storage (see [below for nested schema](#nestedblock--recordings_storage--s3))

<a id="nestedblock--recordings_storage--disk"></a>
### Nested Schema for `recordings_storage.disk`

Required:

- `path` (String) Directory recordings are written to


<a id="nestedblock--recordings_storage--s3"></a>
### Nested Schema for `recordings_storage.s3`

Required:

- `bucket` (String) Bucket name
- `region` (String) Bucket region

Optional:

- `auto_credentials` (Block List, Max: 1) Authenticate with the ambient AWS credential chain (see [below for nested schema](#nestedblock--recordings_storage--s3--auto_credentials))
- `endpoint` (String) Custom endpoint for S3-compatible services. Empty means AWS.
- `path_style` (Boolean) Path-style addressing, required by most S3-compatible services
- `prefix` (String) Key prefix prepended to every object path
- `static_credentials` (Block List, Max: 1) Authenticate with an explicit key pair (see [below for nested schema](#nestedblock--recordings_storage--s3--static_credentials))

<a id="nestedblock--recordings_storage--s3--auto_credentials"></a>
### Nested Schema for `recordings_storage.s3.auto_credentials`


<a id="nestedblock--recordings_storage--s3--static_credentials"></a>
### Nested Schema for `recordings_storage.s3.static_credentials`

Required:

- `access_key_id` (String) Access key ID

Optional:

- `secret_access_key` (String, Sensitive) Secret access key. The API never returns it, so it is carried over from configuration on read; omit to keep the secret already stored in Warpgate.
