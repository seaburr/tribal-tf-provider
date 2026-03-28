terraform {
  required_providers {
    tribal = {
      source = "registry.terraform.io/seaburr/tribal"
    }
  }
}

provider "tribal" {
  host    = "http://localhost:8000"
  api_key = "tribal_sk_..."  # or use TRIBAL_API_KEY env var
}

# Manage organization-wide admin settings (singleton)
resource "tribal_admin_settings" "org" {
  reminder_days           = [60, 30, 14, 7, 1]
  notify_hour             = 9  # 9 AM UTC
  slack_webhook           = "https://hooks.slack.com/services/YOUR/ORG/WEBHOOK"
  alert_on_overdue        = true
  alert_on_delete         = true
  alert_on_review_overdue = true
  review_cadence_months   = 12  # Review resources annually
}

# Manage a team
resource "tribal_team" "platform" {
  name = "Platform Team"
}

# Manage a tracked API key resource, assigned to a team
resource "tribal_resource" "prod_api_key" {
  name                    = "Production API Key"
  dri                     = "platform-team@example.com"
  type                    = "API Key"
  expiration_date         = "2026-12-31"
  purpose                 = "Used by the payment service to authenticate with Stripe"
  generation_instructions = "Log into Stripe dashboard > Developers > API keys > Create new key"
  secret_manager_link     = "https://vault.example.com/secret/stripe-api-key"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"
  team_id                 = tribal_team.platform.id
}

# Manage an SSH deploy key
resource "tribal_resource" "ssh_deploy_key" {
  name                    = "GitHub Deploy Key - myapp"
  dri                     = "devops@example.com"
  type                    = "SSH Key"
  expiration_date         = "2027-06-30"
  purpose                 = "Allows CI/CD to pull from the myapp GitHub repository"
  generation_instructions = "Run: ssh-keygen -t ed25519 -C 'deploy@ci' then add public key to GitHub repo settings"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"
}

# Manage a non-expiring resource
resource "tribal_resource" "service_account" {
  name                    = "Internal Service Account Key"
  dri                     = "platform-team@example.com"
  type                    = "API Key"
  does_not_expire         = true
  purpose                 = "Long-lived service account for internal tooling"
  generation_instructions = "Contact IT to rotate via the admin console"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"
}

# Manage a TLS certificate with automatic expiry refresh via endpoint polling
resource "tribal_resource" "tls_cert" {
  name                    = "api.example.com TLS Certificate"
  dri                     = "infra@example.com"
  type                    = "Certificate"
  expiration_date         = "2026-09-15"
  purpose                 = "TLS certificate for api.example.com"
  generation_instructions = "Run certbot renew or use ACM auto-renewal"
  secret_manager_link     = "arn:aws:acm:us-east-1:123456789012:certificate/abc123"
  slack_webhook           = "https://hooks.slack.com/services/TEAM/CHANNEL/WEBHOOK"
  certificate_url         = "https://api.example.com"
  auto_refresh_expiry     = true
}
