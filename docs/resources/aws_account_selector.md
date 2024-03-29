---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_aws_account_selector Resource - commonfate"
subcategory: ""
description: |-
  A Selector to match AWS Accounts with a criteria based on the 'when' field.
---

# commonfate_aws_account_selector (Resource)

A Selector to match AWS Accounts with a criteria based on the 'when' field.

## Example Usage

```terraform
resource "commonfate_aws_account_selector" "select_all" {
  id                  = "select_all_aws"
  name                = "Select All AWS"
  aws_organization_id = "o-123456789a"
  when                = "true"
}

resource "commonfate_aws_account_selector" "org_unit" {
  id                  = "select_all_aws"
  name                = "Select All AWS"
  aws_organization_id = "o-123456789a"
  when                = <<EOF
  resource in AWS::OrgUnit::"ou-abcd-12345667"
  EOF
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_organization_id` (String) The AWS organization ID
- `id` (String) The ID of the selector
- `when` (String) A Cedar expression with the criteria to match accounts on, e.g: `resource.tag_keys contains "production" && resource in AWS::OrgUnit::"example"`

### Optional

- `name` (String) The unique name of the selector. Call this something memorable and relevant to the resources being selected. For example: `prod-data-eng`


