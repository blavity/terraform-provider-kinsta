package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

func resourceWordPressSite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWordPressSiteCreate,
		ReadContext:   resourceWordPressSiteRead,
		UpdateContext: resourceWordPressSiteUpdate,
		DeleteContext: resourceWordPressSiteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A human-readable name for the site shown in MyKinsta (e.g., `my-production-site`).",
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Data center region where the site is hosted (e.g., `us-central1`, `europe-west1`). See the [Kinsta API docs](https://api-docs.kinsta.com) for the full list of supported regions.",
			},
			"install_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "new",
				ForceNew:    true,
				Description: "WordPress installation mode. Currently only `new` is supported.",
			},
			"admin_email": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Email address for the WordPress admin account.",
			},
			"admin_password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Password for the WordPress admin account.",
			},
			"admin_user": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Username for the WordPress admin account.",
			},
			"site_title": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "WordPress site title displayed in the browser tab and site header.",
			},
			"wp_language": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "en_US",
				ForceNew:    true,
				Description: "WordPress locale code (e.g., `en_US`, `fr_FR`, `de_DE`). Defaults to `en_US`.",
			},
			// Optional WordPress configuration (write-only, ForceNew)
			"is_multisite": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Enable WordPress Multisite. Cannot be read back from the API after creation.",
			},
			"is_subdomain_multisite": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Use subdomain-based multisite instead of subdirectory multisite. Only applies when `is_multisite` is `true`. Cannot be read back from the API after creation.",
			},
			"woocommerce": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Pre-install the WooCommerce plugin. Cannot be read back from the API after creation.",
			},
			"wordpressseo": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Pre-install the Yoast SEO plugin. Cannot be read back from the API after creation.",
			},
			// Computed outputs
			"site_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier for the WordPress site, used to reference this site in other resources.",
			},
			"environment_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier for the live environment automatically created with the site.",
			},
		},
	}
}

func resourceWordPressSiteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)

	req := &client.CreateWordPressSiteRequest{
		Company:              c.CompanyID(),
		DisplayName:          d.Get("display_name").(string),
		Region:               d.Get("region").(string),
		InstallMode:          d.Get("install_mode").(string),
		AdminEmail:           d.Get("admin_email").(string),
		AdminPassword:        d.Get("admin_password").(string),
		AdminUser:            d.Get("admin_user").(string),
		SiteTitle:            d.Get("site_title").(string),
		WPLanguage:           d.Get("wp_language").(string),
		IsMultisite:          d.Get("is_multisite").(bool),
		IsSubdomainMultisite: d.Get("is_subdomain_multisite").(bool),
		WooCommerce:          d.Get("woocommerce").(bool),
		WordPressSEO:         d.Get("wordpressseo").(bool),
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
		if client.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("site_id", resp.Site.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", resp.Site.DisplayName); err != nil {
		return diag.FromErr(err)
	}

	// Extract environment_id for the live environment (site creation auto-creates live)
	for _, env := range resp.Site.Environments {
		if env.Name == "live" {
			if err := d.Set("environment_id", env.ID); err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	return nil
}

func resourceWordPressSiteUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Updates are not supported; all fields are ForceNew
	return diag.Errorf("kinsta_wordpress_site does not support updates; all fields are immutable and require resource replacement")
}

func resourceWordPressSiteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	resp, err := c.DeleteWordPressSite(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := c.PollOperation(ctx, resp.OperationID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
