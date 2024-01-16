resource "commonfate_aws_rds_availabilities" "demo" {
  workflow_id           = "workflow_id"
  aws_rds_selector_id   = "selector_id"
  aws_identity_store_id = "d-12345678"
}
