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

func TestAcc_ResourceWordPressSite_Multisite(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigMultisite,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test_multisite"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_multisite", "is_multisite", "true"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_multisite", "woocommerce", "true"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_multisite", "wordpressseo", "true"),
					resource.TestCheckResourceAttrSet("kinsta_wordpress_site.test_multisite", "site_id"),
					resource.TestCheckResourceAttrSet("kinsta_wordpress_site.test_multisite", "environment_id"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressSite_PlainMode(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigPlain,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test_plain"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_plain", "install_mode", "plain"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressSite_Import(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig,
				Check:  testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
			},
			{
				ResourceName:      "kinsta_wordpress_site.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Write-only fields are not returned by the API and cannot be verified on import
				ImportStateVerifyIgnore: []string{
					"region", "admin_email", "admin_password", "admin_user",
					"site_title", "wp_language", "install_mode",
					"is_multisite", "is_subdomain_multisite", "woocommerce", "wordpressseo",
				},
			},
		},
	})
}

func TestAcc_ResourceWordPressSite_ForceNew(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	var firstSiteID string

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
					resource.TestCheckResourceAttrWith("kinsta_wordpress_site.test", "site_id", func(v string) error {
						firstSiteID = v
						return nil
					}),
				),
			},
			{
				Config: testAccResourceWordPressSiteConfigForceNew,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
					resource.TestCheckResourceAttrWith("kinsta_wordpress_site.test", "site_id", func(v string) error {
						if v == firstSiteID {
							return fmt.Errorf("expected site ID to change after ForceNew field update, got same ID %s", v)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "display_name", "Terraform ForceNew Test"),
				),
			},
		},
	})
}

const testAccResourceWordPressSiteConfigMultisite = `
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_multisite" {
  display_name          = "Terraform Multisite Test"
  region                = "us-central1"
  admin_email           = "test@example.com"
  admin_password        = "SecureP@ssw0rd123"
  admin_user            = "tfadmin"
  site_title            = "Terraform Multisite Test"
  is_multisite          = true
  is_subdomain_multisite = false
  woocommerce           = true
  wordpressseo          = true
}
`

const testAccResourceWordPressSiteConfigPlain = `
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_plain" {
  display_name   = "Terraform Plain Mode Test"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Terraform Plain Mode Test"
  install_mode   = "plain"
}
`

const testAccResourceWordPressSiteConfigForceNew = `
provider "kinsta" {}

resource "kinsta_wordpress_site" "test" {
  display_name   = "Terraform ForceNew Test"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = "Terraform ForceNew Test"
  install_mode   = "new"
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
