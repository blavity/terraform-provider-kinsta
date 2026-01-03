package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_ResourceWordPressEnvironment_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressEnvironmentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressEnvironmentExists("kinsta_wordpress_environment.staging"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.staging", "display_name", "Staging Environment"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.staging", "is_premium", "false"),
					resource.TestCheckResourceAttrSet("kinsta_wordpress_environment.staging", "site_id"),
					resource.TestCheckResourceAttrSet("kinsta_wordpress_environment.staging", "id"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressEnvironment_Premium(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressEnvironmentConfigPremium,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressEnvironmentExists("kinsta_wordpress_environment.premium_staging"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.premium_staging", "is_premium", "true"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.premium_staging", "admin_email", "premium@example.com"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressEnvironment_CustomSettings(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressEnvironmentConfigCustom,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressEnvironmentExists("kinsta_wordpress_environment.custom"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.custom", "php_version", "8.2"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.custom", "wp_debug", "true"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.custom", "wp_debug_display", "true"),
				),
			},
		},
	})
}

const testAccResourceWordPressEnvironmentConfig = `
provider "kinsta" {
  # API key and company ID should be set via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

resource "kinsta_wordpress_site" "test" {
  display_name   = "Terraform Test Site for Env"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Test Site"
  wp_language    = "en_US"
  install_mode   = "new"
}

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.test.site_id
  display_name = "Staging Environment"
  is_premium   = false
}
`

const testAccResourceWordPressEnvironmentConfigPremium = `
provider "kinsta" {
  # API key and company ID should be set via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

resource "kinsta_wordpress_site" "test_premium" {
  display_name   = "Terraform Premium Test Site"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Premium Test Site"
  wp_language    = "en_US"
  install_mode   = "new"
}

resource "kinsta_wordpress_environment" "premium_staging" {
  site_id       = kinsta_wordpress_site.test_premium.site_id
  display_name  = "Premium Staging"
  is_premium    = true
  admin_email   = "premium@example.com"
  admin_password = "PremiumP@ss123"
}
`

const testAccResourceWordPressEnvironmentConfigCustom = `
provider "kinsta" {
  # API key and company ID should be set via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

resource "kinsta_wordpress_site" "test_custom" {
  display_name   = "Terraform Custom Test Site"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Custom Test Site"
  wp_language    = "en_US"
  install_mode   = "new"
}

resource "kinsta_wordpress_environment" "custom" {
  site_id           = kinsta_wordpress_site.test_custom.site_id
  display_name      = "Custom Environment"
  is_premium        = false
  php_version       = "8.2"
  wp_debug          = true
  wp_debug_display  = true
  wp_debug_log      = true
}
`

func testAccCheckWordPressEnvironmentExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		// Verify that site_id is also set
		if _, ok := rs.Primary.Attributes["site_id"]; !ok {
			return fmt.Errorf("site_id is not set")
		}

		return nil
	}
}
