<a href="https://zerodha.tech"><img src="https://zerodha.tech/static/images/github-badge.svg" align="right" /></a>

# Terraform Provider for Warpgate

This Terraform provider allows you to manage [WarpGate](https://github.com/warp-tech/warpgate) resources through Terraform. Warpgate is a smart SSH and HTTPS bastion that provides secure access to your infrastructure.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18 (to build the provider)
- [Warpgate](https://github.com/warp-tech/warpgate) >= 0.13.2

## Building the Provider

1. Clone the repository:

```sh
git clone https://github.com/warp-tech/terraform-provider-warpgate.git
cd terraform-provider-warpgate
```

2. Build the provider:

```sh
make build
```

3. Install the provider for local development:

```sh
make install
```

This will build and install the provider into your `~/.terraform.d/plugins` directory (or equivalent on Windows/macOS).

## Using the Provider

To use the provider, define it in your Terraform configuration:

```hcl
terraform {
  required_providers {
    warpgate = {
      source  = "registry.terraform.io/warp-tech/warpgate"
      version = "~> 1.0.0"
    }
  }
}

provider "warpgate" {
  host  = "https://warpgate.example.com"
  token = var.warpgate_token
}
```

You can also use environment variables to configure the provider:

```sh
export WARPGATE_HOST="https://warpgate.example.com"
export WARPGATE_TOKEN="your-api-token"
```

### Resources and Data Sources

#### Resources

- `warpgate_role` - Manage Warpgate roles
- `warpgate_user` - Manage Warpgate users
- `warpgate_target` - Manage Warpgate targets (SSH, HTTP, MySQL, PostgreSQL)
- `warpgate_user_role` - Manage role assignments to users
- `warpgate_target_role` - Manage role assignments to targets

#### Data Sources

- `warpgate_role` - Retrieve information about a Warpgate role
- `warpgate_user` - Retrieve information about a Warpgate user
- `warpgate_target` - Retrieve information about a Warpgate target

## Example Usage

### Creating a User

```hcl
resource "warpgate_user" "example" {
  username    = "eugene"
  description = "Eugene - WarpGate Developer"
  
  credential_policy {
    http     = ["Password", "Totp"]
    ssh      = ["PublicKey"]
    mysql    = ["Password"]
    postgres = ["Password"]
  }
}
```

### Creating a Role

```hcl
resource "warpgate_role" "developers" {
  name        = "developers"
  description = "Role for development team"
}
```

### Assigning a Role to a User

```hcl
resource "warpgate_user_role" "developer_role" {
  user_id = warpgate_user.example.id
  role_id = warpgate_role.developers.id
}
```

### Creating an SSH Target

```hcl
resource "warpgate_target" "app_server" {
  name        = "app-server"
  description = "Application Server"
  
  ssh_options {
    host     = "10.0.0.10"
    port     = 22
    username = "admin"
    
    # Use either password_auth or public_key_auth
    password_auth {
      password = var.ssh_password
    }
    
    # OR
    # public_key_auth {}
  }
}
```

### Creating an HTTP Target

```hcl
resource "warpgate_target" "web_app" {
  name        = "internal-web-app"
  description = "Internal Web Application"
  
  http_options {
    url = "https://internal.example.com"
    tls {
      mode   = "Required"
      verify = true
    }
    headers = {
      "X-Custom-Header" = "value"
    }
  }
}
```

### Creating a MySQL Target

```hcl
resource "warpgate_target" "database" {
  name        = "mysql-db"
  description = "Production MySQL Database"
  
  mysql_options {
    host     = "db.example.com"
    port     = 3306
    username = "admin"
    password = var.db_password
    tls {
      mode   = "Required"
      verify = true
    }
  }
}
```

### Creating a PostgreSQL Target

```hcl
resource "warpgate_target" "postgres_db" {
  name        = "postgres-db"
  description = "Production PostgreSQL Database"
  
  postgres_options {
    host     = "postgres.example.com"
    port     = 5432
    username = "admin"
    password = var.postgres_password
    tls {
      mode   = "Required"
      verify = true
    }
  }
}
```

### Assigning a Role to a Target

```hcl
resource "warpgate_target_role" "app_server_access" {
  target_id = warpgate_target.app_server.id
  role_id   = warpgate_role.developers.id
}
```

### Using Data Sources

```hcl
data "warpgate_user" "existing_user" {
  id = "existing-user-id"
}

data "warpgate_role" "existing_role" {
  id = "existing-role-id"
}

data "warpgate_target" "existing_target" {
  id = "existing-target-id"
}
```

## Importing Existing Resources

You can import existing Warpgate resources into Terraform state:

```sh
# Import a user
terraform import warpgate_user.example user-uuid

# Import a role
terraform import warpgate_role.example role-uuid

# Import a target
terraform import warpgate_target.example target-uuid

# Import a user-role association
terraform import warpgate_user_role.example user-uuid:role-uuid

# Import a target-role association
terraform import warpgate_target_role.example target-uuid:role-uuid
```

## Authentication

The provider supports authentication using an API token. You can generate the token through the Warpgate admin interface.

## Development

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18

### Generating Documentation

```sh
make docs
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature/my-new-feature`
5. Submit a pull request

## License

This provider is distributed under the [MIT License](LICENSE).
