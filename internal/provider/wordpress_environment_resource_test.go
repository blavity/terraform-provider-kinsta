package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

func TestAcc_ResourceWordPressEnvironment_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
	}

	siteName := acctest.RandomWithPrefix(testAccNamePrefix)
	envName := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressEnvironmentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressEnvironmentConfig(siteName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressEnvironmentExists("kinsta_wordpress_environment.staging"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.staging", "display_name", envName),
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

	siteName := acctest.RandomWithPrefix(testAccNamePrefix)
	envName := acctest.RandomWithPrefix(testAccNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWordPressEnvironmentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressEnvironmentConfigPremium(siteName, envName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressEnvironmentExists("kinsta_wordpress_environment.premium_staging"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.premium_staging", "is_premium", "true"),
					resource.TestCheckResourceAttr("kinsta_wordpress_environment.premium_staging", "admin_email", "premium@example.com"),
				),
			},
		},
	})
}

func testAccResourceWordPressEnvironmentConfig(siteName, envName string) string {
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

resource "kinsta_wordpress_environment" "staging" {
  site_id      = kinsta_wordpress_site.test.site_id
  display_name = %[2]q
  is_premium   = false
}
`, siteName, envName)
}

func testAccResourceWordPressEnvironmentConfigPremium(siteName, envName string) string {
	return fmt.Sprintf(`
provider "kinsta" {}

resource "kinsta_wordpress_site" "test_premium" {
  display_name   = %[1]q
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "SecureP@ssw0rd123"
  admin_user     = "tfadmin"
  site_title     = %[1]q
  wp_language    = "en_US"
  install_mode   = "new"
}

resource "kinsta_wordpress_environment" "premium_staging" {
  site_id        = kinsta_wordpress_site.test_premium.site_id
  display_name   = %[2]q
  is_premium     = true
  admin_email    = "premium@example.com"
  admin_password = "PremiumP@ss123"
}
`, siteName, envName)
}

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

// testAccCheckWordPressEnvironmentDestroy verifies that every
// kinsta_wordpress_environment in the post-test state is gone. The
// environments list is reachable only via GetWordPressSite on the parent
// site; if the parent site is also gone (NotFound), every environment
// under it is implicitly gone too.
func testAccCheckWordPressEnvironmentDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Helper()
		c := testAccClient(t)
		ctx := context.Background()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "kinsta_wordpress_environment" {
				continue
			}
			envID := rs.Primary.ID
			siteID := rs.Primary.Attributes["site_id"]
			if envID == "" || siteID == "" {
				continue
			}

			site, err := c.GetWordPressSite(ctx, siteID)
			if err != nil {
				if client.IsNotFound(err) {
					// Parent site gone → env gone with it.
					continue
				}
				return fmt.Errorf("unexpected error verifying destroy of env %s: %w", envID, err)
			}
			for _, env := range site.Site.Environments {
				if env.ID == envID {
					return fmt.Errorf("kinsta_wordpress_environment %s still present in site %s after destroy", envID, siteID)
				}
			}
		}
		return nil
	}
}
