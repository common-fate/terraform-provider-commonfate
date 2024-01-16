resource "commonfate_aws_rds_integration" "demo" {
  name           = "Demo"
  read_role_arns = ["arn:aws:iam::12345678912:role/common-fate-prod-aws-read-role"]
  regions        = ["ap-southeast-2"]
}
