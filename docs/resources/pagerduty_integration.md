---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_pagerduty_integration Resource - commonfate"
subcategory: ""
description: |-
  Registers a PagerDuty integration
---

# commonfate_pagerduty_integration (Resource)

Registers a PagerDuty integration



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_id` (String) The PagerDuty application Client ID
- `client_secret_secret_path` (String) Path to secret for Client Secret
- `name` (String) The name of the integration: use a short label which is descriptive of the organization you're connecting to

### Read-Only

- `id` (String) The internal Common Fate ID


