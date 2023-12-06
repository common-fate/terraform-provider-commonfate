
terraform {
  required_providers {
    commonfate = {
      source  = "common-fate/commonfate"
      version = "2.0.0"
    }
  }
}

provider "commonfate" {
  issuer_url =         "https://acme-web-app-client.auth.ap-southeast-2.amazoncognito.com"
  deployment_api_url = "http://acme.dev.io"
  oidc_client_id =     var.oidc_client_id
  oidc_client_secret = var.oidc_client_secret
  oidc_issuer =        "https://cognito-idp.ap-southeast-2.amazonaws.com/ap-southeast-2_jieDxjtS"
}