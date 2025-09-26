package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_ResourceWordPressSite(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWordPressSiteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWordPressSiteExists("kinsta_wordpress_site.my_site"),
				),
			},
		},
	})
}

const testAccResourceWordPressSiteConfig = `
provider "kinsta" {
  # Configure your Kinsta provider here
}

resource "kinsta_wordpress_site" "my_site" {
  display_name   = "My Site"
  region         = "us-central1"
  admin_email    = "test@example.com"
  admin_password = "password"
  admin_user     = "admin"
  site_title     = "My Site"
  wp_language    = "en_US"
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

		return nil
	}
}
