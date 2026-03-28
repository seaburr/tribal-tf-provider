---
page_title: "tribal_resource (Resource) - tribal"
description: |-
  Manages a tracked credential or certificate in Tribal.
---

# tribal_resource

Manages a tracked credential or certificate in Tribal. Resources represent expiring secrets such as API keys, SSH keys, TLS certificates, or other credentials that need renewal reminders.

## Example Usage

```terraform
resource "tribal_resource" "example" {
  name                    = "Production API Key"
  dri                     = "platform-team@example.com"
  type                    = "API Key"
  expiration_date         = "2026-12-31"
  purpose                 = "Authenticates the payment service with Stripe"
  generation_instructions = "Log into Stripe dashboard > Developers > API keys > Create new key"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"

  # Optional
  secret_manager_link = "https://vault.example.com/secret/stripe-api-key"
  team_id             = tribal_team.platform.id
}
```

### Non-Expiring Resource

```terraform
resource "tribal_resource" "long_lived_key" {
  name                    = "Internal Service Account Key"
  dri                     = "platform-team@example.com"
  type                    = "API Key"
  does_not_expire         = true
  purpose                 = "Long-lived service account for internal tooling"
  generation_instructions = "Contact IT to rotate via the admin console"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"
}
```

### TLS Certificate with Auto-Refresh

```terraform
resource "tribal_resource" "tls_cert" {
  name                    = "api.example.com TLS Certificate"
  dri                     = "infra@example.com"
  type                    = "Certificate"
  expiration_date         = "2026-09-15"
  purpose                 = "TLS certificate for api.example.com"
  generation_instructions = "Run certbot renew or use ACM auto-renewal"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"

  certificate_url     = "https://api.example.com"
  auto_refresh_expiry = true
}
```

## Argument Reference

### Required

- `name` - Display name of the resource.
- `dri` - Directly Responsible Individual (email or team name) for this resource.
- `type` - Type of resource. One of: `Certificate`, `API Key`, `SSH Key`, `Other`.
- `purpose` - What this credential is used for.
- `generation_instructions` - Steps to renew or regenerate this credential.
- `slack_webhook` - Slack webhook URL for expiration notifications.

### Optional

- `expiration_date` - Expiration date in `YYYY-MM-DD` format. Required unless `does_not_expire` is `true`.
- `does_not_expire` - When `true`, the resource does not expire and `expiration_date` is cleared. Defaults to `false`.
- `secret_manager_link` - URL or ARN pointing to this secret in a secret manager.
- `team_id` - ID of the team this resource belongs to. Auto-assigned to the default team if omitted.
- `certificate_url` - TLS endpoint URL to poll for automatic certificate expiry detection.
- `auto_refresh_expiry` - When `true`, `expiration_date` is automatically updated by polling `certificate_url`. Defaults to `false`.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Numeric resource ID.
- `public_key_pem` - PEM-encoded public certificate, if uploaded.
- `last_reviewed_at` - Timestamp when the resource was last reviewed (managed via the Tribal UI/API).
- `created_at` - Timestamp when the resource was created.
- `updated_at` - Timestamp when the resource was last updated.

## Import

`tribal_resource` resources can be imported using the numeric resource ID:

```shell
terraform import tribal_resource.example 42
```
