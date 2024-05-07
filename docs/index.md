---
page_title: "Provider: Common Fate"
description: |-
  The Common Fate provider is used to configure and manage access to your cloud.
---


The Common Fate provider is used to configure and customize your Common Fate deployment.
Manage things like identitity, creating access policies and just-in-time (JIT) access workflows for permission to resources ranging across all cloud providers.

## Configuration

To get your provider set up you will need some essential variables, these are:

- Deployment API URL
- OIDC Client ID
- OIDC Client Secret
- OIDC Issuer

All of these can be sourced from your Common Fate's deployment Terraform outputs.
For more information on how to find these variables checkout our official documentation [here](https://enterprise.docs.commonfate.io/deploy)

## Example Usage

```terraform
terraform {
  required_providers {
    commonfate = {
      source = "common-fate/commonfate"
      version = "2.1.1"
    }
  }
}


provider "commonfate" {
  api_url            = "http://commonfate.example.com"
  oidc_client_id     = "349dfdfkljwerpoizxckf3fds345xcvv"
  oidc_issuer        = "https://cognito-idp.ap-southeast-2.amazonaws.com/ap-southeast-2_jieDxjtS"
}


resource "commonfate_gcp_integration" "demo" {
  name="demo"
  reader_workload_identiy_config=""
  reader_service_account_credentials_secret_path=""
  organization_id=""
  google_workspace_customer_id=""
}

resource "commonfate_access_workflow" "demo" {
  name="demo"
  access_duration_seconds="7200"
  try_extend_after_seconds="3500"
  priority="100"
  default_duration_seconds="3600"
}

resource "commonfate_gcp_project_selector" "demo" {
  name="demo"
  gcp_organization_id="organization/29034834894"
  when = <<EOF
  resource.tag_keys contains "production" && resource in GCP::Folder::"folders/342982723"
  EOF

}

resource "commonfate_gcp_project_availabilities" "demo" {
  workflow_id=var.commonfate_access_workflow.demo.id
  role = "roles/owner"
  gcp_project_selector_id = var.commonfate_gcp_project_selector.demo.id
  google_workspace_customer_id = "34dFHJ3H4H"
}

resource "commonfate_policyset" "demo" {
  id = "demo"
  text = <<EOH
    permit(
      principal,
      action == Access::Action::"Request",
      resource
    );
    EOH
}

resource "commonfate_pagerduty_integration" "demo" {
  name="demo"
  client_id=""
  client_secret_secret_path=""
}

resource "commonfate_slack_integration" "demo" {
  name="demo"
  client_id=""
  client_secret_secret_path=""
  signing_secret_secret_path=""
}

resource "commonfate_slack_alert" "demo" {
  workflow_id=var.commonfate_access_workflow.demo.id
  slack_channel_id="C2044RFJMWMS"
}
```

