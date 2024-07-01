resource "commonfate_aws_idc_account_availabilities" "demo" {
  workflow_id             = "workflow_id"
  aws_permission_set_arn  = "arn:aws:sso:::permissionSet/ssoins-12345667879812/ps-12345678912"
  aws_account_selector_id = "selector_id"
  aws_identity_store_id   = "d-12345678"
  role_priority           = 100
}
