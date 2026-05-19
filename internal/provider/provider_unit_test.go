package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Provider_Schema(t *testing.T) {
	p := Provider()
	require.NotNil(t, p)
	assert.Contains(t, p.Schema, "api_key")
	assert.Contains(t, p.Schema, "company_id")
	assert.True(t, p.Schema["api_key"].Sensitive, "api_key must be marked Sensitive")
	assert.Contains(t, p.ResourcesMap, "kinsta_wordpress_site")
	assert.Contains(t, p.ResourcesMap, "kinsta_wordpress_environment")
}

func Test_providerConfigure_MissingAPIKey(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"api_key":    "",
		"company_id": "test-company-id",
	})
	_, diags := providerConfigure(context.Background(), d)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "API key")
}

func Test_providerConfigure_MissingCompanyID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"api_key":    "test-api-key",
		"company_id": "",
	})
	_, diags := providerConfigure(context.Background(), d)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Company ID")
}

func Test_providerConfigure_Valid(t *testing.T) {
	d := schema.TestResourceDataRaw(t, Provider().Schema, map[string]interface{}{
		"api_key":    "test-api-key",
		"company_id": "test-company-id",
	})
	client, diags := providerConfigure(context.Background(), d)
	assert.False(t, diags.HasError())
	assert.NotNil(t, client)
}

// Principle IV: KINSTA_API_KEY and KINSTA_COMPANY_ID are the only sanctioned
// ambient credential sources. A typo in either env-var name would silently
// break the documented contract.
func Test_Provider_EnvDefaults(t *testing.T) {
	t.Setenv("KINSTA_API_KEY", "env-api-key-sentinel")
	t.Setenv("KINSTA_COMPANY_ID", "env-company-id-sentinel")

	p := Provider()

	apiKeyDefault, err := p.Schema["api_key"].DefaultFunc()
	require.NoError(t, err)
	assert.Equal(t, "env-api-key-sentinel", apiKeyDefault, "api_key must read from KINSTA_API_KEY")

	companyIDDefault, err := p.Schema["company_id"].DefaultFunc()
	require.NoError(t, err)
	assert.Equal(t, "env-company-id-sentinel", companyIDDefault, "company_id must read from KINSTA_COMPANY_ID")
}
