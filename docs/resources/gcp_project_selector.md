---
page_title: "commonfate_gcp_project_selector Resource - commonfate"
subcategory: ""
description: |-
  A Selector to match GCP projects with a criteria based on the 'when' field.
---

# commonfate_gcp_project_selector (Resource)

A Selector to match GCP projects with a criteria based on the 'when' field.



## Example Usage

```terraform
resource "commonfate_gcp_project_selector" "example" {
  name                = "gcp-prod"
  gcp_organization_id = "organization/29034834894"
  when                = <<EOF
  resource.tag_keys contains "production" && resource in GCP::Folder::"folders/342982723"
  EOF
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `gcp_organization_id` (String) The GCP organization ID
- `id` (String) The ID of the selector
- `when` (String) A Cedar expression with the criteria to match projects on, e.g: `resource.tag_keys contains "production" && resource in GCP::Folder::"folders/342982723"`

### Optional

- `name` (String) The unique name of the selector. Call this something memorable and relevant to the resources being selected. For example: `prod-data-eng`

