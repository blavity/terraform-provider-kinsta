package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_ResourceWordPressSite_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "display_name", "Terraform Test Site"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "region", "us-central1"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "admin_user", "tfadmin"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "site_title", "Terraform Test Site"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "wp_language", "en_US"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "install_mode", "new"),
					resource.TestCheckResourceAttrSet("kinsta_wordpress_site.test", "site_id"),
					resource.TestCheckResourceAttrSet("kinsta_wordpress_site.test", "environment_id"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressSite_CustomLanguage(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigCustomLanguage,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test_fr"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_fr", "wp_language", "fr_FR"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_fr", "install_mode", "new"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressSite_MigrateMode(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigMigrate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test_migrate"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_migrate", "install_mode", "migrate"),
				),
			},
		},
	})
}

const testAccResourceWordPressSiteConfig = `
provider "kinsta" {
  # API key and company ID should be set via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

resource "kinsta_wordpress_site" "test" {
  display_name   = "Terraform Test Site"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Terraform Test Site"
  wp_language    = "en_US"
  install_mode   = "new"
}
`

const testAccResourceWordPressSiteConfigCustomLanguage = `
provider "kinsta" {
  # API key and company ID should be set via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

resource "kinsta_wordpress_site" "test_fr" {
  display_name   = "Site de test Terraform"
  region         = "europe-west1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Site de test Terraform"
  wp_language    = "fr_FR"
}
`

const testAccResourceWordPressSiteConfigMigrate = `
provider "kinsta" {
  # API key and company ID should be set via environment variables:
  # KINSTA_API_KEY and KINSTA_COMPANY_ID
}

resource "kinsta_wordpress_site" "test_migrate" {
  display_name   = "Terraform Migrate Test"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Terraform Migrate Test"
  install_mode   = "migrate"
}
`

func testAccCheckWordPressSiteExists(name string) resource.TestCheckFunc {
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
