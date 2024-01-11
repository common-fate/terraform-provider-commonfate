
resource "commonfate_entra_group_selector" "select_all" {
  id        = "select_all_entra_groups"
  name      = "Select All Entra Groups"
  tenant_id = "00551d20-529d-478f-b1cb-fb04a2653e97"
  when      = "true"
}

resource "commonfate_entra_group_selector" "name_contains" {
  id        = "production_entra_groups"
  name      = "Select Production Entra Groups"
  tenant_id = "00551d20-529d-478f-b1cb-fb04a2653e97"
  when      = <<EOF
  resource.name contains "production"
  EOF
}
