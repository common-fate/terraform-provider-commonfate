---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_okta_group_availabilities Resource - commonfate"
subcategory: ""
description: |-
  A specifier to make Okta Groups available for selection under a particular Access Workflow
---

# commonfate_okta_group_availabilities (Resource)

A specifier to make Okta Groups available for selection under a particular Access Workflow

## Example Usage

```terraform
resource "commonfate_okta_group_availabilities" "demo" {
  workflow_id            = "workflow_id"
  okta_group_selector_id = "selector_id"
  organization_id        = "dev-12345678"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `okta_group_selector_id` (String) The target to make available. Should be a Selector entity.
- `organization_id` (String) The Okta Organization ID
- `workflow_id` (String) The Access Workflow ID

### Read-Only

- `id` (String) The internal Common Fate ID

