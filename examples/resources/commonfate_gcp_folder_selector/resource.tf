
resource "commonfate_gcp_folder_selector" "example" {
  name                = "gcp-prod"
  gcp_organization_id = "organization/29034834894"
  when                = <<EOF
  resource == GCP::Folder::"folders/342982723"
  EOF
}
