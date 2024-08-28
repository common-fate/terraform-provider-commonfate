resource "commonfate-access_workflow" "workflow-demo" {
  name                     = "demo"
  access_duration_seconds  = 60 * 60 * 2
  default_duration_seconds = 50 * 60

  priority = 100

  validation = {
    has_reason = true
  }

  activation_expiry= 60 * 5

  extension_conditions = {
    extension_duration_seconds   = 60 * 30
    maximum_number_of_extensions = 4
  }

}
