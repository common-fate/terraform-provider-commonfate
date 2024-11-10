
resource "commonfate_okta_group_selector" "select_all" {
  id              = "select_all_okta_groups"
  name            = "Select All Okta Groups"
  organization_id = "dev-12345678"
  when            = "true"
}

resource "commonfate_okta_group_selector" "name_contains" {
  id              = "production_okta_groups"
  name            = "Select Production Okta Groups"
  organization_id = "dev-12345678"
  when            = <<EOF
  resource.name like "*production*"
  EOF
}
