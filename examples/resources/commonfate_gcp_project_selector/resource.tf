
resource "commonfate_gcp_project_selector" "example" {
  name                = "gcp-prod"
  gcp_organization_id = "organization/29034834894"
  when                = <<EOF
  resource.tag_keys contains "production" && resource in GCP::Folder::"folders/342982723"
  EOF
}
