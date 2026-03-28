---
page_title: "tribal_admin_settings (Resource) - tribal"
description: |-
  Manages organization-wide notification settings in Tribal.
---

# tribal_admin_settings

Manages organization-wide notification and alert settings in Tribal. This is a singleton resource — only one instance should exist per Terraform workspace.

-> **Note:** `terraform destroy` on `tribal_admin_settings` is a no-op. Admin settings cannot be deleted, only updated.

## Example Usage

```terraform
resource "tribal_admin_settings" "org" {
  reminder_days            = [60, 30, 14, 7, 1]
  notify_hour              = 9
  alert_on_overdue         = true
  alert_on_delete          = true
  alert_on_review_overdue  = true
  review_cadence_months    = 12

  # Optional
  slack_webhook = "https://hooks.slack.com/services/YOUR/ORG/WEBHOOK"
}
```

## Argument Reference

### Required

- `reminder_days` - List of days before expiration at which to send reminders (e.g., `[30, 14, 7]`).
- `notify_hour` - UTC hour (0–23) at which daily reminders are sent.
- `alert_on_overdue` - Whether to send alerts for already-expired resources.
- `alert_on_delete` - Whether to send an admin Slack alert when a resource is deleted.
- `alert_on_review_overdue` - Whether to send alerts when resource reviews are overdue.

### Optional

- `slack_webhook` - Organization-wide Slack webhook URL for expiration notifications.
- `review_cadence_months` - How often resources should be reviewed, in months. Valid values: `6`, `12`, or `24`. Omit or set to `null` to disable periodic reviews.
