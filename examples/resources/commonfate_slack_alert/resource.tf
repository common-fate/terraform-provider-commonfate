resource "commonfate-slack_alert" "demo" {
  id                                  = "demo"
  workflow_id                         = "wrk_123"
  slack_channel_id                    = "demo"
  slack_workspace_id                  = "123"
  use_web_console_for_request_actions = false
}
