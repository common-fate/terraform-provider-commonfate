---
page_title: "commonfate_gcp_integration Resource - commonfate"
subcategory: ""
description: |-
  Registers an integration with Google Cloud
---

# commonfate_gcp_integration (Resource)

Registers an integration with Google Cloud



## Example Usage

```terraform
resource "commonfate_gcp_integration" "demo" {
  name                     = "demo"
  workload_identity_config = jsonencode({}) # include your Workload Identity Federation config here
}
```


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `google_workspace_customer_id` (String) The Google Workspace Customer ID
- `name` (String) The name of the integration: use a short label which is descriptive of the organization you're connecting to
- `organization_id` (String) GCP organization ID

### Optional

- `provisioner_service_account_credentials_secret_path` (String) Path to secret for Service account credentials
- `provisioner_workload_identity_config` (String) GCP Workload Identity Config as a JSON string. If you don't know where to find this, check out our documentation [here](https://enterprise.docs.commonfate.io/deploy)
- `reader_service_account_credentials_secret_path` (String) Path to secret for Service account credentials
- `reader_workload_identity_config` (String) GCP Workload Identity Config as a JSON string. If you don't know where to find this, check out our documentation [here](https://enterprise.docs.commonfate.io/deploy)

### Read-Only

- `id` (String) The internal Common Fate ID

