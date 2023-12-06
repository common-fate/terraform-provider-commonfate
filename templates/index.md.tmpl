---
page_title: "Provider: Common Fate"
description: |-
  The Common Fate provider is used to configure and manage access to your cloud.
---


The Common Fate provider is used to configure and customize your Common Fate deployment. 
Manage things like identitity, creating access policies and just-in-time (JIT) access workflows for permission to resources ranging across all cloud providers.

## Example Usage

```terraform
terraform {
  required_providers {
    commonfate = {
      source = "common-fate/commonfate"
      version = "2.0.0"
    }
  }
}

provider "commonfate" {
  issuer_url =         "https://acme-web-app-client.auth.ap-southeast-2.amazoncognito.com"
  deployment_api_url = "http://acme.dev.io"
  oidc_client_id =     var.oidc_client_id
  oidc_client_secret = var.oidc_client_secret
  oidc_issuer =        "https://cognito-idp.ap-southeast-2.amazonaws.com/ap-southeast-2_jieDxjtS"
}

data "commonfate_pagerduty_schedule" "on-call-billing" {
  name = "on-call-billing"
}


resource "commonfate_schedule" "on-call-billing-schedule" {
  id = data.commonfate_pagerduty_schedule.on-call-billing.id
  name = data.commonfate_pagerduty_schedule.on-call-billing.name
}

resource "commonfate_approval_workflow" "default" {
  notify_slack_channel="C08R74G352"
  name="default"
  default=true
}

resource "commonfate_gcp_jit_policy" "demo" {
  name="demo"
  priority=100
  duration="2h"
  notify_slack_channel="C08R74G352"
  role="roles/editor"
  match_project_ids=[
    "sandbox-1"
  ]

  match_project_folders= [
    "2138582926"
  ]

}

```

## Configuration

To get your provider set up you will need some essential variables, these are:
- Issuer URL
- Deployment API URL
- OIDC Client ID
- OIDC Client Secret
- OIDC Issuer

All of these can be sourced from your Common Fate's deployment Terraform outputs.

