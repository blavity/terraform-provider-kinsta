package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kinsta/terraform-provider-kinsta/internal/client"
)

func resourceWordPressSite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWordPressSiteCreate,
		ReadContext:   resourceWordPressSiteRead,
		UpdateContext: resourceWordPressSiteUpdate,
		DeleteContext: resourceWordPressSiteDelete,
		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"install_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "new",
			},
			"admin_email": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"admin_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"admin_user": {
				Type:     schema.TypeString,
				Required: true,
			},
			"site_title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"wp_language": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "en_US",
			},
			// Computed outputs
			"site_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWordPressSiteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)

	req := &client.CreateWordPressSiteRequest{
		Company:       c.CompanyID(),
		DisplayName:   d.Get("display_name").(string),
		Region:        d.Get("region").(string),
		InstallMode:   d.Get("install_mode").(string),
		AdminEmail:    d.Get("admin_email").(string),
		AdminPassword: d.Get("admin_password").(string),
		AdminUser:     d.Get("admin_user").(string),
		SiteTitle:     d.Get("site_title").(string),
		WPLanguage:    d.Get("wp_language").(string),
	}

	// Create site (async operation)
	resp, err := c.CreateWordPressSite(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	// Poll operation until complete
	siteID, err := c.PollOperation(ctx, resp.OperationID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the site_id as the Terraform resource ID
	d.SetId(siteID)

	// Read the site to populate computed attributes
	return resourceWordPressSiteRead(ctx, d, m)
}

func resourceWordPressSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	siteID := d.Id()

	resp, err := c.GetWordPressSite(ctx, siteID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("site_id", resp.Site.ID)
	d.Set("display_name", resp.Site.DisplayName)

	// Extract environment_id from the first environment (usually "live")
	if len(resp.Site.Environments) > 0 {
		d.Set("environment_id", resp.Site.Environments[0].ID)
	}

	return nil
}

func resourceWordPressSiteUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceWordPressSiteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	_, err := c.DeleteWordPressSite(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
