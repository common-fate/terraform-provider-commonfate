---
"@common-fate/terraform-provider-commonfate": patch
---

Fixes an issue where terraform plan would always show a change for commonfate_slack_alert when the send_direct_message_to_approvers field is true
