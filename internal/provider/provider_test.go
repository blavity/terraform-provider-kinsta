package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

// testAccNamePrefix is the prefix every acceptance-test resource name starts
// with. The sweeper (#65) keys off this prefix to identify orphans, and the
// random suffix tail prevents collisions across parallel runs.
const testAccNamePrefix = "tf-acc-test"

var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"kinsta": func() (*schema.Provider, error) { //nolint:unparam // error always nil; matches required factory signature
			return Provider(), nil
		},
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("KINSTA_API_KEY") == "" {
		t.Fatal("KINSTA_API_KEY must be set for acceptance tests")
	}
	if os.Getenv("KINSTA_COMPANY_ID") == "" {
		t.Fatal("KINSTA_COMPANY_ID must be set for acceptance tests")
	}
}

// testAccClient returns a real Kinsta client built from the env vars
// validated in testAccPreCheck. CheckDestroy helpers use it to verify
// resources are gone from the API after each test.
func testAccClient(t *testing.T) client.KinstaClient {
	t.Helper()
	return client.New(os.Getenv("KINSTA_API_KEY"), os.Getenv("KINSTA_COMPANY_ID"))
}
