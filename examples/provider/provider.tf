
terraform {
  required_providers {
    commonfate = {
      source  = "commonfate.com/commonfate/commonfate"
      version = "~> 1.0"
    }
  }
}

provider "commonfate" {
  governance_api_url = "https://commonfate-api.example.com"
}
