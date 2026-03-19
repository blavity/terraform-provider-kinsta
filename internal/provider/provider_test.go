package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"kinsta": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}

func testAccPreCheck(t *testing.T) {
	// Pre-check for acceptance tests. Add credential validation here if needed.
}
