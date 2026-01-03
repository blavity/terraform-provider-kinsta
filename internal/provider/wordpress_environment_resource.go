package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWordPressEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWordPressEnvironmentCreate,
		ReadContext:   resourceWordPressEnvironmentRead,
		DeleteContext: resourceWordPressEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWordPressEnvironmentImport,
		},
		Schema: map[string]*schema.Schema{
			"site_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the WordPress site to create the environment in",
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Display name for the environment (e.g., 'staging', 'premium-staging')",
			},
			"site_title": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "WordPress site title for this environment (write-only, cannot be read back)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress diff if value already exists in state (after create/import)
					return old != ""
				},
			},
			"is_premium": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Whether this is a premium staging environment (true) or standard staging (false) (write-only, cannot be read back)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"admin_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "WordPress admin email (write-only, cannot be read back)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"admin_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "WordPress admin password (write-only, cannot be read back)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"admin_user": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "WordPress admin username (write-only, cannot be read back)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			"wp_language": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "WordPress language (default: en_US) (write-only, cannot be read back)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old != ""
				},
			},
			// Computed output
			"environment_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the created environment",
			},
		},
	}
}

func resourceWordPressEnvironmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)

	// Get wp_language with default
	wpLanguage := d.Get("wp_language").(string)
	if wpLanguage == "" {
		wpLanguage = "en_US"
	}

	req := &client.CreateWordPressEnvironmentRequest{
		DisplayName:   d.Get("display_name").(string),
		SiteTitle:     d.Get("site_title").(string),
		IsPremium:     d.Get("is_premium").(bool),
		AdminEmail:    d.Get("admin_email").(string),
		AdminPassword: d.Get("admin_password").(string),
		AdminUser:     d.Get("admin_user").(string),
		WPLanguage:    wpLanguage,
	}

	siteID := d.Get("site_id").(string)

	// Get existing environment IDs before creation (for deterministic identification)
	beforeResp, err := c.GetWordPressSite(ctx, siteID)
	if err != nil {
		return diag.FromErr(err)
	}
	existingEnvIDs := make(map[string]bool)
	for _, env := range beforeResp.Site.Environments {
		existingEnvIDs[env.ID] = true
	}

	// Create environment (async operation)
	// Note: Kinsta has eventual consistency after environment deletion - the display_name
	// may still be reserved for up to ~30 seconds. Retry with exponential backoff.
	var resp *client.CreateWordPressEnvironmentResponse
	maxRetries := 6 // 2^6 = 64 seconds max wait (1+2+4+8+16+32)
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = c.CreateWordPressEnvironment(ctx, siteID, req)
		if err == nil {
			break
		}

		// Check if error is due to display_name conflict (eventual consistency)
		if attempt < maxRetries && strings.Contains(err.Error(), "display name") && strings.Contains(err.Error(), "already used") {
			waitTime := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-ctx.Done():
				return diag.FromErr(ctx.Err())
			case <-time.After(waitTime):
				continue
			}
		}

		// Not a retryable error, or max retries exceeded
		return diag.FromErr(err)
	}

	// Poll operation until complete
	// Note: Environment creation operations don't return idEnv in response data
	_, err = c.PollOperation(ctx, resp.OperationID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the site's environments after creation
	afterResp, err := c.GetWordPressSite(ctx, siteID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find the new environment by comparing before/after lists
	var envID string
	for _, env := range afterResp.Site.Environments {
		if !existingEnvIDs[env.ID] {
			envID = env.ID
			break
		}
	}

	if envID == "" {
		return diag.Errorf("failed to identify newly created environment (no new environment ID found)")
	}

	// Set the environment_id as the Terraform resource ID
	d.SetId(envID)

	// Preserve write-only fields in state (can't be read back from API)
	d.Set("site_title", req.SiteTitle)
	d.Set("is_premium", req.IsPremium)
	d.Set("admin_email", req.AdminEmail)
	d.Set("admin_password", req.AdminPassword)
	d.Set("admin_user", req.AdminUser)
	d.Set("wp_language", wpLanguage)

	// Read the environment to populate computed attributes
	return resourceWordPressEnvironmentRead(ctx, d, m)
}

func resourceWordPressEnvironmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	envID := d.Id()
	siteID := d.Get("site_id").(string)

	// Get the site to access its environments list
	// Individual environment GET endpoint returns 404, must use site's environments list
	siteResp, err := c.GetWordPressSite(ctx, siteID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Find the environment in the site's environments list
	var foundEnv *client.WordPressEnvironment

	for i := range siteResp.Site.Environments {
		if siteResp.Site.Environments[i].ID == envID {
			foundEnv = &siteResp.Site.Environments[i]
			break
		}
	}

	if foundEnv == nil {
		// Environment not found, mark as deleted
		d.SetId("")
		return nil
	}

	d.Set("environment_id", foundEnv.ID)
	d.Set("display_name", foundEnv.DisplayName)

	return nil
}

func resourceWordPressEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Environments don't support updates - all fields are ForceNew
	var diags diag.Diagnostics
	return diags
}

func resourceWordPressEnvironmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.KinstaClient)
	id := d.Id()

	_, err := c.DeleteWordPressEnvironment(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceWordPressEnvironmentImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Import ID format: site_id:environment_id
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid import ID format, expected 'site_id:environment_id', got: %s", d.Id())
	}

	siteID := parts[0]
	envID := parts[1]

	// Set the site_id in state
	d.Set("site_id", siteID)
	// Set the environment_id as the resource ID
	d.SetId(envID)

	// Read to populate remaining attributes
	diags := resourceWordPressEnvironmentRead(ctx, d, m)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to read environment during import: %v", diags)
	}

	return []*schema.ResourceData{d}, nil
}
