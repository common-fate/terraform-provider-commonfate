terraform {
  required_providers {
    commonfate = {
      source = "commonfate.com/commonfate/commonfate"
      version = "1.0.0"

    }
  }
}

provider "commonfate" {
  host = "http://localhost:8889"
  
}


resource "commonfate_access_rule" "sandbox-sso-admin" {
  name ="This was made in terraform 11"
  description="Access rule made in terraform"
  groups=["common_fate_administrators"]
  approval= {
      users=["jack@commonfate.io", "jack+1@commonfate.io"]
  }
  
  target=[
    {
      field="accountId"
      value=["632700053629"]
    },
    {
      field="permissionSetArn"
      value=["arn:aws:sso:::permissionSet/ssoins-825968feece9a0b6/ps-dda57372ebbfeb94"]
    }
  ]
  target_provider_id="aws-sso-v2"
  duration="3600"
}

resource "commonfate_access_rule" "azure-developer-group" {
  name ="azure groups rule new name"
  description="Access rule made in terraform for adding users to a group in azure"
  groups=["common_fate_administrators"]
  approval= {
      users=["jack@commonfate.io"]
  }
  
  target=[
    {
      field="groupId"
      value=["4e552c12-d8b3-4fec-a3bf-cfc741f6f02b"]
    },
    
  ]
  target_provider_id="azure-ad"
  duration="3600"
}