terraform {
  required_providers {
    commonfate = {
      source = "commonfate.com/commonfate/commonfate"
      version = "1.0.0"

    }
  }
}

provider "commonfate" {
  host = "abc"
  password = "abc"
  username = "abc"
}

# resource "commonfate_access_rule" "test" {
#   name ="demo-test"
#   description="test"
#   groups=["internal group"]
#   # approval={
#   #   users=[
#   #     "jack@commonfate.io"
#   #   ]
#   # }
#   # target={
#   #   field="accountId"
#   #   value=["12345678912"]
#   # }
#   target_provider_id="aws-sso"
#   duration="3600"
# }

resource "commonfate_access_rule" "test-2" {
  name ="This was made in terraform 2"
  description="Access rule made in terraform"
  groups=["common_fate_administrators"]
  approval= {
      users=["jack@commonfate.io"]
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