terraform {
  required_providers {
    kinsta = {
      source = "blavity.com/platform/kinsta"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.5"
    }
  }
}

# Get Kinsta credentials from Vault
data "vault_kv_secret_v2" "kinsta_api" {
  mount = "platform"
  name  = "kinsta/prod/api-credentials"
}

provider "kinsta" {
  api_key    = data.vault_kv_secret_v2.kinsta_api.data["api_key"]
  company_id = data.vault_kv_secret_v2.kinsta_api.data["company_id"]
}

provider "vault" {
  address = "https://vault.blavity.com"
}

resource "kinsta_wordpress_site" "staging" {
  display_name   = "Blavityinc Staging"
  region         = "us-central1"
  install_mode   = "new"
  admin_email    = "platform@blavity.com"
  admin_password = "temporary-password-change-me"
  admin_user     = "blavityinc_admin"
  site_title     = "Blavityinc Staging Site"
  wp_language    = "en_US"
}

output "site_id" {
  value = kinsta_wordpress_site.staging.site_id
}

output "environment_id" {
  value = kinsta_wordpress_site.staging.environment_id
}
