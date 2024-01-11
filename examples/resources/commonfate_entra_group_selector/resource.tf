
resource "commonfate_entra_group_selector" "select_all" {
  id        = "select_all_entra_groups"
  name      = "Select All Entra Groups"
  tenant_id = "abcd-1234-abcd-1234"
  when      = "true"
}

resource "commonfate_entra_group_selector" "name_contains" {
  id        = "production_entra_groups"
  name      = "Select Production Entra Groups"
  tenant_id = "abcd-1234-abcd-1234"
  when      = <<EOF
  resource.name contains "production"
  EOF
}
