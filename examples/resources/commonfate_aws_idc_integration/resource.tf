resource "commonfate_aws_idc_integration" "demo" {
  identity_store_id = "d-12345678"
  name              = "Demo"
  reader_role_arn   = "arn:aws:iam::12345678912:role/common-fate-prod-idc-reader-role"
  sso_instance_arn  = "abcd-1234"
  sso_region        = "ap-southeast-2"
}
