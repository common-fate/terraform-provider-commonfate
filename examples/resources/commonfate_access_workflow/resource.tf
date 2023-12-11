resource "commonfate-access_workflow" "workflow-demo" {
  
  name="demo"
  access_duration="2h"
  try_extend_after="5m"
  priority="100"
}
