resource "commonfate-access_workflow" "workflow-demo" {
  name                     = "demo"
  access_duration_seconds  = 60 * 60
  try_extend_after_seconds = 10 * 60
  priority                 = "100"
}
