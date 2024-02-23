resource "commonfate_okta_integration" "demo" {
  name                = "Example"
  organization_id     = "dev-12345678"
  api_key_secret_path = "/common-fate/prod/okta-api-key"
}
