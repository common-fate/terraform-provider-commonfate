---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_okta_integration Resource - commonfate"
subcategory: ""
description: |-
  Registers a Okta integration
---

# commonfate_okta_integration (Resource)

Registers a Okta integration

## Example Usage

```terraform
resource "commonfate_okta_integration" "demo" {
  name                = "Example"
  organization_id     = "dev-12345678"
  api_key_secret_path = "/common-fate/prod/okta-api-key"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_key_secret_path` (String) Path to secret for the Okta API Key
- `name` (String) The name of the integration: use a short label which is descriptive of the organization you're connecting to
- `organization_id` (String) The Okta Organization ID

### Read-Only

- `id` (String) The internal Common Fate ID

