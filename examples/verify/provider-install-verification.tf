terraform {
  required_providers {
    commonfate = {
      source = "registry.terraform.io/example/commonfate"
      version = "~> 1.0.0"

    }
  }
}

provider "commonfate" {}

resource "commonfate_access_rule" "test" {
  name ="test"
  description="test"
  approval={
    users=[
      "jack@commonfate.io"
    ]
  }
  target={
    field="accountId"
    value=["12345678912"]
  }
  target_provider_id="aws-sso"
  duration="12h"
}