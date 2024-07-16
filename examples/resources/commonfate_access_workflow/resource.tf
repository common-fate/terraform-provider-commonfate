resource "commonfate-access_workflow" "workflow-demo" {
  name                     = "demo"
  access_duration_seconds  = 60 * 60
  priority                 = "100"
  default_duration_seconds = 30 * 60
}
