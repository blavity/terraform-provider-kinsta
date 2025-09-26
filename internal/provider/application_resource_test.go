package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_ResourceApplication(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApplicationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists("kinsta_application.my_application"),
				),
			},
		},
	})
}

const testAccResourceApplicationConfig = `
provider "kinsta" {
  # Configure your Kinsta provider here
}

resource "kinsta_application" "my_application" {
  name         = "my-application"
  display_name = "My Application"
  region       = "us-central1"
}
`

func testAccCheckApplicationExists(name string) resource.TestCheckFunc {
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
