resource "commonfate_datastax_organization_selector" "demo" {
  datastax_organization_id = "12345-12345-12345-12345"
}

resource "commonfate_datastax_organization_availabilities" "demo" {
  workflow_id                       = "workflow_id"
  datastax_organization_selector_id = commonfate_datastax_organization_selector.demo.id
  datastax_organization_id          = "12345-12345-12345-12345"
  role_id                           = "abcdef-abcdef-abcdef-abcdef-abcdef"
}
