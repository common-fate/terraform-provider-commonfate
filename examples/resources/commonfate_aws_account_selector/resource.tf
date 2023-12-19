
resource "commonfate_aws_account_selector" "select_all" {
  id                  = "select_all_aws"
  name                = "Select All AWS"
  aws_organization_id = "o-123456789a"
  when                = "true"
}

resource "commonfate_aws_account_selector" "org_unit" {
  id                  = "select_all_aws"
  name                = "Select All AWS"
  aws_organization_id = "o-123456789a"
  when                = <<EOF
  resource in AWS::OrgUnit::"ou-abcd-12345667"
  EOF
}
