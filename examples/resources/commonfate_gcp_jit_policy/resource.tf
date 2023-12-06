resource "commonfate_gcp_jit_policy" "jit-policy-1" {
  name="demo"
  priority=100
  duration="2h"
  notify_slack_channel="R023FDHJ34"
  role="roles/editor"
  match_project_ids=[
    "sandbox-1"
  ]
  match_project_folders= [
    "4443422451"
  ]

}