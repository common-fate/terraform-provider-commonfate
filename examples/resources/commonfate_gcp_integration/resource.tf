resource "commonfate_gcp_integration" "demo" {
  
  name="demo"
  workload_identity_config=<<EOH
  {}
  EOH
}