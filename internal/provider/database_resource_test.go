package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAcc_ResourceDatabase(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDatabaseConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("kinsta_database.my_database"),
				),
			},
		},
	})
}

func testAccPreCheck(t *testing.T) {
	// You can add any pre-check logic here
}

const testAccResourceDatabaseConfig = `
provider "kinsta" {
  # Configure your Kinsta provider here
}

resource "kinsta_database" "my_database" {
  name         = "my-database"
  display_name = "My Database"
  region       = "us-central1"
  db_type      = "postgresql"
  version      = "15"
  size         = "db1"
}
`

func testAccCheckDatabaseExists(name string) resource.TestCheckFunc {
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
