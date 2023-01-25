
terraform {
  required_providers {
    commonfate = {
      source = "common-fate/commonfate/commonfate"
      version = "1.0.1"
    }
  }
}



provider "commonfate" {
  host = ""
}


resource "commonfate_access_rule" "aws-admin" {
  name ="This was made in terraform demo"
  description="Access rule made in terraform"
  groups=["common_fate_administrators"]
  approval= {
      users=[""]
  }
  
  target=[
    {
      field="accountId"
      value=["123456789012"]
    },
    {
      field="permissionSetArn"
      value=[""]
    }
  ]
  target_provider_id="aws-sso-v2"
  duration="3600"
}