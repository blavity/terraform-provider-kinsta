package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

func resourceWordPressSite() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWordPressSiteCreate,
		ReadContext:   resourceWordPressSiteRead,
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
			// Write-only fields: sent on creation but not returned by the Kinsta API.
			// Marked Optional+Computed so import works (state will be empty after import;
			// subsequent plans will show a diff only if the config value differs from "").
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Data center region where the site is hosted (e.g., `us-central1`, `europe-west1`). See the [Kinsta API docs](https://api-docs.kinsta.com) for the full list of supported regions. Write-only: not returned by the API after creation.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"admin_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Email address for the WordPress admin account. Write-only: not returned by the API after creation.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"admin_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Password for the WordPress admin account. Write-only: not returned by the API after creation.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"admin_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Username for the WordPress admin account. Write-only: not returned by the API after creation.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"site_title": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "WordPress site title displayed in the browser tab and site header. Write-only: not returned by the API after creation.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			// Optional fields with defaults (safe after import)
			"install_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "new",
				ForceNew: true,
				// Explicit allowlist surfaces typos and unsupported values at
				// plan time. `clone` is intentionally excluded — the upstream
				// API still accepts it but documents it as deprecated, so we
				// don't expose it. Future upstream additions will require a
				// matching entry here: a deliberate, reviewable change
				// rather than silent passthrough.
				ValidateFunc: validation.StringInSlice([]string{"new", "plain", "migrate"}, false),
				Description: "WordPress installation mode. " +
					"`new` (default) provisions the full WordPress install template — default theme, sample content, and the admin user from `admin_user`/`admin_email`/`admin_password`. " +
					"`plain` creates an empty WordPress container with no install template (matches the \"Empty site\" option in the MyKinsta UI), suitable for sites whose contents are pushed by a downstream pipeline (e.g., Bedrock). " +
					"`migrate` provisions an empty container in preparation for a migration request submitted via the MyKinsta UI; the migration flow itself is out of scope for this provider. " +
					"Write-only credentials are still sent in all modes and apply once content lands.",
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

	installMode := d.Get("install_mode").(string)
	companyID := c.CompanyID()
	displayName := d.Get("display_name").(string)
	region := d.Get("region").(string)

	// Spec note: POST /sites has two body shapes — `addWPSite-Body`
	// (full install) and `addPlainWPSite-Body` (empty container, only
	// company/display_name/region). install_mode = "plain" is selected
	// by *which body we send*, not by passing install_mode = "plain"
	// in the full body. Hybrid bodies risk being rejected by a
	// spec-strict server, so we branch here.
	var (
		resp *client.CreateWordPressSiteResponse
		err  error
	)
	if installMode == "plain" {
		resp, err = c.CreatePlainWordPressSite(ctx, &client.CreatePlainWordPressSiteRequest{
			Company:     companyID,
			DisplayName: displayName,
			Region:      region,
		})
	} else {
		resp, err = c.CreateWordPressSite(ctx, &client.CreateWordPressSiteRequest{
			Company:              companyID,
			DisplayName:          displayName,
			Region:               region,
			InstallMode:          installMode,
			AdminEmail:           d.Get("admin_email").(string),
			AdminPassword:        d.Get("admin_password").(string),
			AdminUser:            d.Get("admin_user").(string),
			SiteTitle:            d.Get("site_title").(string),
			WPLanguage:           d.Get("wp_language").(string),
			IsMultisite:          d.Get("is_multisite").(bool),
			IsSubdomainMultisite: d.Get("is_subdomain_multisite").(bool),
			WooCommerce:          d.Get("woocommerce").(bool),
			WordPressSEO:         d.Get("wordpressseo").(bool),
		})
	}
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

	// Explicitly persist write-only fields — the API does not return these, so they
	// must be saved from the input before Read is called (which would otherwise
	// leave them empty in state for Optional+Computed fields). In `plain`
	// mode the admin/site_title fields aren't sent on the wire, but they may
	// still be present in the configuration — preserve whatever the user
	// declared so subsequent plans don't show a churn.
	for k, v := range map[string]interface{}{
		"region":         region,
		"admin_email":    d.Get("admin_email").(string),
		"admin_password": d.Get("admin_password").(string),
		"admin_user":     d.Get("admin_user").(string),
		"site_title":     d.Get("site_title").(string),
	} {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}

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
