terraform {
  required_providers {
    kinsta = {
      source  = "blavity/kinsta"
      version = "~> 0.1"
    }
  }
}

provider "kinsta" {
  # api_key and company_id can be set here or via environment variables:
  #   KINSTA_API_KEY
  #   KINSTA_COMPANY_ID
}
