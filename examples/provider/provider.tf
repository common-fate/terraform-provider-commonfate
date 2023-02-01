
terraform {
  required_providers {
    commonfate = {
      source  = "commonfate.com/commonfate/commonfate"
      version = "~> 1.0"
    }
  }
}

provider "commonfate" {
  host = "https://commonfate-api.example.com"
}
