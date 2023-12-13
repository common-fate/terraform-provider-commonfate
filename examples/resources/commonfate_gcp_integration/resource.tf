resource "commonfate_gcp_integration" "demo" {
  name                     = "demo"
  workload_identity_config = jsonencode({}) # include your Workload Identity Federation config here
}
