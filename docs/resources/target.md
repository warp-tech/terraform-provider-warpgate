---
page_title: "warpgate_target Resource - terraform-provider-warpgate"
subcategory: ""
description: |-
  Manages a target in WarpGate. A target represents a destination system that users can connect to.
---

# warpgate_target (Resource)

Manages a target in WarpGate. A target represents a destination system that users can connect to, such as an SSH server, HTTP service, or database. Targets define the connection details and authentication methods used to access the underlying service.

## Example Usage

### SSH Target

```hcl
resource "warpgate_target" "web_server" {
  name        = "web-server"
  description = "Production web server"
  
  ssh_options {
    host     = "10.0.0.1"
    port     = 22
    username = "admin"
    
    # You must choose either password_auth or public_key_auth
    password_auth {
      password = "supersecret"
    }
    
    # OR
    # public_key_auth {}
  }
}
```

### HTTP Target

```hcl
resource "warpgate_target" "api_server" {
  name        = "api-server"
  description = "Internal API server"
  
  http_options {
    url = "https://api.internal.example.com"
    tls {
      mode   = "Required"
      verify = true
    }
    
    headers = {
      "X-API-Version" = "v1"
      "X-Custom-Header" = "custom-value"
    }
    
    external_host = "api.external.example.com"  # Optional
  }
}
```

### MySQL Target

```hcl
resource "warpgate_target" "mysql_db" {
  name        = "mysql-db"
  description = "Production MySQL database"
  
  mysql_options {
    host     = "db.example.com"
    port     = 3306
    username = "app_user"
    password = "dbpassword"
    tls {
      mode   = "Required"
      verify = true
    }
  }
}
```

### PostgreSQL Target

```hcl
resource "warpgate_target" "postgres_db" {
  name        = "postgres-db"
  description = "Analytics PostgreSQL database"
  
  postgres_options {
    host     = "analytics-db.example.com"
    port     = 5432
    username = "analyst"
    password = "dbpassword"
    tls {
      mode   = "Required"
      verify = true
    }
  }
}
```

### Kubernetes Target (Token Auth)

```hcl
resource "warpgate_target" "k8s_cluster" {
  name        = "k8s-cluster"
  description = "Production Kubernetes cluster"
  
  kubernetes_options {
    cluster_url = "https://k8s.example.com:6443"
    tls {
      mode   = "Required"
      verify = true
    }
    
    token_auth {
      token = "eyJhbGciOiJSUzI1NiIs..."
    }
  }
}
```

### Kubernetes Target (Certificate Auth)

```hcl
resource "warpgate_target" "k8s_cluster_cert" {
  name        = "k8s-cluster-cert"
  description = "Kubernetes cluster with certificate auth"
  
  kubernetes_options {
    cluster_url = "https://k8s.example.com:6443"
    tls {
      mode   = "Required"
      verify = true
    }
    
    certificate_auth {
      certificate = file("client.crt")
      private_key = file("client.key")
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the target. Must be unique within the WarpGate instance.
* `description` - (Optional) A human-readable description of the target.
* `group_id` - (Optional) The ID of the target group this target is assigned to.

One of the following option blocks must be specified:

* `ssh_options` - (Optional) SSH target configuration block.
  * `host` - (Required) The SSH server hostname or IP address.
  * `port` - (Required) The SSH server port.
  * `username` - (Required) The SSH username.
  * `allow_insecure_algos` - (Optional) Allow insecure SSH algorithms. Default: `false`.
  * `password_auth` - (Optional) Password authentication for SSH. Conflicts with `public_key_auth`.
    * `password` - (Required) The password for SSH authentication.
  * `public_key_auth` - (Optional) Public key authentication for SSH. Conflicts with `password_auth`. No additional properties needed.

* `http_options` - (Optional) HTTP target configuration block.
  * `url` - (Required) The HTTP server URL.
  * `tls` - (Required) TLS configuration block.
    * `mode` - (Required) TLS mode. Valid values: `Disabled`, `Preferred`, `Required`.
    * `verify` - (Required) Verify TLS certificates.
  * `headers` - (Optional) HTTP headers to include in requests.
  * `external_host` - (Optional) External host for HTTP requests.

* `mysql_options` - (Optional) MySQL target configuration block.
  * `host` - (Required) The MySQL server hostname or IP address.
  * `port` - (Required) The MySQL server port.
  * `username` - (Required) The MySQL username.
  * `password` - (Optional) The MySQL password.
  * `tls` - (Required) TLS configuration block.
    * `mode` - (Required) TLS mode. Valid values: `Disabled`, `Preferred`, `Required`.
    * `verify` - (Required) Verify TLS certificates.

* `postgres_options` - (Optional) PostgreSQL target configuration block.
  * `host` - (Required) The PostgreSQL server hostname or IP address.
  * `port` - (Required) The PostgreSQL server port.
  * `username` - (Required) The PostgreSQL username.
  * `password` - (Optional) The PostgreSQL password.
  * `tls` - (Required) TLS configuration block.
    * `mode` - (Required) TLS mode. Valid values: `Disabled`, `Preferred`, `Required`.
    * `verify` - (Required) Verify TLS certificates.

* `kubernetes_options` - (Optional) Kubernetes target configuration block.
  * `cluster_url` - (Required) The Kubernetes cluster URL.
  * `tls` - (Required) TLS configuration block.
    * `mode` - (Required) TLS mode. Valid values: `Disabled`, `Preferred`, `Required`.
    * `verify` - (Required) Verify TLS certificates.
  * `token_auth` - (Optional) Token authentication for Kubernetes. Conflicts with `certificate_auth`.
    * `token` - (Required) The bearer token for Kubernetes authentication.
  * `certificate_auth` - (Optional) Certificate authentication for Kubernetes. Conflicts with `token_auth`.
    * `certificate` - (Required) The client certificate PEM.
    * `private_key` - (Required) The client private key PEM.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the target.
* `allow_roles` - The list of roles allowed to access this target (computed from role assignments).

## Import

Targets can be imported using their ID:

```
$ terraform import warpgate_target.web_server 12345678-1234-1234-1234-123456789012
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the target

### Optional

- `description` (String) The description of the target
- `group_id` (String) Which target group this target is assigned to
- `http_options` (Block List, Max: 1) HTTP target options (see [below for nested schema](#nestedblock--http_options))
- `kubernetes_options` (Block List, Max: 1) Kubernetes target options (see [below for nested schema](#nestedblock--kubernetes_options))
- `mysql_options` (Block List, Max: 1) MySQL target options (see [below for nested schema](#nestedblock--mysql_options))
- `postgres_options` (Block List, Max: 1) PostgreSQL target options (see [below for nested schema](#nestedblock--postgres_options))
- `ssh_options` (Block List, Max: 1) SSH target options (see [below for nested schema](#nestedblock--ssh_options))

### Read-Only

- `allow_roles` (List of String) The list of roles allowed to access this target
- `id` (String) The ID of this resource.

<a id="nestedblock--http_options"></a>
### Nested Schema for `http_options`

Required:

- `tls` (Block List, Min: 1, Max: 1) TLS configuration (see [below for nested schema](#nestedblock--http_options--tls))
- `url` (String) The HTTP server URL

Optional:

- `external_host` (String) External host for HTTP requests
- `headers` (Map of String) HTTP headers to include in requests

<a id="nestedblock--http_options--tls"></a>
### Nested Schema for `http_options.tls`

Required:

- `mode` (String) TLS mode (Disabled, Preferred, Required)
- `verify` (Boolean) Verify TLS certificates



<a id="nestedblock--kubernetes_options"></a>
### Nested Schema for `kubernetes_options`

Required:

- `cluster_url` (String) The Kubernetes cluster URL
- `tls` (Block List, Min: 1, Max: 1) TLS configuration (see [below for nested schema](#nestedblock--kubernetes_options--tls))

Optional:

- `certificate_auth` (Block List, Max: 1) Certificate authentication for Kubernetes (see [below for nested schema](#nestedblock--kubernetes_options--certificate_auth))
- `token_auth` (Block List, Max: 1) Token authentication for Kubernetes (see [below for nested schema](#nestedblock--kubernetes_options--token_auth))

<a id="nestedblock--kubernetes_options--tls"></a>
### Nested Schema for `kubernetes_options.tls`

Required:

- `mode` (String) TLS mode (Disabled, Preferred, Required)
- `verify` (Boolean) Verify TLS certificates

<a id="nestedblock--kubernetes_options--token_auth"></a>
### Nested Schema for `kubernetes_options.token_auth`

Required:

- `token` (String, Sensitive) The bearer token for Kubernetes authentication

<a id="nestedblock--kubernetes_options--certificate_auth"></a>
### Nested Schema for `kubernetes_options.certificate_auth`

Required:

- `certificate` (String) The client certificate PEM
- `private_key` (String, Sensitive) The client private key PEM


<a id="nestedblock--mysql_options"></a>
### Nested Schema for `mysql_options`

Required:

- `host` (String) The MySQL server hostname or IP address
- `port` (Number) The MySQL server port
- `tls` (Block List, Min: 1, Max: 1) TLS configuration (see [below for nested schema](#nestedblock--mysql_options--tls))
- `username` (String) The MySQL username

Optional:

- `password` (String, Sensitive) The MySQL password

<a id="nestedblock--mysql_options--tls"></a>
### Nested Schema for `mysql_options.tls`

Required:

- `mode` (String) TLS mode (Disabled, Preferred, Required)
- `verify` (Boolean) Verify TLS certificates



<a id="nestedblock--postgres_options"></a>
### Nested Schema for `postgres_options`

Required:

- `host` (String) The PostgreSQL server hostname or IP address
- `port` (Number) The PostgreSQL server port
- `tls` (Block List, Min: 1, Max: 1) TLS configuration (see [below for nested schema](#nestedblock--postgres_options--tls))
- `username` (String) The PostgreSQL username

Optional:

- `password` (String, Sensitive) The PostgreSQL password

<a id="nestedblock--postgres_options--tls"></a>
### Nested Schema for `postgres_options.tls`

Required:

- `mode` (String) TLS mode (Disabled, Preferred, Required)
- `verify` (Boolean) Verify TLS certificates



<a id="nestedblock--ssh_options"></a>
### Nested Schema for `ssh_options`

Required:

- `host` (String) The SSH server hostname or IP address
- `port` (Number) The SSH server port
- `username` (String) The SSH username

Optional:

- `allow_insecure_algos` (Boolean) Allow insecure SSH algorithms
- `password_auth` (Block List, Max: 1) Password authentication for SSH (see [below for nested schema](#nestedblock--ssh_options--password_auth))
- `public_key_auth` (Block List, Max: 1) Public key authentication for SSH (see [below for nested schema](#nestedblock--ssh_options--public_key_auth))

<a id="nestedblock--ssh_options--password_auth"></a>
### Nested Schema for `ssh_options.password_auth`

Required:

- `password` (String, Sensitive) The password for SSH authentication


<a id="nestedblock--ssh_options--public_key_auth"></a>
### Nested Schema for `ssh_options.public_key_auth`
