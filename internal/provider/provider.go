package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KINSTA_API_KEY", nil),
				Description: "The API key for the Kinsta API.",
			},
			"company_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KINSTA_COMPANY_ID", nil),
				Description: "The ID of your Kinsta company.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"kinsta_database":         resourceDatabase(),
			"kinsta_application":      resourceApplication(),
			"kinsta_wordpress_site": resourceWordPressSite(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	companyID := d.Get("company_id").(string)
	var diags diag.Diagnostics

	if apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API key is missing",
			Detail:   "API key for Kinsta API is missing. Please provide it in the provider configuration or set the KINSTA_API_KEY environment variable.",
		})
	}

	if companyID == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Company ID is missing",
			Detail:   "Company ID for Kinsta API is missing. Please provide it in the provider configuration or set the KINSTA_COMPANY_ID environment variable.",
		})
	}

	if diags.HasError() {
		return nil, diags
	}

	c := client.New(apiKey, companyID)

	return c, diags
}
