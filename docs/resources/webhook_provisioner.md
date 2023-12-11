---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commonfate_webhook_provisioner Resource - commonfate"
subcategory: ""
description: |-
  Registers a provisioner with a webhook URL to provision access.
---

# commonfate_webhook_provisioner (Resource)

Registers a provisioner with a webhook URL to provision access.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `capabilities` (Attributes List) The resources and integrations that this provisioner supports. (see [below for nested schema](#nestedatt--capabilities))

### Optional

- `url` (String) The webhook URL.

### Read-Only

- `id` (String) The resource ID

<a id="nestedatt--capabilities"></a>
### Nested Schema for `capabilities`

Required:

- `belonging_to` (Attributes) (see [below for nested schema](#nestedatt--capabilities--belonging_to))
- `role_type` (String) The type of target such as `GCP::Project` or `AWS::Account`
- `target_type` (String) The type of target such as `GCP::Project` or `AWS::Account`

<a id="nestedatt--capabilities--belonging_to"></a>
### Nested Schema for `capabilities.belonging_to`

Required:

- `id` (String) The entity ID
- `type` (String) The entity type

