package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

func TestAcc_ResourceWordPressSite_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	name := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "display_name", name),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "region", "us-central1"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "admin_user", "tfadmin"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "site_title", name),
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

	name := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigCustomLanguage(name),
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

	name := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigMigrate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test_migrate"),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test_migrate", "install_mode", "migrate"),
				),
			},
		},
	})
}

func TestAcc_ResourceWordPressSite_Multisite(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	name := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigMultisite(name),
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

	name := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfigPlain(name),
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

	name := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig(name),
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

	firstName := acctest.RandomWithPrefix(testAccNamePrefix)
	secondName := acctest.RandomWithPrefix(testAccNamePrefix)

	var firstSiteID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressSiteDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig(firstName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
					resource.TestCheckResourceAttrWith("kinsta_wordpress_site.test", "site_id", func(v string) error {
						firstSiteID = v
						return nil
					}),
				),
			},
			{
				Config: testAccResourceWordPressSiteConfig(secondName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.test"),
					resource.TestCheckResourceAttrWith("kinsta_wordpress_site.test", "site_id", func(v string) error {
						if v == firstSiteID {
							return fmt.Errorf("expected site ID to change after ForceNew field update, got same ID %s", v)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("kinsta_wordpress_site.test", "display_name", secondName),
				),
			},
		},
	})
}

// Config builders. Each takes a randomized name (prefixed `tf-acc-test`) so
// parallel runs and incomplete cleanups can't collide on display_name —
// MyKinsta rejects duplicates within a company.
func testAccResourceWordPressSiteConfig(name string) string {
	return fmt.Sprintf(`
provider "kinsta" {}

resource "kinsta_wordpress_site" "test" {
  display_name   = %[1]q
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = %[1]q
  wp_language    = "en_US"
  install_mode   = "new"
}
`, name)
}

func testAccResourceWordPressSiteConfigCustomLanguage(name string) string {
	return fmt.Sprintf(`
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_fr" {
  display_name   = %[1]q
  region         = "europe-west1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = %[1]q
  wp_language    = "fr_FR"
}
`, name)
}

func testAccResourceWordPressSiteConfigMigrate(name string) string {
	return fmt.Sprintf(`
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_migrate" {
  display_name   = %[1]q
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = %[1]q
  install_mode   = "migrate"
}
`, name)
}

func testAccResourceWordPressSiteConfigMultisite(name string) string {
	return fmt.Sprintf(`
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_multisite" {
  display_name           = %[1]q
  region                 = "us-central1"
  admin_email            = "test@example.com"
  admin_password         = "SecureP@ssw0rd123"
  admin_user             = "tfadmin"
  site_title             = %[1]q
  is_multisite           = true
  is_subdomain_multisite = false
  woocommerce            = true
  wordpressseo           = true
}
`, name)
}

func testAccResourceWordPressSiteConfigPlain(name string) string {
	return fmt.Sprintf(`
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_plain" {
  display_name   = %[1]q
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = %[1]q
  install_mode   = "plain"
}
`, name)
}

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

// testAccCheckWordPressSiteDestroy verifies that every kinsta_wordpress_site
// in the post-test state has actually been removed from the MyKinsta API.
// Hits the live API with the credentials from KINSTA_API_KEY /
// KINSTA_COMPANY_ID (validated by testAccPreCheck).
func testAccCheckWordPressSiteDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Helper()
		c := testAccClient(t)
		ctx := context.Background()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "kinsta_wordpress_site" {
				continue
			}
			id := rs.Primary.ID
			if id == "" {
				continue
			}

			_, err := c.GetWordPressSite(ctx, id)
			if err == nil {
				return fmt.Errorf("kinsta_wordpress_site %s still exists after destroy", id)
			}
			if !client.IsNotFound(err) {
				return fmt.Errorf("unexpected error verifying destroy of site %s: %w", id, err)
			}
		}
		return nil
	}
}
