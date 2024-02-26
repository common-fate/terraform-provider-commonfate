resource "commonfate_entra_integration" "demo" {
  name                      = "Example"
  tenant_id                 = "00551d20-529d-478f-b1cb-fb04a2653e97"
  client_id                 = "a0c612e6-2ef6-4467-9ce5-2c14f2c31adc"
  client_secret_secret_path = "/common-fate/prod/entra-client-secret"
}
