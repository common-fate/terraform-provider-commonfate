
terraform {
  required_providers {
    commonfate = {
      source = "commonfate.com/commonfate/commonfate"
      version = "1.0.0"

    }
  }
}

provider "commonfate" {
  host = "https://lpscp2am1g.execute-api.ap-southeast-2.amazonaws.com/prod/"
}


module "sso_permission-sets" {
  source  = "cloudposse/sso/aws//modules/permission-sets"
  version = "0.7.1"
}

module "permission_sets" {
  source = "github.com/cloudposse/terraform-aws-sso.git//modules/permission-sets?ref=master"

  permission_sets = [
    {
      name                                = "S3ListBuckets",
      description                         = "Allow List Buckets in account",
      relay_state                         = "",
      session_duration                    = "",
      tags                                = {},
      inline_policy                       = data.aws_iam_policy_document.S3Access.json,
      policy_attachments                  = []
      customer_managed_policy_attachments = []
    }
  ]
}

data "aws_iam_policy_document" "S3Access" {
  statement {
    sid = "1"

    actions = ["s3:ListBucket"]

    resources = [
      "arn:aws:s3:::*",
    ]
  }
}

resource "aws_iam_policy" "S3Access" {
  name   = "S3Access"
  path   = "/"
  policy = data.aws_iam_policy_document.S3Access.json
}


resource "commonfate_access_rule" "s3-example" {
  name ="s3ListBuckets"
  description="Allows users to view buckets in AWS"
  groups=["common_fate_administrators"]
 
  
  target=[
    {
      field="accountId"
      value=["616777145260"]
    },
    {
      field="permissionSetArn"
      value=[module.permission_sets.permission_sets["S3ListBuckets"].arn]
    }
  ]
  target_provider_id="aws-sso-v2"
  duration="3600"

  depends_on = [
    module.permission_sets.permission_sets
  ]
}
