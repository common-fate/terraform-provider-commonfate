resource "commonfate_opsgenie_integration" "opsgenie" {
  name                = "Demo"
  api_key_secret_path = "/common-fate/prod/opsgenie-api-key"
}
