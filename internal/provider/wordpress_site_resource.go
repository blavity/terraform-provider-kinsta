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
			"admin_email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"admin_password": {
				Type:     schema.TypeString,
				Required: true,
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
				Required: true,
			},
		},
	}
}

func resourceWordPressSiteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)

	req := &client.CreateWordPressSiteRequest{
		Company:      c.CompanyID(),
		DisplayName:  d.Get("display_name").(string),
		Region:       d.Get("region").(string),
		AdminEmail:   d.Get("admin_email").(string),
		AdminPassword: d.Get("admin_password").(string),
		AdminUser:    d.Get("admin_user").(string),
		SiteTitle:    d.Get("site_title").(string),
		WPLanguage:   d.Get("wp_language").(string),
	}

	resp, err := c.CreateWordPressSite(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.OperationID)

	return resourceWordPressSiteRead(ctx, d, m)
}

func resourceWordPressSiteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	resp, err := c.GetWordPressSite(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("display_name", resp.Site.DisplayName)
	d.Set("region", resp.Site.Region)

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
