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
  ssh_banner                               = "Authorized access only"
  web_ssh_enabled                          = true
  analytics_consent                        = "Off"
  analytics_normal                         = false
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
* `ssh_banner` - (Optional) Banner shown to SSH clients before authentication.
* `web_ssh_enabled` - (Optional) Enable web-based SSH sessions.
* `analytics_consent` - (Optional) Whether the instance reports anonymous usage analytics. Allowed values: `Undecided`, `Off`, `On`.
* `analytics_normal` - (Optional) Enable the normal analytics payload level.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of this resource (always set to "parameters").

## Import

Parameters can be imported with a dummy ID (e.g. `parameters`).

### Declarative Import Block (Terraform 1.5+ / OpenTofu 1.6+)

```hcl
import {
  to = warpgate_parameters.global_settings
  id = "parameters"
}
```

### CLI Command (Terraform / OpenTofu)

```shell
# Using Terraform CLI
terraform import warpgate_parameters.global_settings parameters

# Using OpenTofu CLI
tofu import warpgate_parameters.global_settings parameters
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `allow_own_credential_management` (Boolean) Allow users to manage their own credentials

### Optional

- `analytics_consent` (String) Whether the instance reports anonymous usage analytics.
- `analytics_normal` (Boolean) Enable the normal analytics payload level.
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
- `show_session_menu` (Boolean) When enabled, Warpgate injects a session menu into HTTP sessions, allowing users to log out or return to the home page.
- `ssh_banner` (String) Banner shown to SSH clients before authentication.
- `ssh_client_auth_keyboard_interactive` (Boolean) Enable SSH keyboard interactive authentication
- `ssh_client_auth_password` (Boolean) Enable SSH password authentication
- `ssh_client_auth_publickey` (Boolean) Enable SSH public key authentication
- `target_click_action` (String) Action to take when clicking a target.
- `ticket_auto_approve_existing_access` (Boolean) Automatically approve ticket requests when the requester already has access.
- `ticket_max_duration_seconds` (Number) Maximum ticket duration in seconds.
- `ticket_max_uses` (Number) Maximum number of uses for tickets.
- `ticket_request_show_all_targets` (Boolean) Show all targets when requesting tickets.
- `ticket_require_description` (Boolean) Require a description for ticket requests.
- `ticket_self_service_enabled` (Boolean) Enable ticket self-service.
- `web_ssh_enabled` (Boolean) Enable web-based SSH sessions.

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
