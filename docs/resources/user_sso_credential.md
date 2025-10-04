---
page_title: "warpgate_user_sso_credential Resource - terraform-provider-warpgate"
subcategory: ""
description: |-
  Manages an SSO credential for a user in WarpGate.
---

# warpgate_user_sso_credential (Resource)

Manages an SSO (Single Sign-On) credential for a user in WarpGate. SSO credentials allow users to authenticate using external identity providers such as Google, GitHub, Okta, Azure AD, and other SAML/OIDC providers.

## Example Usage

```hcl
# Create a user with SSO credential policy
resource "warpgate_user" "john_doe" {
  username    = "john.doe"
  description = "John Doe - Software Engineer"

  credential_policy {
    http     = ["Sso"]
    ssh      = ["Sso", "PublicKey"]
    mysql    = ["Sso"]
    postgres = ["Sso"]
  }
}

# Add Google SSO credential
resource "warpgate_user_sso_credential" "john_google" {
  user_id  = warpgate_user.john_doe.id
  sso_provider = "google"
  email    = "john.doe@company.com"
}

# Add GitHub SSO credential for the same user
resource "warpgate_user_sso_credential" "john_github" {
  user_id  = warpgate_user.john_doe.id
  sso_provider = "github"
  email    = "john.doe@company.com"
}

# Add Okta SSO credential
resource "warpgate_user_sso_credential" "john_okta" {
  user_id  = warpgate_user.john_doe.id
  sso_provider = "okta"
  email    = "john.doe@company.com"
}
```

## Multiple SSO Providers

You can configure multiple SSO providers for the same user:

```hcl
resource "warpgate_user" "multi_sso_user" {
  username    = "alice"
  description = "Alice - Multi-SSO user"

  credential_policy {
    http = ["Sso"]
    ssh  = ["Sso", "PublicKey"]
  }
}

# Primary corporate SSO
resource "warpgate_user_sso_credential" "alice_okta" {
  user_id  = warpgate_user.multi_sso_user.id
  sso_provider = "okta"
  email    = "alice@company.com"
}

# Backup SSO for personal projects
resource "warpgate_user_sso_credential" "alice_google" {
  user_id  = warpgate_user.multi_sso_user.id
  sso_provider = "google"
  email    = "alice.personal@gmail.com"
}
```

## Argument Reference

The following arguments are supported:

* `user_id` - (Required, Forces new resource) The ID of the user to add the SSO credential to. Changing this forces a new resource to be created.
* `sso_provider` - (Required) The SSO provider name. Common values include `google`, `github`, `okta`, `azure`, or any custom SAML/OIDC provider configured in WarpGate.
* `email` - (Required) The email address associated with the SSO provider. This should match the email address in the user's SSO provider account.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the SSO credential.

## Supported SSO Providers

WarpGate supports various SSO providers. Common providers include:

### OAuth/OIDC Providers
- `google` - Google Workspace / Gmail OAuth
- `github` - GitHub OAuth
- `microsoft` - Microsoft Azure AD / Office 365
- `gitlab` - GitLab OAuth
- `discord` - Discord OAuth

### Enterprise Identity Providers
- `okta` - Okta SAML/OIDC
- `azure` - Azure Active Directory
- `auth0` - Auth0 identity platform
- `ping` - PingIdentity
- `adfs` - Active Directory Federation Services

### Custom Providers
- Any SAML 2.0 compatible provider
- Any OpenID Connect (OIDC) compatible provider

## Prerequisites

Before creating SSO credentials, ensure that:

1. The SSO provider is properly configured in your WarpGate instance
2. The user's email address exists in the SSO provider
3. The user has appropriate permissions in the SSO provider
4. The user's credential policy includes `Sso` for the desired protocols

## Import

SSO credentials can be imported using a composite ID in the format `user_id:credential_id`:

```
$ terraform import warpgate_user_sso_credential.example 12345678-1234-1234-1234-123456789012:87654321-4321-4321-4321-210987654321
```

To find the credential ID, you can use the `warpgate_user` data source:

```hcl
data "warpgate_user" "existing_user" {
  id = "12345678-1234-1234-1234-123456789012"
}

# The credential IDs will be available in:
# data.warpgate_user.existing_user.sso_credentials[*].id
```

## Common Patterns

### Service Account with SSO

```hcl
resource "warpgate_user" "service_account" {
  username    = "ci-cd-bot"
  description = "CI/CD Service Account"

  credential_policy {
    http = ["Sso"]
    ssh  = ["PublicKey"]  # Use public key for SSH, SSO for HTTP
  }
}

resource "warpgate_user_sso_credential" "service_google" {
  user_id  = warpgate_user.service_account.id
  sso_provider = "google"
  email    = "ci-cd-bot@company.com"
}
```

### Development vs Production SSO

```hcl
# Development environment - multiple SSO options
resource "warpgate_user" "dev_user" {
  username    = "developer"
  description = "Development user with flexible SSO"

  credential_policy {
    http = ["Sso", "Password"]  # Allow both SSO and password
    ssh  = ["Sso", "PublicKey"]
  }
}

# Production environment - strict SSO only
resource "warpgate_user" "prod_user" {
  username    = "prod-developer"
  description = "Production user with strict SSO"

  credential_policy {
    http     = ["Sso"]  # SSO only
    ssh      = ["Sso"]
    mysql    = ["Sso"]
    postgres = ["Sso"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `email` (String) The email address associated with the SSO provider
- `sso_provider` (String) The SSO provider name (e.g., 'google', 'github', 'okta')
- `user_id` (String) The ID of the user to add the SSO credential to

### Read-Only

- `id` (String) The ID of this resource.
