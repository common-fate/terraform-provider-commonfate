# @common-fate/terraform-provider-commonfate

## 2.14.0

### Minor Changes

- c4a3323: Add support for configuring the Terraform provider entirely via environment variables.
- 4a3a3f0: Adds `gcp_bigquery_table_selector` and `gcp_bigquery_table_availabilities`, used for just-in-time access to BigQuery Tables.
- 4a3a3f0: Adds `gcp_bigquery_dataset_selector` and `gcp_bigquery_dataset_availabilities`, used for just-in-time access to BigQuery Datasets.
- 4a3a3f0: Adds `gcp_organization_selector` and `gcp_organization_availabilities`, used for just-in-time access to organization-level GCP roles.
- 794a2d8: Add Webhook Integration resource

### Patch Changes

- 950bf6d: Fix an issue where Terraform would prompt to set the 'use_web_console_for_approval_action' to null each plan/apply.

## 2.13.1

### Patch Changes

- f9e3d72: Add additional config for Slack Alert to optionally perform approvals via the web app.

## 2.13.0

### Minor Changes

- 956ea4a: add ability to link slack notifier to a slack integration via its ID

## 2.12.0

### Minor Changes

- 8930396: Added DataStax integration resources.

## 2.11.0

### Minor Changes

- 15361af: Added DataStax integration

## 2.10.1

### Patch Changes

- 8403479: Fix availability_spec resource update api call causing 500 errors

## 2.10.0

### Minor Changes

- 3fbaac4: Add resources for Okta JIT integration

## 2.9.0

### Minor Changes

- 8a8ef14: Remove the RDS integration and add support for audit role to the aws IDC integration.

## 2.8.1

### Patch Changes

- c123aec: Update documentation

## 2.8.0

### Minor Changes

- 742fc86: add AWS IAM Identity Center group resources
