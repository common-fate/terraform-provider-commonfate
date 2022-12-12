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

resource "commonfate_access_rule" "test" {
  name ="demo-test"
  # description="test"
  # approval={
  #   users=[
  #     "jack@commonfate.io"
  #   ]
  # }
  # target={
  #   field="accountId"
  #   value=["12345678912"]
  # }
  # target_provider_id="aws-sso"
  # duration="12h"
}