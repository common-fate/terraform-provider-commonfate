# @common-fate/terraform-provider-commonfate

## 2.18.1

### Patch Changes

- 1041e78: Add sso_access_portal_url field to aws_idc_integration field

## 2.18.0

### Minor Changes

- eb15d50: Adds support for configuring Amazon S3 log destinations using the Terraform provider.

### Patch Changes

- 7555944: Update SDK for policy API Client
- bc57b5c: Support renaming commonfate_policyset resources
- 24450ff: Fixes an issue where terraform plan would always show a change for commonfate_slack_alert when the send_direct_message_to_approvers field is true
- 1de9dd6: Fixes commonfate_datastax_organization_selector always wanting to update if name is not set

## 2.17.0

### Minor Changes

- eaa448a: Adds support for configuring the Common Fate Auth0 integration.

### Patch Changes

- 79e779e: Fix typo in access selector documentation.
- 2cd50fb: Adds validation options to the commonfate-access-workflow resource. You can now configure workflows to require a reason to be provided.
- 7adc52a: Add default duration to access workflow.

## 2.16.1

### Patch Changes

- 7a104b7: Fix an issue causing the provider to panic.

## 2.16.0

### Minor Changes

- 3bf9ab9: Added custom resource for GCP Role Group

### Patch Changes

- de45f99: Add default duration to access workflow.
- 74b047e: Adds option to slack alert to send direct messages to approvers

## 2.15.0

### Minor Changes

- cbd0f85: adds variable to workflows to configure expiry time for closing approved requests

## 2.14.1

### Patch Changes

- b3d6971: Fix for config file does not exist error

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
