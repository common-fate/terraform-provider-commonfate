resource "commonfate_aws_rds_selector" "select_all" {
  id                  = "select_all_aws_rdss"
  name                = "Select All Databases"
  aws_organization_id = "o-123456789a"
  when                = "true"
}
