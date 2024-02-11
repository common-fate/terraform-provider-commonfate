---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_aws_idc_account_availabilities Resource - commonfate"
subcategory: ""
description: |-
  A specifier to make AWS accounts available for selection under a particular Access Workflow
---

# commonfate_aws_idc_account_availabilities (Resource)

A specifier to make AWS accounts available for selection under a particular Access Workflow

## Example Usage

```terraform
resource "commonfate_aws_idc_account_availabilities" "demo" {
  workflow_id             = "workflow_id"
  aws_permission_set_arn  = "arn:aws:sso:::permissionSet/ssoins-12345667879812/ps-12345678912"
  aws_account_selector_id = "selector_id"
  aws_identity_store_id   = "d-12345678"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_account_selector_id` (String) The target to make available. Should be a Selector entity.
- `aws_identity_store_id` (String) The IAM Identity Center identity store ID
- `aws_permission_set_arn` (String) The AWS Permission Set to make available
- `workflow_id` (String) The Access Workflow ID

### Read-Only

- `id` (String) The internal Common Fate ID

